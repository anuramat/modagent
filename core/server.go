package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type CallArgs struct {
	Prompt       string
	JsonOutput   bool
	Conversation string
	Filepaths    []string
	Readonly     bool
	BashCmd      string
	Role         string
}

type ServerConfig interface {
	ParseArgs(args map[string]any) (CallArgs, error)
	GetDefaultRole(readonly bool) string
}

type BaseServer struct {
	config ServerConfig
}

func NewBaseServer(config ServerConfig) *BaseServer {
	return &BaseServer{config: config}
}

func (s *BaseServer) HandleCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCallWithReadonly(ctx, request, false)
}

func (s *BaseServer) HandleCallReadonly(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCallWithReadonly(ctx, request, true)
}

func (s *BaseServer) handleCallWithReadonly(ctx context.Context, request mcp.CallToolRequest, readonly bool) (*mcp.CallToolResult, error) {
	params, err := s.config.ParseArgs(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params.Readonly = readonly

	cmd := buildModsCmd(params, s.config.GetDefaultRole)

	stdin, tempDir, err := prepareStdin(params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cmd.Stdin = &stdin

	stdout, stderr, err := runCommand(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("command failed: %v, stderr: %s", err, stderr)), nil
	}

	conversationID := extractConversationID(stderr)

	result, err := buildResponse(stdout, conversationID, tempDir, params.JsonOutput)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result), nil
}

func buildModsCmd(a CallArgs, getDefaultRole func(bool) string) *exec.Cmd {
	cmdArgs := []string{}
	if a.JsonOutput {
		cmdArgs = append(cmdArgs, "-j")
	}
	if a.Conversation != "" {
		cmdArgs = append(cmdArgs, "--continue="+a.Conversation)
	}
	if a.Role != "" {
		cmdArgs = append(cmdArgs, "-R", a.Role)
	} else {
		cmdArgs = append(cmdArgs, "-R", getDefaultRole(a.Readonly))
	}
	cmdArgs = append(cmdArgs, a.Prompt)
	return exec.Command("mods", cmdArgs...)
}

func prepareStdin(a CallArgs) (bytes.Buffer, string, error) {
	var stdinBuffer bytes.Buffer
	var tempDir string

	if a.BashCmd != "" {
		bashExec := exec.Command("bash", "-c", a.BashCmd)
		var bashStdout, bashStderr bytes.Buffer
		bashExec.Stdout = &bashStdout
		bashExec.Stderr = &bashStderr

		exitStatus := 0
		if err := bashExec.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitStatus = exitError.ExitCode()
			} else {
				exitStatus = 1
			}
		}

		// Create temp directory and save outputs
		baseTmpDir := os.Getenv("TMPDIR")
		if baseTmpDir == "" {
			baseTmpDir = "/tmp"
		}
		timestamp := time.Now().Format("20060102-150405-000000")
		tempDir = filepath.Join(baseTmpDir, fmt.Sprintf("modagent-%s", timestamp))

		if err := os.MkdirAll(tempDir, 0o755); err != nil {
			return stdinBuffer, "", fmt.Errorf("failed to create temp directory: %v", err)
		}

		if err := os.WriteFile(filepath.Join(tempDir, "stdout"), bashStdout.Bytes(), 0o644); err != nil {
			return stdinBuffer, "", fmt.Errorf("failed to write stdout file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, "stderr"), bashStderr.Bytes(), 0o644); err != nil {
			return stdinBuffer, "", fmt.Errorf("failed to write stderr file: %v", err)
		}

		stdinBuffer.WriteString(fmt.Sprintf("<bash command=\"%s\" exit_status=\"%d\"><stdout>%s</stdout><stderr>%s</stderr></bash>\n", a.BashCmd, exitStatus, bashStdout.String(), bashStderr.String()))
	}

	for _, filepath := range a.Filepaths {
		content, err := os.ReadFile(filepath)
		if err != nil {
			return stdinBuffer, tempDir, fmt.Errorf("failed to read file %s: %v", filepath, err)
		}
		stdinBuffer.WriteString(fmt.Sprintf("<file path=%s>\n%s</file path=%s>\n", filepath, string(content), filepath))
	}

	return stdinBuffer, tempDir, nil
}

func runCommand(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func extractConversationID(stderrOutput string) string {
	conversationID := ""
	if lines := strings.Split(stderrOutput, "\n"); len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if lastLine == "" && len(lines) > 1 {
			lastLine = lines[len(lines)-2]
		}
		re := regexp.MustCompile(`Conversation saved:\s+(\w+)`)
		if matches := re.FindStringSubmatch(lastLine); len(matches) > 1 {
			conversationID = matches[1]
		}
	}
	return conversationID
}

func buildResponse(output, conversationID, tempDir string, jsonOutput bool) (string, error) {
	responseObj := map[string]any{
		"response":     output,
		"conversation": conversationID,
	}

	if tempDir != "" {
		responseObj["temp_dir"] = tempDir
	}

	if jsonOutput {
		var result any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			return "", fmt.Errorf("failed to parse JSON: %v", err)
		}
		responseObj["response"] = result
	}

	jsonBytes, _ := json.Marshal(responseObj)
	return string(jsonBytes), nil
}
