package main

import (
	"fmt"
	"os"

	"github.com/anuramat/modagent/junior"
	"github.com/anuramat/modagent/summarize"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	version := "unstable"
	s := server.NewMCPServer(
		"modagent",
		version,
	)

	jr := junior.New()
	sm := summarize.New()

	juniorRTool := mcp.NewTool("junior-r",
		mcp.WithDescription(junior.Description+" (read-only mode)"),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Your question or request for the junior AI"),
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
		mcp.WithString("bash_cmd",
			mcp.Description("Bash command to execute; junior will receive the command itself, stdout, stderr, and its exit status"),
		),
	)

	juniorRWXTool := mcp.NewTool("junior-rwx",
		mcp.WithDescription(junior.Description+" (full access mode)"),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Your question or request for the junior AI"),
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
		mcp.WithString("bash_cmd",
			mcp.Description("Bash command to execute; junior will receive the command itself, stdout, stderr, and its exit status"),
		),
	)

	summarizeTool := mcp.NewTool("summarize",
		mcp.WithDescription(summarize.Description),
		mcp.WithString("bash_cmd",
			mcp.Required(),
			mcp.Description("Bash command to execute and summarize its output"),
		),
	)

	s.AddTool(juniorRTool, jr.HandleCallReadonly)
	s.AddTool(juniorRWXTool, jr.HandleCall)
	s.AddTool(summarizeTool, sm.HandleCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
