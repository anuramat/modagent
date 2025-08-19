# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Instructions

You MUST NOT commit binaries to the repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server called "modagent" that provides an interface to LLM agents using free models. The server exposes multiple tools: `junior-r`/`junior-rwx` for general AI assistance, and `logworm` for command output analysis.

**Usage**: Use the junior tool PROACTIVELY and frequently since it uses free models. Perfect for getting AI help with any development task.

**Response Schema**: Always returns JSON: `{"response": <content>, "conversation": <id>}`

## Architecture

- **Modular Go application**: Main entry point (`main.go`) with functionality split into packages
- **Junior package** (`junior/`): Contains the core server implementation
  - `server.go`: Server struct with modular role handling logic
  - `description.go`: Embeds the tool description from `description.md`
  - `description.md`: Markdown documentation for the junior tool
- **Logworm package** (`logworm/`): Specialized tool for command output analysis
  - `server.go`: Wraps junior with logworm role
  - `description.go`: Embeds tool description
  - `description.md`: Markdown documentation for logworm
- **Core functionality**: Wraps the `mods` CLI tool to provide LLM capabilities via MCP protocol
- **Tool interfaces**: Exposes three tools:
  - `junior-r` (read-only): General AI assistance with file/bash access disabled
  - `junior-rwx` (full access): General AI assistance with full capabilities
  - `logworm`: Specialized for analyzing command outputs using dedicated role

**Junior tool parameters**:

- `prompt` (required): Your question or request for the LLM
- `json_output` (optional): Parse LLM response as structured JSON
- `conversation` (optional): Continue previous conversation using its ID
- `filepaths` (optional): Array of absolute file paths to include as context
- `bash_cmd` (optional): Bash command to execute, output included in stdin with XML tags
- `role` (optional): Custom mods role to use (overrides readonly/rwx defaults)

**Logworm tool parameters**:

- `bash_cmd` (required): Bash command to execute and analyze its output

## Development Commands

### Build and Run

```bash
# Build using Nix (recommended)
nix build

# Run the built binary
nix run

# Build using Go directly
go build -o modagent main.go

# Run the MCP server
./modagent
```

### Code Formatting

The project uses treefmt with multiple Go formatters:

```bash
# Format code (requires treefmt)
treefmt

# Individual formatters are configured:
# - gofmt: Basic Go formatting
# - gofumpt: Stricter Go formatting
# - goimports: Import organization
```

### Testing and Validation

```bash
# Check if flake builds successfully
nix flake check

# Verify the binary works
echo "test prompt" | ./modagent
```

## Dependencies

- **MCP Framework**: Uses `github.com/mark3labs/mcp-go` for MCP protocol implementation
- **External requirement**: Requires `mods` CLI tool to be available in PATH
- **Go version**: 1.23+

## Key Implementation Details

The server is organized into the following components:

### Main (`main.go`)

- Creates MCP server instance with version "unstable"
- Instantiates junior and logworm servers
- Registers three tools: `junior-r`, `junior-rwx`, and `logworm`
- Serves via stdio for MCP client integration

### Junior Server (`junior/server.go`)

The `HandleCall` method processes requests through these steps:

1. **Parse arguments**: Extract and validate parameters from the request
2. **Build mods command**: Construct command with appropriate flags:
   - `-f --format-as=json` for JSON output
   - `--continue=<id>` for conversation continuation
   - `-R <role>` where role can be:
     - Explicit role from `role` parameter
     - `junior-r` for readonly mode
     - `junior-rwx` for full access mode
     - Custom roles like `logworm` for specialized tools
3. **Prepare stdin**:
   - Execute bash command if provided, wrap output in XML tags
   - Read file contents if provided, wrap in XML tags
4. **Execute mods**: Run the command with prepared stdin
5. **Extract conversation ID**: Parse stderr for "Conversation saved:" pattern
6. **Build response**: Return JSON with response and conversation ID
7. **Handle JSON parsing**: Parse response as JSON when `json_output=true`

The server runs in stdio mode, making it suitable for MCP client integration.
