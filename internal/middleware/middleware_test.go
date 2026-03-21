package middleware

import (
	"os"
	"testing"

	"github.com/vekshinnikita/pulse_watch/internal/testutils"
)

func TestMain(m *testing.M) {
	// подготовка перед тестами
	testutils.MiddlewaresSetup()

	// запуск всех тестов
	code := m.Run()

	os.Exit(code)
}
