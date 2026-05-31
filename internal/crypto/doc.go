// Package crypto is netctl's crypto abstraction: the single place where
// cryptographic primitives are used, so a FIPS 140-3 validated module can be
// compiled in later (CLAUDE.md §7 guardrail 3). Handlers and services must never
// call crypto/* directly — they route through here.
//
// S2 seeds only Hash, used by the tamper-evident audit chain (internal/audit).
// S3 expands this into the full provider interface (hash, symmetric
// encrypt/decrypt, sign/verify, random), envelope encryption, and mTLS/SPIFFE
// identity.
package crypto
