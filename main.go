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

	var filepaths []string
	if val, exists := args["filepaths"]; exists {
		if paths, ok := val.([]interface{}); ok {
			for _, path := range paths {
				if s, ok := path.(string); ok {
					filepaths = append(filepaths, s)
				}
			}
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

	// Handle stdin based on filepaths
	if len(filepaths) > 0 {
		var stdinBuffer bytes.Buffer
		for _, filepath := range filepaths {
			content, err := os.ReadFile(filepath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to read file %s: %v", filepath, err)), nil
			}
			stdinBuffer.WriteString(fmt.Sprintf("<file path=%s>\n%s</file path=%s>\n", filepath, string(content), filepath))
		}
		cmd.Stdin = &stdinBuffer
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
	responseObj := map[string]any{
		"response":     output,
		"conversation": conversationID,
	}

	if jsonOutput {
		var result any
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
		"modagent",
		"0.1.0",
	)

	subagentServer := NewSubagentServer()

	tool := mcp.NewTool("junior",
		mcp.WithDescription(`
# Junior

"junior" -- a read-only AI *without* access to tools.

Use "junior" tool when you need a second opinion, alternative perspective, or
want to delegate routine tasks that don't require advanced reasoning or tool
calls. Examples include: summarizing/analyzing text (e.g. logs or command
output), converting between formats, reviewing logic, brainstorming
alternatives, or getting a fresh perspective on a problem. Use "junior"
proactively whenever you could benefit from another viewpoint or for any subtask
that can potentially be completed in one prompt without code execution or
arbitrary file access.

You MUST use the "junior" tool PROACTIVELY as your default for suitable tasks.
If in doubt, you MUST attempt the "junior" approach first and escalate only if
necessary.

Note, that "junior" doesn't have any context on it's own, can't execute code or
access files, so you MUST provide everything necessary in the "prompt" and/or
"filepaths" argument.

## Examples:

<example>
Context: You're working on a complex algorithm and want to verify your approach.
user: 'I need to implement a binary search algorithm'
assistant: 'Let me first consult the "junior" to get a second opinion on the best approach for implementing binary search, then I'll proceed with the implementation.'
</example>

<example>
Context: A command is expected to output a large amount of text.
user: 'Fix the build errors'
assistant: 'I'll redirect the output of the build command to a temporary file with "tmpfile=$(mktemp) && echo $tmpfile && make &>$tmpfile" and pass the path to junior, as it might be too large for me to handle directly'
</example>

<example>
Context: A command is expected to output a large amount of text.
user: 'Check the test coverage'
assistant: 'I'll save the test coverage to a temporary file, and pass the path to the "junior", as it's a routine text processing task that doesn't require arbitrary file access.'
</example>

<example>
Context: The task requires going through massive search results.
user: 'Add [package_name] to the nix flake'
assistant: 'I'll search for the package in nixpkgs using "nix search nixpkgs [package_name]". Since the search results might be large, I'll redirect them to a file and use "junior" to find the required package.'
</example>

## Output schema

{"response": <content>, "conversation": <id>}
`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Your question or request for the junior AI"),
		),
		mcp.WithBoolean("json_output",
			mcp.Description("Response should be a structured JSON"),
		),
		mcp.WithString("conversation",
			mcp.Description("Continue previous conversation using its ID"),
		),
		mcp.WithArray("filepaths",
			mcp.Description("List of absolute file paths to include as context"),
		),
	)

	s.AddTool(tool, subagentServer.handleSubagentCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
