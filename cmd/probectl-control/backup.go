// SPDX-License-Identifier: LicenseRef-probectl-TBD

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/imfeelingtheagi/probectl/internal/backup"
	"github.com/imfeelingtheagi/probectl/internal/crypto"
	"github.com/imfeelingtheagi/probectl/internal/tenantcrypto"
)

// backup-seal / backup-open (OPS-002): stdin→stdout envelope-encryption
// filters so a backup CronJob can pipe a dump straight through encryption —
// `pg_dump ... | probectl-control backup-seal --key-file=K > out.pbk` — and
// nothing ever lands on disk in plaintext. Restore: `probectl-control
// backup-open --key-file=K < out.pbk | pg_restore ...`.
//
// The key is the SAME deployment KEK as the at-rest envelope (Sprint 8): a
// base64 env value (PROBECTL_ENVELOPE_KEY) or a key file
// (PROBECTL_ENVELOPE_KEY_FILE / --key-file), so a restore on a fresh node
// only needs that one key.

func backupSeal(args []string) error { return runBackup(args, true) }
func backupOpen(args []string) error { return runBackup(args, false) }

func runBackup(args []string, seal bool) error {
	name := "backup-open"
	if seal {
		name = "backup-seal"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	keyFile := fs.String("key-file", os.Getenv("PROBECTL_ENVELOPE_KEY_FILE"), "path to the base64 KEK file (or set PROBECTL_ENVELOPE_KEY)")
	defKeyID := os.Getenv("PROBECTL_ENVELOPE_KEY_ID")
	if defKeyID == "" {
		defKeyID = "file"
	}
	keyID := fs.String("key-id", defKeyID, "KEK id stamped into / matched against the container header")
	if err := fs.Parse(args); err != nil {
		return err
	}

	keys, err := backupKeyProvider(*keyFile, *keyID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if seal {
		if err := backup.Seal(ctx, os.Stdout, os.Stdin, keys); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}
	if err := backup.Open(ctx, os.Stdout, os.Stdin, keys); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// backupKeyProvider builds the KEK key provider from the env key or a key
// file (the Sprint 8 sources), failing closed if neither is present — a
// keyless backup would defeat OPS-002.
func backupKeyProvider(keyFile, keyID string) (crypto.KeyProvider, error) {
	if b64 := os.Getenv("PROBECTL_ENVELOPE_KEY"); b64 != "" {
		return crypto.NewStaticKeyProviderFromBase64(keyID, b64)
	}
	if keyFile != "" {
		// LoadOrGenerate would MINT a key on a restore node that lacks it,
		// silently producing an unreadable mismatch — for backups we require
		// the existing key, so read it without generating.
		b64, generated, err := tenantcrypto.LoadOrGenerateKeyFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("backup key file: %w", err)
		}
		if generated {
			return nil, fmt.Errorf("backup key file %q did not exist — refusing to MINT a key for a backup (a sealed backup needs its ORIGINAL KEK; provide the existing key)", keyFile)
		}
		return crypto.NewStaticKeyProviderFromBase64(keyID, b64)
	}
	return nil, fmt.Errorf("no envelope key: set PROBECTL_ENVELOPE_KEY or --key-file (OPS-002 — backups are never written unencrypted)")
}
