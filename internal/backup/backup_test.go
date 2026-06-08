// SPDX-License-Identifier: LicenseRef-probectl-TBD

package backup

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/imfeelingtheagi/probectl/internal/crypto"
)

func testKeys(t *testing.T) KeyProvider {
	t.Helper()
	kek, err := crypto.Random(crypto.KeySize)
	if err != nil {
		t.Fatal(err)
	}
	kp, err := crypto.NewStaticKeyProvider("backup-test", kek)
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

// OPS-002: a backup round-trips through the encrypted container, and the
// sealed bytes NEVER contain the plaintext — tenant telemetry never lands on
// disk in the clear. Restore from the encrypted backup is exact.
func TestSealOpenRoundTripAndNoPlaintext(t *testing.T) {
	keys := testKeys(t)
	ctx := context.Background()

	// A multi-chunk "dump" with a recognizable secret marker.
	secret := "TENANT-SECRET-acct-4111111111111111"
	var plain bytes.Buffer
	plain.WriteString("PGDMP fake header\n")
	for i := 0; i < 3000; i++ { // > chunkSize to exercise chunking
		plain.WriteString("row data " + secret + " more data\n")
	}
	original := plain.Bytes()

	var sealed bytes.Buffer
	if err := Seal(ctx, &sealed, bytes.NewReader(original), keys); err != nil {
		t.Fatalf("seal: %v", err)
	}

	// The on-disk artifact must NOT contain the plaintext or its secret.
	if bytes.Contains(sealed.Bytes(), []byte(secret)) {
		t.Fatal("OPS-002 VIOLATION: plaintext secret present in the sealed backup")
	}
	if bytes.Contains(sealed.Bytes(), []byte("PGDMP fake header")) {
		t.Fatal("OPS-002 VIOLATION: plaintext header present in the sealed backup")
	}
	if sealed.Len() <= len(magic) {
		t.Fatal("sealed output suspiciously small")
	}

	// Restore is byte-exact.
	var restored bytes.Buffer
	if err := Open(ctx, &restored, bytes.NewReader(sealed.Bytes()), keys); err != nil {
		t.Fatalf("open: %v", err)
	}
	if !bytes.Equal(restored.Bytes(), original) {
		t.Fatalf("restore mismatch: %d vs %d bytes", restored.Len(), len(original))
	}
}

// The empty backup still round-trips (clean header + terminator only).
func TestSealOpenEmpty(t *testing.T) {
	keys := testKeys(t)
	ctx := context.Background()
	var sealed bytes.Buffer
	if err := Seal(ctx, &sealed, bytes.NewReader(nil), keys); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Open(ctx, &out, bytes.NewReader(sealed.Bytes()), keys); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("empty backup restored %d bytes", out.Len())
	}
}

// A wrong KEK cannot open the backup (the wrapped DEK fails to unwrap).
func TestOpenWrongKeyFails(t *testing.T) {
	ctx := context.Background()
	var sealed bytes.Buffer
	if err := Seal(ctx, &sealed, strings.NewReader("hello backup"), testKeys(t)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Open(ctx, &out, bytes.NewReader(sealed.Bytes()), testKeys(t)); err == nil {
		t.Fatal("a different KEK must not open the backup")
	}
}

// Tamper + truncation are detected — backups are verified, not trusted.
func TestTamperAndTruncationDetected(t *testing.T) {
	keys := testKeys(t)
	ctx := context.Background()
	var sealed bytes.Buffer
	if err := Seal(ctx, &sealed, strings.NewReader(strings.Repeat("x", 5000)), keys); err != nil {
		t.Fatal(err)
	}
	good := sealed.Bytes()

	// Flip a byte deep in the ciphertext body → chunk auth fails.
	tampered := append([]byte(nil), good...)
	tampered[len(tampered)-10] ^= 0xff
	if err := Open(ctx, &bytes.Buffer{}, bytes.NewReader(tampered), keys); err == nil {
		t.Fatal("tampered chunk must fail authentication")
	}

	// Drop the terminator (writer died mid-stream) → truncation reported.
	truncated := good[:len(good)-4]
	err := Open(ctx, &bytes.Buffer{}, bytes.NewReader(truncated), keys)
	if err == nil || !strings.Contains(err.Error(), "truncated") {
		t.Fatalf("truncated container must be reported: %v", err)
	}
}
