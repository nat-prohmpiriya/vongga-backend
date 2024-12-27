package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random ID with the specified length in bytes
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
