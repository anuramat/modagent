package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// SetupTestConfig creates a temporary config directory for tests
func SetupTestConfig(t *testing.T) (string, func()) {
	originalConfigHome := xdg.ConfigHome
	tempDir, err := os.MkdirTemp("", "modagent-test-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp config dir: %v", err)
	}

	// Override XDG config home
	xdg.ConfigHome = tempDir

	cleanup := func() {
		xdg.ConfigHome = originalConfigHome
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// WriteTestConfig writes a test config file
func WriteTestConfig(t *testing.T, configDir string, content string) {
	configPath := filepath.Join(configDir, "modagent", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}

// ParseJSONResponse parses a JSON response string
func ParseJSONResponse(t *testing.T, response string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	return result
}

// GetTestFixturePath returns path to test fixture file
func GetTestFixturePath(filename string) string {
	return filepath.Join("testdata", filename)
}

// CreateTestFile creates a temporary test file with content
func CreateTestFile(t *testing.T, content string) (string, func()) {
	tempFile, err := os.CreateTemp("", "modagent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	cleanup := func() { os.Remove(tempFile.Name()) }
	return tempFile.Name(), cleanup
}

// AssertJSONEqual compares two JSON strings for equality
func AssertJSONEqual(t *testing.T, expected, actual string) {
	var expectedObj, actualObj interface{}

	if err := json.Unmarshal([]byte(expected), &expectedObj); err != nil {
		t.Fatalf("Failed to parse expected JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(actual), &actualObj); err != nil {
		t.Fatalf("Failed to parse actual JSON: %v", err)
	}

	expectedBytes, _ := json.Marshal(expectedObj)
	actualBytes, _ := json.Marshal(actualObj)

	if string(expectedBytes) != string(actualBytes) {
		t.Fatalf("JSON not equal:\nExpected: %s\nActual: %s", expectedBytes, actualBytes)
	}
}

// TableTest represents a table-driven test case
type TableTest struct {
	Name     string
	Input    interface{}
	Expected interface{}
	WantErr  bool
}

// RunTableTests executes table-driven tests
func RunTableTests(t *testing.T, tests []TableTest, testFunc func(*testing.T, TableTest)) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testFunc(t, tt)
		})
	}
}
