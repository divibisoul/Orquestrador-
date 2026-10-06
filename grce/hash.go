package grce

import (
	"crypto/sha256"
	"encoding/hex"
)

func sha256Bytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
