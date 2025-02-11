package rsautil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"

	"github.com/pkg/errors"
)

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

func MarshalRSAPublicKeyToJWK(publicKey *rsa.PublicKey, kid string) (publicKeyJWK []byte, err error) {
	jwt := struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		Use string `json:"use"`
		N   string `json:"n"`
		E   string `json:"e"`
	}{
		Alg: "RS256",
		Kid: kid,
		Kty: "RSA",
		Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
	}

	if publicKeyJWK, err = json.MarshalIndent(jwt, "", "  "); err != nil {
		return nil, errors.Wrap(err, "unable to marshal public key to JSON")
	}

	return
}

func GeneratePrivateKey(bitSize int) (privateKey *rsa.PrivateKey, err error) {
	if privateKey, err = rsa.GenerateKey(rand.Reader, bitSize); err != nil {
		return nil, err
	}

	if err = privateKey.Validate(); err != nil {
		return nil, err
	}

	return
}
