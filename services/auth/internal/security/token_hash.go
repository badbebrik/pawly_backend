package security

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashTokenSHA256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
