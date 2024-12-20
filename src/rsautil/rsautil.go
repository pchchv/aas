package rsautil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) (privatePEM []byte) {
	// ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privatePEM = pem.EncodeToMemory(&privBlock)

	return
}

func EncodePublicKeyToPEM(publicKey *rsa.PublicKey) (pubkey_pem []byte, err error) {
	if pubkey_bytes, err := x509.MarshalPKIXPublicKey(publicKey); err != nil {
		return nil, err
	} else {
		pubkey_pem = pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: pubkey_bytes,
			},
		)
	}

	return
}
