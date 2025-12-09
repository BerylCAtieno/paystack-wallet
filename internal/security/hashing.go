package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func GenerateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("sk_live_%s", hex.EncodeToString(bytes))
}

func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
