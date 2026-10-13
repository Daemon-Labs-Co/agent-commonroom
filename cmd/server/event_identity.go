package main

// Cryptographic identifiers and audit chain hashing.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func id() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func checksum(previous, kind, payload string) string {
	h := sha256.Sum256([]byte(previous + "\n" + kind + "\n" + payload))
	return hex.EncodeToString(h[:])
}
