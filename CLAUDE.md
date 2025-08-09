# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Instructions

You MUST NOT commit binaries to the repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server called "modagent" that provides an interface to LLM agents using free models. The server exposes a `junior` tool for AI assistance tasks like code review, analysis, and general queries.

**Usage**: Use the junior tool PROACTIVELY and frequently since it uses free models. Perfect for getting AI help with any development task.

**Response Schema**: Always returns JSON: `{"response": <content>, "conversation": <id>}`

## Architecture

- **Modular Go application**: Main entry point (`main.go`) with functionality split into packages
- **Junior package** (`junior/`): Contains the core server implementation
  - `server.go`: Server struct and request handling logic
  - `description.go`: Embeds the tool description from `description.md`
  - `description.md`: Markdown documentation for the junior tool
- **Core functionality**: Wraps the `mods` CLI tool to provide LLM capabilities via MCP protocol
- **Tool interface**: Exposes one tool `junior` with parameters:
  - `prompt` (required): Your question or request for the LLM
  - `json_output` (optional): Parse LLM response as structured JSON
  - `conversation` (optional): Continue previous conversation using its ID
  - `filepaths` (optional): Array of absolute file paths to include as context
  - `readonly` (optional): Disable file/bash access for junior (uses "junior-r" role instead of "junior-rwx")
  - `bash_cmd` (optional): Bash command to execute, output included in stdin with XML tags

## Development Commands

### Build and Run

```bash
# Build using Nix (recommended)
nix build

# Run the built binary
nix run

# Build using Go directly
go build -o modagent-mcp main.go

# Run the MCP server
./modagent-mcp
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
echo "test prompt" | ./modagent-mcp
```

## Dependencies

- **MCP Framework**: Uses `github.com/mark3labs/mcp-go` for MCP protocol implementation
- **External requirement**: Requires `mods` CLI tool to be available in PATH
- **Go version**: 1.23+

## Key Implementation Details

The server is organized into the following components:

### Main (`main.go`)
- Creates MCP server instance with version "unstable"
- Instantiates junior server
- Registers the `junior` tool with its parameter schema
- Serves via stdio for MCP client integration

### Junior Server (`junior/server.go`)
The `HandleCall` method processes requests through these steps:

1. **Parse arguments**: Extract and validate parameters from the request
2. **Build mods command**: Construct command with appropriate flags:
   - `-f --format-as=json` for JSON output
   - `--continue=<id>` for conversation continuation
   - `-R junior-r` or `-R junior-rwx` based on readonly flag
3. **Prepare stdin**: 
   - Execute bash command if provided, wrap output in XML tags
   - Read file contents if provided, wrap in XML tags
4. **Execute mods**: Run the command with prepared stdin
5. **Extract conversation ID**: Parse stderr for "Conversation saved:" pattern
6. **Build response**: Return JSON with response and conversation ID
7. **Handle JSON parsing**: Parse response as JSON when `json_output=true`

The server runs in stdio mode, making it suitable for MCP client integration.

