package encryption

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptText(t *testing.T) {
	encryptionKey := []byte("thisis32bitlongpassphraseimusing!")
	text := "Hello, World!"

	encryptedText, err := EncryptText(text, encryptionKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedText)

	decryptedText, err := DecryptText(encryptedText, encryptionKey)
	assert.NoError(t, err)
	assert.Equal(t, text, decryptedText)
}

func TestEncryptText_EmptyText(t *testing.T) {
	encryptionKey := []byte("thisis32bitlongpassphraseimusing!")

	encryptedText, err := EncryptText("", encryptionKey)
	assert.Error(t, err)
	assert.Nil(t, encryptedText)
}

func TestEncryptText_InvalidKeyLength(t *testing.T) {
	encryptionKey := []byte("shortkey")
	text := "Hello, World!"

	encryptedText, err := EncryptText(text, encryptionKey)
	assert.Error(t, err)
	assert.Nil(t, encryptedText)
}

func TestDecryptText_EmptyText(t *testing.T) {
	encryptionKey := []byte("thisis32bitlongpassphraseimusing!")

	decryptedText, err := DecryptText(nil, encryptionKey)
	assert.Error(t, err)
	assert.Empty(t, decryptedText)
}

func TestDecryptText_InvalidKeyLength(t *testing.T) {
	encryptionKey := []byte("shortkey")
	encryptedText := []byte("someencryptedtext")

	decryptedText, err := DecryptText(encryptedText, encryptionKey)
	assert.Error(t, err)
	assert.Empty(t, decryptedText)
}

func TestAesGcmEncryption(t *testing.T) {
	clientSecret := "thisis32bitlongpassphraseimusing!"
	idTokenUnencrypted := "idToken123"

	encryptedText, err := AesGcmEncryption(idTokenUnencrypted, clientSecret)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedText)

	decodedEncryptedText, err := base64.StdEncoding.DecodeString(encryptedText)
	assert.NoError(t, err)

	decryptedText, err := DecryptText(decodedEncryptedText, []byte(clientSecret))
	assert.NoError(t, err)
	assert.Equal(t, idTokenUnencrypted, decryptedText)
}
