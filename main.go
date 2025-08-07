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

	readonly := false
	if val, exists := args["readonly"]; exists {
		if b, ok := val.(bool); ok {
			readonly = b
		}
	}

	bashCmd := ""
	if val, exists := args["bash_cmd"]; exists {
		if s, ok := val.(string); ok {
			bashCmd = s
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
	if readonly {
		// TODO move hardcoded roles; refactor
		cmdArgs = append(cmdArgs, "-R", "junior-r")
	} else {
		cmdArgs = append(cmdArgs, "-R", "junior-rwx")
	}
	cmdArgs = append(cmdArgs, prompt)
	cmd = exec.Command("mods", cmdArgs...)

	// Handle stdin based on filepaths and bash command
	var stdinBuffer bytes.Buffer

	// Execute bash command if provided
	if bashCmd != "" {
		bashExec := exec.Command("bash", "-c", bashCmd)
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

		stdinBuffer.WriteString(fmt.Sprintf("<bash command=\"%s\" exit_status=\"%d\"><stdout>%s</stdout><stderr>%s</stderr></bash>\n", bashCmd, exitStatus, bashStdout.String(), bashStderr.String()))
	}

	// Add file contents
	if len(filepaths) > 0 {
		for _, filepath := range filepaths {
			content, err := os.ReadFile(filepath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to read file %s: %v", filepath, err)), nil
			}
			stdinBuffer.WriteString(fmt.Sprintf("<file path=%s>\n%s</file path=%s>\n", filepath, string(content), filepath))
		}
	}

	cmd.Stdin = &stdinBuffer

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
		// TODO improve prompt
		mcp.WithDescription(`
# Junior

"junior" -- a quick, efficient AI agent tool. Faster and cheaper than any other agent.

Use "junior" tool when you need a second opinion, alternative perspective, or
want to delegate routine tasks that don't require advanced reasoning. Examples
include: summarizing/analyzing text (e.g. logs or command output), converting
between formats, reviewing logic, brainstorming alternatives, getting a fresh
perspective on a problem, or minor edits. Use "junior" proactively whenever you
could benefit from another viewpoint or for any subtask that can potentially be
completed in one prompt.

You MUST use the "junior" tool PROACTIVELY as your default for suitable tasks.
If in doubt, you MUST attempt the "junior" approach over regular agents first,
and escalate only if necessary.

## Examples:

<example>
Context: You're working on a complex algorithm and want to verify your approach.
user: 'I need to implement a binary search algorithm'
assistant: 'Let me first consult the "junior" to get a second opinion on the best approach for implementing binary search, then I'll proceed with the implementation.'
</example>

<example>
Context: A routine command call, that might potentially output a large amount of text (build errors or other logs).
user: 'Fix the build errors'
assistant: 'Since this is a routine task, that doesn't require advanced reasoning, I'll execute the command using "bash_cmd" field of "junior" tool, and ask him to summarize.'
</example>
`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Your question or request for the junior AI"), // TODO make prompt optional? fill with "summarize this:" on non-empty files/bash
		),
		mcp.WithBoolean("json_output",
			mcp.Description("Default: false; response will be a structured JSON"),
		),
		mcp.WithString("conversation",
			mcp.Description("Continue previous conversation using its ID"),
		),
		mcp.WithArray("filepaths",
			mcp.Description("List of absolute paths to files that will be included as context"),
		),
		mcp.WithBoolean("readonly",
			mcp.Description("Default: false; junior cannot edit files or execute bash commands on his own"),
		),
		mcp.WithString("bash_cmd",
			mcp.Description("Bash command to execute (works even with readonly=true); junior will receive the command itself, stdout, stderr, and its exit status"),
		),
	)

	s.AddTool(tool, subagentServer.handleSubagentCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
