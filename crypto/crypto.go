package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomHexKey generates a cryptographically secure random hex string of 64 characters (32 bytes).
func GenerateRandomHexKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
