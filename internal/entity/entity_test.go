package entity

import (
	"github.com/zhiting-tech/smartassistant/internal/config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config.TestSetup()
	code := m.Run()
	config.TestTeardown()
	os.Exit(code)
}
