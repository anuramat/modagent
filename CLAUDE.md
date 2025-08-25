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
- **Core package** (`core/`): Shared server implementation and business logic
  - `server.go`: `BaseServer` with `HandleCall`/`HandleCallReadonly` methods, argument parsing, mods command building, stdin preparation, and response formatting
  - `description.go`: Utility for embedding descriptions
- **Config package** (`config/`): XDG-compliant configuration system
  - `config.go`: Config loading, validation, and generation with tool description overrides
- **Junior package** (`junior/`): General AI assistance tool implementation
  - `server.go`: Lightweight wrapper around `core.BaseServer` with junior-specific role configuration
  - `description.go`: Embeds the tool description from `description.md`
  - `description.md`: Markdown documentation for the junior tool
- **Logworm package** (`logworm/`): Specialized tool for command output analysis
  - `server.go`: Wraps `core.BaseServer` with custom `HandleCall` that transforms logworm requests into core requests with logworm role
  - `description.go`: Embeds tool description
  - `description.md`: Markdown documentation for logworm
- **Testing utilities** (`testutils/`): Shared test helpers and mocks for comprehensive test coverage
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

## Configuration

The server supports XDG-compliant configuration for customizing tool descriptions:

**Config Location**: `$XDG_CONFIG_HOME/modagent/config.yaml` (fallback: `~/.config/modagent/config.yaml`)

**Generate Default Config**:
```bash
./modagent -generate-config
```

**Config Format**:
```yaml
tools:
  junior-r:
    description:
      text: "Custom description for junior-r tool"
  junior-rwx:
    description:
      path: "descriptions/junior-rwx.md"
  logworm:
    description:
      text: ""  # Empty description
```

**Features**:
- Two ways to specify descriptions: `text` (inline) or `path` (file reference)
- `text` and `path` are mutually exclusive per tool
- Relative paths resolved relative to config directory
- Validates tool names on startup (fails if unknown tools specified)
- Validates file paths exist when using `path` option
- Backwards compatible (works without config file)
- Generated config contains actual default descriptions using `text`
- Empty strings in `text` are treated as valid empty descriptions

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
# Run all tests
go test ./...

# Check if flake builds successfully
nix flake check

# Verify the binary works
echo "test prompt" | ./modagent
```

## Dependencies

- **MCP Framework**: Uses `github.com/mark3labs/mcp-go` for MCP protocol implementation
- **XDG Support**: Uses `github.com/adrg/xdg` for cross-platform config directory handling
- **External requirement**: Requires `mods` CLI tool to be available in PATH
- **Go version**: 1.23+

## Key Implementation Details

The server is organized into the following components:

### Main (`main.go`)

- Handles command-line flags (`-generate-config`, `-logworm-only`)
- Loads and validates configuration on startup
- Creates MCP server instance with version "unstable"
- Instantiates junior and logworm servers using their respective constructors
- Registers tools with configurable descriptions:
  - `junior-r` and `junior-rwx` (unless `-logworm-only` flag is used)
  - `logworm` (always enabled)
- Serves via stdio for MCP client integration

### Core Server (`core/server.go`)

The `BaseServer` provides shared functionality for all tools:

- **`HandleCall`/`HandleCallReadonly`**: Main entry points that delegate to internal handler with readonly flag
- **Request Processing Pipeline**:
  1. **Parse arguments**: Extract and validate parameters using `ParseArgs`
  2. **Build mods command**: Construct command with appropriate flags via `buildModsCmd`:
     - `-j` for JSON output
     - `--continue=<id>` for conversation continuation 
     - `-R <role>` using tool-specific `GetDefaultRole` or explicit role parameter
  3. **Prepare stdin**: Execute bash commands, read files, wrap outputs in XML tags via `prepareStdin`
  4. **Execute mods**: Run the command and capture stdout/stderr via `runCommand`
  5. **Extract conversation ID**: Parse stderr for "Conversation saved:" pattern
  6. **Build response**: Format JSON response with response, conversation ID, and temp directory info
  7. **Handle JSON parsing**: Parse LLM response as JSON when `json_output=true`

- **Temporary Directory Management**: Creates timestamped temp directories for bash command outputs, saves stdout/stderr files
- **ServerConfig Interface**: Allows different tools to provide their own role configuration via `GetDefaultRole` method

### Config System (`config/config.go`)

- **XDG Directory Support**: Uses standard config directories across platforms
- **Config Loading**: Automatically loads config on startup, gracefully handles missing files
- **Validation**: Prevents startup if config contains unknown tool names
- **Generation**: `-generate-config` flag creates default config with actual default descriptions
- **Override Logic**: Uses config descriptions when present, falls back to embedded defaults

### Junior Server (`junior/server.go`)

- **Lightweight Wrapper**: Embeds `core.BaseServer` and implements `core.ServerConfig` interface
- **Role Configuration**: `GetDefaultRole` returns `"junior-r"` for readonly mode, `"junior-rwx"` for full access
- **Inheritance**: Inherits all functionality from `BaseServer` with no additional request processing

### Logworm Server (`logworm/server.go`)

- **Specialized Handler**: Custom `HandleCall` method that transforms logworm requests into core requests
- **Request Transformation**: Takes `bash_cmd` parameter and creates internal request with:
  - Fixed prompt: "Parse and analyze this command output"
  - Original bash command
  - Fixed role: "logworm"
- **Role Configuration**: `GetDefaultRole` always returns `"logworm"` regardless of readonly flag
