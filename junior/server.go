package junior

import "github.com/anuramat/modagent/core"

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

func (c *Config) GetDefaultRole(readonly bool) string {
	if readonly {
		return "junior-r"
	}
	return "junior-rwx"
}
