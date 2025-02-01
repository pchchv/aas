package inputsanitizer

import (
	"log/slog"

	"github.com/sym01/htmlsanitizer"
)

type InputSanitizer struct {
}

func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{}
}

func (i *InputSanitizer) Sanitize(str string) string {
	if sanitizedHTML, err := htmlsanitizer.SanitizeString(str); err != nil {
		slog.Error("unable to sanitize string: " + err.Error())
		return str
	} else {
		return sanitizedHTML
	}
}
