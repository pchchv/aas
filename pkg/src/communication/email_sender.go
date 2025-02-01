package communication

import (
	"context"

	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/encryption"
	"github.com/pchchv/aas/pkg/src/enums"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
	mail "github.com/xhit/go-simple-mail/v2"
)

type SendEmailInput struct {
	To       string
	Subject  string
	HtmlBody string
}

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (e *EmailSender) SendEmail(ctx context.Context, input *SendEmailInput) error {
	settings := ctx.Value(constants.ContextKeySettings).(*models.Settings)
	server := mail.NewSMTPClient()
	server.Host = settings.SMTPHost
	server.Port = settings.SMTPPort
	if len(settings.SMTPUsername) > 0 {
		server.Username = settings.SMTPUsername
	}

	if len(settings.SMTPPasswordEncrypted) > 0 {
		if decryptedPassword, err := encryption.DecryptText(settings.SMTPPasswordEncrypted, settings.AESEncryptionKey); err != nil {
			return errors.Wrap(err, "unable to decrypt the SMTP password")
		} else {
			server.Password = decryptedPassword
		}
	}

	smtpEnc, err := enums.SMTPEncryptionFromString(settings.SMTPEncryption)
	if err != nil {
		return errors.Wrap(err, "unable to parse the SMTP encryption")
	}

	switch smtpEnc {
	case enums.SMTPEncryptionSSLTLS:
		server.Encryption = mail.EncryptionSSLTLS
	case enums.SMTPEncryptionSTARTTLS:
		server.Encryption = mail.EncryptionSTARTTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	email := mail.NewMSG()
	email.SetFrom(settings.SMTPFromName + " <" + settings.SMTPFromEmail + ">").AddTo(input.To).SetSubject(input.Subject)
	email.SetBody(mail.TextHTML, input.HtmlBody)
	smtpClient, err := server.Connect()
	if err != nil {
		return errors.Wrap(err, "unable to connect to SMTP server")
	}

	if err = email.Send(smtpClient); err != nil {
		return errors.Wrap(err, "unable to send SMTP message")
	}

	return nil
}
