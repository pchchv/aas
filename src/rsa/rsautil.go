package rsautil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	return pem.EncodeToMemory(&privBlock)
}

func EncodePublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return pubkey_pem, nil
}
