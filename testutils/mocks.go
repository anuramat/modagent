package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockConfig for testing core.ServerConfig interface
type MockConfig struct {
	DefaultRoles map[bool]string
}

func (m *MockConfig) GetDefaultRole(readonly bool) string {
	if role, exists := m.DefaultRoles[readonly]; exists {
		return role
	}
	return "default"
}

// MockExecCommand replaces exec.Command in tests
var MockExecCommand = func(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestMockCommand", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// MockCommandResults stores expected outputs for commands
var MockCommandResults = map[string]MockResult{}

type MockResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// SetMockCommand configures expected result for a command
func SetMockCommand(cmdLine string, result MockResult) {
	MockCommandResults[cmdLine] = result
}

// ClearMockCommands resets all mock commands
func ClearMockCommands() {
	MockCommandResults = map[string]MockResult{}
}

// TestMockCommand is a helper for mocked exec.Command calls
func TestMockCommand() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmdLine := strings.Join(args, " ")
	if result, exists := MockCommandResults[cmdLine]; exists {
		fmt.Fprint(os.Stdout, result.Stdout)
		fmt.Fprint(os.Stderr, result.Stderr)
		os.Exit(result.ExitCode)
	}

	// Default behavior for unmocked commands
	fmt.Fprintf(os.Stderr, "Unmocked command: %s\n", cmdLine)
	os.Exit(1)
}

// MockFileSystem provides file system mocking utilities
type MockFileSystem struct {
	Files map[string]string
	Dirs  map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string]string),
		Dirs:  make(map[string]bool),
	}
}

func (m *MockFileSystem) WriteFile(path, content string) {
	m.Files[path] = content
	m.Dirs[filepath.Dir(path)] = true
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, exists := m.Files[path]; exists {
		return []byte(content), nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Exists(path string) bool {
	_, fileExists := m.Files[path]
	dirExists := m.Dirs[path]
	return fileExists || dirExists
}

// CreateTempDir creates a temporary directory for tests
func CreateTempDir(prefix string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tempDir) }
	return tempDir, cleanup, nil
}

// CreateMCPRequest creates a mock MCP CallToolRequest
func CreateMCPRequest(toolName string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}
}

// AssertNoError fails test if error is not nil
func AssertNoError(t interface{ Fatalf(string, ...interface{}) }, err error) {
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError fails test if error is nil
func AssertError(t interface{ Fatalf(string, ...interface{}) }, err error) {
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

// AssertEqual fails test if values are not equal
func AssertEqual(t interface{ Fatalf(string, ...interface{}) }, expected, actual interface{}) {
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertContains fails test if str doesn't contain substr
func AssertContains(t interface{ Fatalf(string, ...interface{}) }, str, substr string) {
	if !strings.Contains(str, substr) {
		t.Fatalf("Expected %q to contain %q", str, substr)
	}
}

// MockModsOutput configures expected mods command output
func MockModsOutput(prompt, role string, stdout, stderr string, exitCode int) {
	args := []string{"-R", role, prompt}
	cmdLine := "mods " + strings.Join(args, " ")
	SetMockCommand(cmdLine, MockResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	})
}

// MockBashOutput configures expected bash command output
func MockBashOutput(cmd, stdout, stderr string, exitCode int) {
	cmdLine := "bash -c " + cmd
	SetMockCommand(cmdLine, MockResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	})
}
