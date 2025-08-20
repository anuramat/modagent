package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anuramat/modagent/testutils"
)

func TestLoadConfig(t *testing.T) {
	tests := []testutils.TableTest{
		{
			Name:     "no config file",
			Input:    "",
			Expected: &Config{Tools: make(map[string]ToolConfig)},
			WantErr:  false,
		},
		{
			Name: "valid config",
			Input: `tools:
  junior-r:
    description:
      text: "Custom junior-r description"
  logworm:
    description:
      text: "Custom logworm description"`,
			Expected: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Text: stringPtr("Custom junior-r description")}},
					"logworm":  {Description: Description{Text: stringPtr("Custom logworm description")}},
				},
			},
			WantErr: false,
		},
		{
			Name:     "invalid yaml",
			Input:    "invalid: yaml: content: [",
			Expected: nil,
			WantErr:  true,
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		configDir, cleanup := testutils.SetupTestConfig(t)
		defer cleanup()

		if tt.Input != "" {
			testutils.WriteTestConfig(t, configDir, tt.Input.(string))
		}

		cfg, err := LoadConfig()
		if tt.WantErr {
			testutils.AssertError(t, err)
			return
		}

		testutils.AssertNoError(t, err)
		if tt.Expected != nil {
			expected := tt.Expected.(*Config)
			testutils.AssertEqual(t, len(expected.Tools), len(cfg.Tools))
		}
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []testutils.TableTest{
		{
			Name: "valid config",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Text: stringPtr("test")}},
				},
			},
			WantErr: false,
		},
		{
			Name: "unknown tool",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"unknown": {Description: Description{Text: stringPtr("test")}},
				},
			},
			WantErr: true,
		},
		{
			Name: "text and path both set",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{
						Text: stringPtr("test"),
						Path: stringPtr("file.md"),
					}},
				},
			},
			WantErr: true,
		},
		{
			Name: "path file not exists",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Path: stringPtr("nonexistent.md")}},
				},
			},
			WantErr: true,
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		_, cleanup := testutils.SetupTestConfig(t)
		defer cleanup()

		cfg := tt.Input.(*Config)
		err := validateConfig(cfg)

		if tt.WantErr {
			testutils.AssertError(t, err)
		} else {
			testutils.AssertNoError(t, err)
		}
	})
}

func TestGenerateDefaultConfig(t *testing.T) {
	_, cleanup := testutils.SetupTestConfig(t)
	defer cleanup()

	err := GenerateDefaultConfig()
	testutils.AssertNoError(t, err)

	// Load and validate generated config
	cfg, err := LoadConfig()
	testutils.AssertNoError(t, err)
	testutils.AssertEqual(t, 3, len(cfg.Tools))

	// Check all tools are present
	for _, toolName := range validToolNames {
		if _, exists := cfg.Tools[toolName]; !exists {
			t.Fatalf("Tool %s not found in generated config", toolName)
		}
	}
}

func TestGetToolDescription(t *testing.T) {
	configDir, cleanup := testutils.SetupTestConfig(t)
	defer cleanup()

	// Create test description file
	descFile := filepath.Join(configDir, "modagent", "test.md")
	err := os.MkdirAll(filepath.Dir(descFile), 0o755)
	testutils.AssertNoError(t, err)
	err = os.WriteFile(descFile, []byte("File description"), 0o644)
	testutils.AssertNoError(t, err)

	tests := []testutils.TableTest{
		{
			Name: "text description",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Text: stringPtr("Custom description")}},
				},
			},
			Expected: "Custom description",
		},
		{
			Name: "path description",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Path: stringPtr("test.md")}},
				},
			},
			Expected: "File description",
		},
		{
			Name: "empty text description",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Text: stringPtr("")}},
				},
			},
			Expected: "",
		},
		{
			Name:     "no config for tool",
			Input:    &Config{Tools: make(map[string]ToolConfig)},
			Expected: "default description",
		},
		{
			Name: "absolute path",
			Input: &Config{
				Tools: map[string]ToolConfig{
					"junior-r": {Description: Description{Path: stringPtr(descFile)}},
				},
			},
			Expected: "File description",
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		cfg := tt.Input.(*Config)
		result := cfg.GetToolDescription("junior-r", "default description")
		testutils.AssertEqual(t, tt.Expected, result)
	})
}

func TestGetToolDescriptionFileReadError(t *testing.T) {
	_, cleanup := testutils.SetupTestConfig(t)
	defer cleanup()

	cfg := &Config{
		Tools: map[string]ToolConfig{
			"junior-r": {Description: Description{Path: stringPtr("nonexistent.md")}},
		},
	}

	// Should return default when file can't be read
	result := cfg.GetToolDescription("junior-r", "default description")
	testutils.AssertEqual(t, "default description", result)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
