package main

import (
	"fmt"
	"os"

	"github.com/anuramat/modagent/junior"
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

	tool := mcp.NewTool("junior",
		mcp.WithDescription(junior.Description),
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

	s.AddTool(tool, jr.HandleCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
