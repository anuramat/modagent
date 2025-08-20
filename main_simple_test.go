package main

import (
	"testing"

	"github.com/anuramat/modagent/config"
	"github.com/anuramat/modagent/testutils"
)

func TestGenerateConfigFunction(t *testing.T) {
	_, cleanup := testutils.SetupTestConfig(t)
	defer cleanup()

	err := config.GenerateDefaultConfig()
	testutils.AssertNoError(t, err)

	// Load and validate generated config
	cfg, err := config.LoadConfig()
	testutils.AssertNoError(t, err)
	testutils.AssertEqual(t, 3, len(cfg.Tools))
}
