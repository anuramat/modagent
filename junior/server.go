package junior

import (
	"fmt"

	"github.com/anuramat/modagent/core"
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

func (c *Config) ParseArgs(args map[string]any) (core.CallArgs, error) {
	var a core.CallArgs

	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return a, fmt.Errorf("prompt is required and must be a string")
	}
	a.Prompt = " " + prompt

	if val, ok := args["json_output"].(bool); ok {
		a.JsonOutput = val
	}
	if val, ok := args["conversation"].(string); ok {
		a.Conversation = val
	}
	if val, exists := args["filepaths"]; exists {
		if paths, ok := val.([]interface{}); ok {
			for _, p := range paths {
				if s, ok := p.(string); ok {
					a.Filepaths = append(a.Filepaths, s)
				}
			}
		}
	}
	if val, ok := args["readonly"].(bool); ok {
		a.Readonly = val
	}
	if val, ok := args["bash_cmd"].(string); ok {
		a.BashCmd = val
	}
	if val, ok := args["role"].(string); ok {
		a.Role = val
	}
	return a, nil
}

func (c *Config) GetDefaultRole(readonly bool) string {
	if readonly {
		return "junior-r"
	}
	return "junior-rwx"
}
