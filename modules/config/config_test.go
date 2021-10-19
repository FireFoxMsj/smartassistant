package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfigTest(t *testing.T) {
	assert.True(t, GetConf().Debug)
	assert.True(t, alreadyInitConfig)
	assert.Equal(t, options, *GetConf())
}

func TestMain(m *testing.M) {
	TestSetup()
	code := m.Run()
	TestTeardown()
	os.Exit(code)
}
