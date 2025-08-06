package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
	prompt = " " + prompt

	jsonOutput := false
	if val, exists := args["json_output"]; exists {
		if b, ok := val.(bool); ok {
			jsonOutput = b
		}
	}

	conversation := ""
	if val, exists := args["conversation"]; exists {
		if s, ok := val.(string); ok {
			conversation = s
		}
	}

	filepath := ""
	if val, exists := args["filepath"]; exists {
		if s, ok := val.(string); ok {
			filepath = s
		}
	}

	var cmd *exec.Cmd
	cmdArgs := []string{}
	if jsonOutput {
		cmdArgs = append(cmdArgs, "-f", "--format-as=json")
	}
	if conversation != "" {
		cmdArgs = append(cmdArgs, "--continue="+conversation)
	}
	cmdArgs = append(cmdArgs, prompt)
	cmd = exec.Command("mods", cmdArgs...)

	// Handle stdin based on filepath
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to open file %s: %v", filepath, err)), nil
		}
		defer file.Close()
		cmd.Stdin = file
	} else {
		cmd.Stdin = bytes.NewBufferString("")
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("command failed: %v, stderr: %s", err, stderr.String())), nil
	}

	output := stdout.String()
	stderrOutput := stderr.String()

	// Extract conversation ID from stderr
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

	// Wrap response in JSON object
	responseObj := map[string]interface{}{
		"response":     output,
		"conversation": conversationID,
	}

	if jsonOutput {
		var result interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to parse JSON: %v", err)), nil
		}
		responseObj["response"] = result
	}

	jsonBytes, _ := json.Marshal(responseObj)
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func main() {
	s := server.NewMCPServer(
		"subagent-server",
		"1.0.0",
	)

	subagentServer := NewSubagentServer()

	tool := mcp.NewTool("subagent",
		mcp.WithDescription("Query an LLM agent with free models. Use often for AI assistance, analysis, code review, etc."),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Your question or request for the LLM"),
		),
		mcp.WithBoolean("json_output",
			mcp.Description("Parse LLM response as structured JSON"),
		),
		mcp.WithString("conversation",
			mcp.Description("Continue previous conversation using its ID"),
		),
		mcp.WithString("filepath",
			mcp.Description("File path to include as context (sent as stdin)"),
		),
	)

	s.AddTool(tool, subagentServer.handleSubagentCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
