package sessions

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const refreshTokenBytes = 32

func newRefreshToken() (string, error) {
	data := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
