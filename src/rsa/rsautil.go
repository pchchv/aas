package rsautil

import (
	"crypto/rand"
	"crypto/rsa"
)

func GeneratePrivateKey(bitSize int) (privateKey *rsa.PrivateKey, err error) {
	if privateKey, err = rsa.GenerateKey(rand.Reader, bitSize); err != nil {
		return nil, err
	}

	if err = privateKey.Validate(); err != nil {
		return nil, err
	}

	return privateKey, nil
}
