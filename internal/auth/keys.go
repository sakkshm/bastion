package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(n int) (string, error) {
	result := make([]byte, n)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func GenerateAPIKey() (publicID, secret, fullKey string, err error) {
	publicID, err = generateRandomString(6)
	if err != nil {
		return
	}

	secret, err = generateRandomString(32)
	if err != nil {
		return
	}

	fullKey = "bastion_" + publicID + "_" + secret
	return
}

func SplitAPIKey(key string) (string, string, error) {
	parts := strings.Split(key, "_")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("Invalid API key")
	}

	publicID := parts[1]
	secret := parts[2]

	return publicID, secret, nil
}

func HashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}
