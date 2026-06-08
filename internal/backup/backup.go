// Package backup is probectl's at-rest backup encryption (OPS-002): a
// streaming envelope-encrypted container so a pg_dump / ClickHouse BACKUP
// never lands on disk in plaintext. It reuses the Sprint 8 at-rest key
// management (internal/crypto envelope: a fresh DEK per backup, wrapped by
// the deployment KEK) — crypto goes through internal/crypto only (guardrail
// 3).
//
// Format (.pbk container), chunked so an arbitrarily large dump streams
// without buffering in memory:
//
//	magic "PBK1" || uint16(len keyID) || keyID
//	     || uint32(len wrappedDEK) || wrappedDEK
//	     || repeated: uint32(len chunkCiphertext) || chunkCiphertext
//	     || uint32(0)   # zero-length terminator = clean EOF
//
// Each chunk is sealed with AES-256-GCM under the DEK, with the chunk INDEX
// as additional data — so a truncated or reordered file fails to open
// (tamper/truncation evident). A missing terminator (the writer died
// mid-stream) is reported as a corrupt backup rather than a silent short
// read.
package backup

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/imfeelingtheagi/probectl/internal/crypto"
)

const (
	magic     = "PBK1"
	chunkSize = 1 << 20 // 1 MiB plaintext per sealed chunk
)

// KeyProvider wraps/unwraps the data key (the deployment KEK side). The
// crypto.KeyProvider from internal/crypto satisfies it; the CLI builds one
// from the envelope key.
type KeyProvider = crypto.KeyProvider

// Seal streams src → dst, envelope-encrypted. A single fresh DEK is minted
// for the whole backup and wrapped by keys; each chunk is independently
// sealed. dst receives the .pbk container.
func Seal(ctx context.Context, dst io.Writer, src io.Reader, keys KeyProvider) error {
	// One fresh DEK for the whole backup, wrapped by the KEK in the header;
	// each chunk is then sealed under that DEK with its index as AAD.
	hdr, dek, err := sealHeader(ctx, keys)
	if err != nil {
		return err
	}
	if _, err := dst.Write(hdr); err != nil {
		return fmt.Errorf("backup: write header: %w", err)
	}

	buf := make([]byte, chunkSize)
	var index uint64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, rerr := io.ReadFull(src, buf)
		if n > 0 {
			ct, err := crypto.Encrypt(dek, buf[:n], chunkAAD(index))
			if err != nil {
				return fmt.Errorf("backup: seal chunk %d: %w", index, err)
			}
			if err := writeFrame(dst, ct); err != nil {
				return err
			}
			index++
		}
		if rerr == io.EOF || rerr == io.ErrUnexpectedEOF {
			break
		}
		if rerr != nil {
			return fmt.Errorf("backup: read input: %w", rerr)
		}
	}
	// Zero-length terminator = clean EOF.
	if err := binary.Write(dst, binary.BigEndian, uint32(0)); err != nil {
		return fmt.Errorf("backup: write terminator: %w", err)
	}
	return nil
}

// Open streams a .pbk container src → dst, decrypted. A truncated file (no
// terminator) or a tampered chunk fails — backups are verified, not trusted.
func Open(ctx context.Context, dst io.Writer, src io.Reader, keys KeyProvider) error {
	dek, err := openHeader(ctx, src, keys)
	if err != nil {
		return err
	}
	var index uint64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var n uint32
		if err := binary.Read(src, binary.BigEndian, &n); err != nil {
			if errors.Is(err, io.EOF) {
				return errors.New("backup: truncated container (missing terminator) — incomplete or corrupt backup")
			}
			return fmt.Errorf("backup: read frame length: %w", err)
		}
		if n == 0 {
			return nil // clean EOF
		}
		ct := make([]byte, n)
		if _, err := io.ReadFull(src, ct); err != nil {
			return fmt.Errorf("backup: short read on chunk %d (corrupt): %w", index, err)
		}
		pt, err := crypto.Decrypt(dek, ct, chunkAAD(index))
		if err != nil {
			return fmt.Errorf("backup: chunk %d failed authentication (tampered/reordered): %w", index, err)
		}
		if _, err := dst.Write(pt); err != nil {
			return fmt.Errorf("backup: write output: %w", err)
		}
		index++
	}
}

// --- header (carries the wrapped DEK) ---

// sealHeader mints a DEK, wraps it under the KEK, and renders the container
// header. The raw DEK is returned for chunk AEAD.
func sealHeader(ctx context.Context, keys KeyProvider) (hdr, dek []byte, err error) {
	dek, err = crypto.Random(crypto.KeySize)
	if err != nil {
		return nil, nil, err
	}
	wrapped, err := keys.WrapKey(ctx, dek)
	if err != nil {
		return nil, nil, fmt.Errorf("backup: wrap dek: %w", err)
	}
	keyID := keys.KeyID()
	var b []byte
	b = append(b, magic...)
	b = appendU16(b, uint16(len(keyID)))
	b = append(b, keyID...)
	b = appendU32(b, uint32(len(wrapped)))
	b = append(b, wrapped...)
	return b, dek, nil
}

// openHeader parses the header and unwraps the DEK.
func openHeader(ctx context.Context, src io.Reader, keys KeyProvider) ([]byte, error) {
	m := make([]byte, len(magic))
	if _, err := io.ReadFull(src, m); err != nil {
		return nil, fmt.Errorf("backup: read magic: %w", err)
	}
	if string(m) != magic {
		return nil, fmt.Errorf("backup: bad magic %q (not a probectl backup container)", m)
	}
	keyIDLen, err := readU16(src)
	if err != nil {
		return nil, err
	}
	keyID := make([]byte, keyIDLen)
	if _, err := io.ReadFull(src, keyID); err != nil {
		return nil, fmt.Errorf("backup: read key id: %w", err)
	}
	wrappedLen, err := readU32(src)
	if err != nil {
		return nil, err
	}
	wrapped := make([]byte, wrappedLen)
	if _, err := io.ReadFull(src, wrapped); err != nil {
		return nil, fmt.Errorf("backup: read wrapped dek: %w", err)
	}
	dek, err := keys.UnwrapKey(ctx, wrapped)
	if err != nil {
		return nil, fmt.Errorf("backup: unwrap dek (wrong KEK for key id %q?): %w", keyID, err)
	}
	return dek, nil
}

func chunkAAD(index uint64) []byte {
	aad := make([]byte, 8)
	binary.BigEndian.PutUint64(aad, index)
	return aad
}

func writeFrame(w io.Writer, payload []byte) error {
	if err := binary.Write(w, binary.BigEndian, uint32(len(payload))); err != nil {
		return fmt.Errorf("backup: write frame length: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("backup: write frame: %w", err)
	}
	return nil
}

func appendU16(b []byte, v uint16) []byte { return append(b, byte(v>>8), byte(v)) }
func appendU32(b []byte, v uint32) []byte {
	return append(b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
func readU16(r io.Reader) (uint16, error) {
	var v uint16
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}
func readU32(r io.Reader) (uint32, error) {
	var v uint32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}
