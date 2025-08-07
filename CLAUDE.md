# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Instructions

You MUST NOT commit binaries to the repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server called "modagent" that provides an interface to LLM agents using free models. The server exposes a `junior` tool for AI assistance tasks like code review, analysis, and general queries.

**Usage**: Use the junior tool PROACTIVELY and frequently since it uses free models. Perfect for getting AI help with any development task.

**Response Schema**: Always returns JSON: `{"response": <content>, "conversation": <id>}`

## Architecture

- **Single file Go application** (`main.go`): Contains the complete MCP server implementation
- **Core functionality**: Wraps the `mods` CLI tool to provide LLM capabilities via MCP protocol
- **Tool interface**: Exposes one tool `junior` with parameters:
  - `prompt` (required): Your question or request for the LLM
  - `json_output` (optional): Parse LLM response as structured JSON
  - `conversation` (optional): Continue previous conversation using its ID
  - `filepaths` (optional): Array of absolute file paths to include as context
  - `readonly` (optional): Disable tools access by adding --mcp-disable=tools to mods call
  - `bash_cmd` (optional): Bash command to execute before mods call, output included in stdin

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

The server implements a single tool handler `handleSubagentCall` that:

1. Validates the required `prompt` parameter
2. Optionally enables JSON output formatting and conversation continuation
3. Optionally executes bash command if `bash_cmd` is provided and includes output in stdin with XML tags
4. Reads file contents if `filepaths` are provided and includes them in stdin with XML tags
5. Adds `--mcp-disable=tools` flag to mods command if `readonly` is true
6. Executes the `mods` command with the prompt as the last argument
7. Parses stderr to extract conversation ID from "Conversation saved:" lines
8. Returns response wrapped in JSON object: `{"response": <output>, "conversation": <id>}`
9. Handles JSON parsing of the response when `json_output=true`

The server runs in stdio mode, making it suitable for MCP client integration.

