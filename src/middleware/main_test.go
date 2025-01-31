package middleware

import (
	"os"
	"testing"

	"github.com/pchchv/aas/src/config"
)

func TestMain(m *testing.M) {
	config.Init("AuthServer")
	code := m.Run()
	os.Exit(code)
}
