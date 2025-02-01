package otp

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"strings"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

type OTPSecretGenerator struct {
}

func NewOTPSecretGenerator() *OTPSecretGenerator {
	return &OTPSecretGenerator{}
}

// Returns base64 of QR code image, secret key
func (g *OTPSecretGenerator) GenerateOTPSecret(email string, appName string) (string, string, error) {
	if strings.TrimSpace(email) == "" {
		return "", "", errors.New("email is empty")
	}

	if strings.TrimSpace(appName) == "" {
		return "", "", errors.New("app name is empty")
	}

	if len(email) > 64 {
		return "", "", errors.New("email is too long")
	}

	if len(appName) > 32 {
		return "", "", errors.New("app name is too long")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appName,
		AccountName: email,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "unable to generate otp for user "+email)
	}

	img, err := key.Image(180, 180)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to generate otp png image for user "+email)
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		return "", "", errors.Wrap(err, "unable to encode otp png image for user "+email)
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Str, key.Secret(), nil
}
