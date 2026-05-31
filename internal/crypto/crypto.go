package crypto

import "crypto/sha256"

// Hash returns the SHA-256 digest of data. This is the only import of a crypto/*
// primitive in netctl; S3 routes it through the swappable provider interface so a
// FIPS module can supply the implementation.
func Hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}
