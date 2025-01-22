package util

import (
	"crypto"
	"crypto/ed25519"

	"github.com/golang-jwt/jwt/v5"
)

func LoadSigningKey(publicKeyPem string) (crypto.PublicKey, error) {
	pub, err := jwt.ParseEdPublicKeyFromPEM([]byte(publicKeyPem))
	if err != nil {
		return nil, err
	}

	return pub.(ed25519.PublicKey), nil
}
