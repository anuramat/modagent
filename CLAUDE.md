# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Instructions

You MUST NOT commit binaries to the repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server called "modagent" that provides a bridge to the `mods` CLI tool. The server exposes a single tool called "subagent" that executes LLM prompts through the `mods` command and returns the results.

## Architecture

- **Single file Go application** (`main.go`): Contains the complete MCP server implementation
- **Core functionality**: Wraps the `mods` CLI tool to provide LLM capabilities via MCP protocol
- **Tool interface**: Exposes one tool `subagent` with parameters:
  - `prompt` (required): The prompt to send to the LLM (passed as last argument to `mods`)
  - `json_output` (optional): Whether to parse and return structured JSON output
  - `conversation` (optional): Conversation ID to continue from a previous conversation
  - `filepath` (optional): Absolute path to a file to pass as stdin to `mods`

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
3. Reads file content if `filepath` is provided and uses it as stdin
4. Executes the `mods` command with the prompt as the last argument
5. Parses stderr to extract conversation ID from "Conversation saved:" lines
6. Returns response wrapped in JSON object: `{"response": <output>, "conversation": <id>}`
7. Handles JSON parsing of the response when `json_output=true`

The server runs in stdio mode, making it suitable for MCP client integration.

