package logworm

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/anuramat/modagent/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type Server struct {
	*core.BaseServer
	passthroughThreshold int
}

type Config struct{}

func New(passthroughThreshold int) *Server {
	config := &Config{}
	return &Server{
		BaseServer:           core.NewBaseServer(config),
		passthroughThreshold: passthroughThreshold,
	}
}

func (s *Server) HandleCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	bashCmd, ok := args["bash_cmd"].(string)
	if !ok || bashCmd == "" {
		return mcp.NewToolResultError("bash_cmd is required and must be a string"), nil
	}

	// Execute command and check output length for passthrough
	cmd := exec.CommandContext(ctx, "bash", "-c", bashCmd)
	output, err := cmd.Output()
	if err != nil {
		return mcp.NewToolResultError("Failed to execute command: " + err.Error()), nil
	}

	// If output is shorter than threshold, return it directly
	if len(output) < s.passthroughThreshold {
		response := map[string]interface{}{
			"response":     string(output),
			"conversation": "",
		}
		jsonResponse, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}

	// Otherwise, use the normal logworm processing
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
