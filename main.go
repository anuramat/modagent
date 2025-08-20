package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anuramat/modagent/config"
	"github.com/anuramat/modagent/junior"
	"github.com/anuramat/modagent/logworm"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	generateConfig := flag.Bool("generate-config", false, "Generate default config file and exit")
	flag.Parse()

	if *generateConfig {
		if err := config.GenerateDefaultConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate config: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	version := "unstable"
	s := server.NewMCPServer(
		"modagent",
		version,
	)

	jr := junior.New()
	lw := logworm.New()

	juniorParams := []mcp.ToolOption{
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Your question or request for the junior AI")),
		mcp.WithBoolean("json_output", mcp.Description("Default: false; response will be a structured JSON")),
		mcp.WithString("conversation", mcp.Description("Continue previous conversation using its ID")),
		mcp.WithArray("filepaths", mcp.Description("List of absolute paths to files that will be included as context")),
		mcp.WithString("bash_cmd", mcp.Description("Bash command to execute; junior will receive the command itself, stdout, stderr, and its exit status")),
	}

	juniorRTool := mcp.NewTool("junior-r", append([]mcp.ToolOption{
		mcp.WithDescription(cfg.GetToolDescription("junior-r", junior.Description+" (read-only mode)")),
	}, juniorParams...)...)

	juniorRWXTool := mcp.NewTool("junior-rwx", append([]mcp.ToolOption{
		mcp.WithDescription(cfg.GetToolDescription("junior-rwx", junior.Description+" (full access mode)")),
	}, juniorParams...)...)

	logwormTool := mcp.NewTool("logworm",
		mcp.WithDescription(cfg.GetToolDescription("logworm", logworm.Description)),
		mcp.WithString("bash_cmd",
			mcp.Required(),
			mcp.Description("Bash command to execute and analyze its output"),
		),
	)

	s.AddTool(juniorRTool, jr.HandleCallReadonly)
	s.AddTool(juniorRWXTool, jr.HandleCall)
	s.AddTool(logwormTool, lw.HandleCall)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
