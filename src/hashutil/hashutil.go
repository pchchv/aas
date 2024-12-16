package hashutil

import (
	"crypto/sha256"
	"fmt"

	"github.com/pkg/errors"
)

// HashString can hash strings of any length
func HashString(s string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(s)); err != nil {
		return "", errors.Wrap(err, "unable to hash")
	}

	bs := h.Sum(nil)
	hex := fmt.Sprintf("%x", bs)
	return hex, nil
}

func VerifyStringHash(hashedString string, s string) bool {
	if hash, err := HashString(s); err != nil {
		return false
	} else {
		return hash == hashedString
	}
}
