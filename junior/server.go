package junior

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

type Server struct{}

func New() *Server {
	return &Server{}
}

type callArgs struct {
	prompt       string
	jsonOutput   bool
	conversation string
	filepaths    []string
	readonly     bool
	bashCmd      string
	role         string
}

func parseCallArgs(args map[string]any) (callArgs, error) {
	var a callArgs

	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return a, fmt.Errorf("prompt is required and must be a string")
	}
	a.prompt = " " + prompt

	if val, ok := args["json_output"].(bool); ok {
		a.jsonOutput = val
	}
	if val, ok := args["conversation"].(string); ok {
		a.conversation = val
	}
	if val, exists := args["filepaths"]; exists {
		if paths, ok := val.([]interface{}); ok {
			for _, p := range paths {
				if s, ok := p.(string); ok {
					a.filepaths = append(a.filepaths, s)
				}
			}
		}
	}
	if val, ok := args["readonly"].(bool); ok {
		a.readonly = val
	}
	if val, ok := args["bash_cmd"].(string); ok {
		a.bashCmd = val
	}
	if val, ok := args["role"].(string); ok {
		a.role = val
	}
	return a, nil
}

func buildModsCmd(a callArgs) *exec.Cmd {
	cmdArgs := []string{}
	if a.jsonOutput {
		cmdArgs = append(cmdArgs, "-f", "--format-as=json")
	}
	if a.conversation != "" {
		cmdArgs = append(cmdArgs, "--continue="+a.conversation)
	}
	if a.role != "" {
		cmdArgs = append(cmdArgs, "-R", a.role)
	} else if a.readonly {
		cmdArgs = append(cmdArgs, "-R", "junior-r")
	} else {
		cmdArgs = append(cmdArgs, "-R", "junior-rwx")
	}
	cmdArgs = append(cmdArgs, a.prompt)
	return exec.Command("mods", cmdArgs...)
}

func prepareStdin(a callArgs) (bytes.Buffer, string, error) {
	var stdinBuffer bytes.Buffer
	var tempDir string

	if a.bashCmd != "" {
		bashExec := exec.Command("bash", "-c", a.bashCmd)
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

		stdinBuffer.WriteString(fmt.Sprintf("<bash command=\"%s\" exit_status=\"%d\"><stdout>%s</stdout><stderr>%s</stderr></bash>\n", a.bashCmd, exitStatus, bashStdout.String(), bashStderr.String()))
	}

	for _, filepath := range a.filepaths {
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

func (s *Server) HandleCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCallWithReadonly(ctx, request, false)
}

func (s *Server) HandleCallReadonly(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCallWithReadonly(ctx, request, true)
}

func (s *Server) handleCallWithReadonly(ctx context.Context, request mcp.CallToolRequest, readonly bool) (*mcp.CallToolResult, error) {
	params, err := parseCallArgs(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params.readonly = readonly

	cmd := buildModsCmd(params)

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

	result, err := buildResponse(stdout, conversationID, tempDir, params.jsonOutput)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result), nil
}
