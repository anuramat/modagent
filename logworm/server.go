package logworm

import (
	"context"

	"github.com/anuramat/modagent/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type Server struct {
	*core.BaseServer
}

type Config struct{}

func New() *Server {
	config := &Config{}
	return &Server{
		BaseServer: core.NewBaseServer(config),
	}
}

func (s *Server) HandleCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	bashCmd, ok := args["bash_cmd"].(string)
	if !ok || bashCmd == "" {
		return mcp.NewToolResultError("bash_cmd is required and must be a string"), nil
	}

	coreArgs := map[string]any{
		"prompt":   "Parse and analyze this command output",
		"bash_cmd": bashCmd,
		"role":     "logworm",
	}

	coreRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "logworm",
			Arguments: coreArgs,
		},
	}

	return s.BaseServer.HandleCall(ctx, coreRequest)
}

func (c *Config) GetDefaultRole(readonly bool) string {
	return "logworm"
}
