package entity

import (
	"os"
	"testing"

	"github.com/zhiting-tech/smartassistant/modules/config"
)

func TestMain(m *testing.M) {
	config.TestSetup()
	code := m.Run()
	config.TestTeardown()
	os.Exit(code)
}
