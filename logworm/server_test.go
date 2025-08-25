package logworm

import (
	"testing"

	"github.com/anuramat/modagent/testutils"
)

func TestNew(t *testing.T) {
	server := New(2000)
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	if server.BaseServer == nil {
		t.Fatal("Expected BaseServer to be initialized")
	}
	if server.passthroughThreshold != 2000 {
		t.Fatalf("Expected passthrough threshold to be 2000, got %d", server.passthroughThreshold)
	}
}

func TestConfigGetDefaultRole(t *testing.T) {
	config := &Config{}

	tests := []testutils.TableTest{
		{
			Name:     "readonly mode",
			Input:    true,
			Expected: "logworm",
		},
		{
			Name:     "full access mode",
			Input:    false,
			Expected: "logworm",
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		result := config.GetDefaultRole(tt.Input.(bool))
		testutils.AssertEqual(t, tt.Expected, result)
	})
}
