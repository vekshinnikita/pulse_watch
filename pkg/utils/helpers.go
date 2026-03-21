package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashStringWithSalt(value, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(salt + value))
	return hex.EncodeToString(hasher.Sum(nil))
}
