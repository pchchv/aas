package communication

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/encryption"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pchchv/aas/pkg/src/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSendEmail(t *testing.T) {
	emailSender := NewEmailSender()
	aesEncryptionKey := []byte("aes_encryption_key_0000000000000")
	passwordEncrypted, err := encryption.EncryptText("password", aesEncryptionKey)
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, &models.Settings{
		SMTPHost:              "mailhog",
		SMTPPort:              1025,
		SMTPUsername:          "user",
		SMTPPasswordEncrypted: passwordEncrypted,
		SMTPEncryption:        "starttls",
		SMTPFromName:          "Test Sender",
		SMTPFromEmail:         "sender@example.com",
		AESEncryptionKey:      aesEncryptionKey,
	})

	recipient := gofakeit.Email()
	input := &SendEmailInput{
		To:       recipient,
		Subject:  "Test Email",
		HtmlBody: "<p>This is a test email</p>",
	}

	err = emailSender.SendEmail(ctx, input)
	assert.NoError(t, err)
	testutil.AssertEmailSent(t, recipient, "<p>This is a test email</p>")
}
