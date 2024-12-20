package rsautil

import (
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodePrivateKeyToPEM(t *testing.T) {
	privateKey, err := GeneratePrivateKey(2048)
	assert.NoError(t, err)

	privatePEM := EncodePrivateKeyToPEM(privateKey)
	assert.NotNil(t, privatePEM)

	block, _ := pem.Decode(privatePEM)
	assert.NotNil(t, block)
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)
}

func TestEncodePublicKeyToPEM(t *testing.T) {
	privateKey, err := GeneratePrivateKey(2048)
	assert.NoError(t, err)

	publicKey := &privateKey.PublicKey
	publicPEM, err := EncodePublicKeyToPEM(publicKey)
	assert.NoError(t, err)
	assert.NotNil(t, publicPEM)

	block, _ := pem.Decode(publicPEM)
	assert.NotNil(t, block)
	assert.Equal(t, "RSA PUBLIC KEY", block.Type)
}
