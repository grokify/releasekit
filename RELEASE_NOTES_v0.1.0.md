# Release Notes: v0.1.0

**Release Date:** 2026-01-26

## Overview

Initial release of ReleaseKit, a release management toolkit for Go projects providing git operations, commit analysis, and release workflow automation.

## Highlights

- **Language Auto-Detection**: Automatically detects Go, TypeScript, JavaScript, Python, and Rust projects
- **Pre-Release Validation**: Pluggable checker framework for build, test, lint, and format checks
- **Multi-Agent Spec Output**: Results conform to AgentResult schema for agent workflow interoperability

## Features

### CLI Commands

| Command | Description |
|---------|-------------|
| `validate` | Run pre-release checks with language auto-detection |
| `check` | Validate version consistency between tags and CHANGELOG.json |
| `commits` | Parse and analyze git commits |
| `status` | Show git repository status |
| `modules` | List Go module dependencies |
| `version` | Print version information |

### Language Support

**Go Projects**:
- `go build ./...` - Build all packages
- `go test ./...` - Run tests
- `golangci-lint run` - Lint (if installed)
- `go mod tidy` - Verify module tidiness
- Coverage check (optional)

**TypeScript/JavaScript Projects**:
- Build via npm/yarn
- Test via npm/yarn
- Lint via npm/yarn

### Output Formats

All commands support multiple output formats via `--format`:

- `toon` (default) - Token-efficient format for LLM consumption
- `json` - Standard JSON output
- `text` - Human-readable text

### Execution Modes

- **Real execution** - Default mode, runs commands
- **Dry-run** - Preview commands without executing (`--dry-run`)
- **Recording** - Capture command execution for audit trails

## Installation

```bash
go install github.com/grokify/releasekit/cmd/releasekit@v0.1.0
```

## Dependencies

- `github.com/agentplexus/multi-agent-spec/sdk/go` v0.5.0 - AgentResult schema
- `github.com/spf13/cobra` v1.10.2 - CLI framework
- `github.com/toon-format/toon-go` - TOON output format
