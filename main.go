package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type SubagentServer struct{}

func NewSubagentServer() *SubagentServer {
	return &SubagentServer{}
}

func (s *SubagentServer) handleSubagentCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return mcp.NewToolResultError("prompt is required and must be a string"), nil
	}

	jsonOutput := false
	if val, exists := args["json_output"]; exists {
		if b, ok := val.(bool); ok {
			jsonOutput = b
		}
	}

	cmd := exec.Command("mods")
	cmd.Stdin = bytes.NewBufferString(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("command failed: %v, stderr: %s", err, stderr.String())), nil
	}

	output := stdout.String()

	if jsonOutput {
		var result interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to parse JSON: %v", err)), nil
		}
		jsonBytes, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	return mcp.NewToolResultText(output), nil
}

func main() {
	s := server.NewMCPServer(
		"subagent-server",
		"1.0.0",
	)

	subagentServer := NewSubagentServer()

	tool := mcp.NewTool("subagent",
		mcp.WithDescription("Execute a subagent call using the 'mods' command with prompt as stdin"),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("The prompt to pass to the subagent"),
		),
		mcp.WithBoolean("json_output",
			mcp.Description("Whether to parse and return stdout as JSON (default: false)"),
		),
	)

	s.AddTool(tool, subagentServer.handleSubagentCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
