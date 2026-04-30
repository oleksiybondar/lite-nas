package authtoken

import (
	"crypto/rand"
	"encoding/base64"
)

const tokenIDBytes = 16

func newTokenID() (string, error) {
	data := make([]byte, tokenIDBytes)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}
