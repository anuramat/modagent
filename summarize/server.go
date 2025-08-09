package summarize

import (
	"context"

	"github.com/anuramat/modagent/junior"
	"github.com/mark3labs/mcp-go/mcp"
)

type Server struct {
	juniorServer *junior.Server
}

func New() *Server {
	return &Server{
		juniorServer: junior.New(),
	}
}

func (s *Server) HandleCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	bashCmd, ok := args["bash_cmd"].(string)
	if !ok || bashCmd == "" {
		return mcp.NewToolResultError("bash_cmd is required and must be a string"), nil
	}

	juniorArgs := map[string]any{
		"prompt":   "Summarize this command output",
		"bash_cmd": bashCmd,
		"readonly": true,
	}

	juniorRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "junior",
			Arguments: juniorArgs,
		},
	}

	return s.juniorServer.HandleCall(ctx, juniorRequest)
}
