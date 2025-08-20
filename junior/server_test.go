package junior

import (
	"testing"

	"github.com/anuramat/modagent/testutils"
)

func TestNew(t *testing.T) {
	server := New()
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	if server.BaseServer == nil {
		t.Fatal("Expected BaseServer to be initialized")
	}
}

func TestConfigGetDefaultRole(t *testing.T) {
	config := &Config{}

	tests := []testutils.TableTest{
		{
			Name:     "readonly mode",
			Input:    true,
			Expected: "junior-r",
		},
		{
			Name:     "full access mode",
			Input:    false,
			Expected: "junior-rwx",
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		result := config.GetDefaultRole(tt.Input.(bool))
		testutils.AssertEqual(t, tt.Expected, result)
	})
}
