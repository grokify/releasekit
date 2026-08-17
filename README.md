# ReleaseKit

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/releasekit/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/releasekit/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/releasekit/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/releasekit/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/releasekit/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/releasekit/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/releasekit
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/releasekit
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://grokify.github.io/releasekit
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Freleasekit
 [loc-svg]: https://tokei.rs/b1/github/grokify/releasekit
 [repo-url]: https://github.com/grokify/releasekit
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/releasekit/blob/main/LICENSE

Release management toolkit for Go projects. Provides git operations, commit analysis, and release workflow automation.


## Features

- **Language Auto-Detection**: Automatically detects Go, TypeScript, JavaScript, Python, and Rust projects
- **Pre-Release Validation**: Runs build, test, lint, and format checks
- **Git Operations**: Tag management, branch analysis, commit parsing
- **Conventional Commits**: Parse and categorize commits for changelogs
- **Multi-Agent Spec Output**: Results conform to the AgentResult schema for agent workflow interoperability
- **Multiple Output Formats**: TOON (token-efficient), JSON, and text formats

## Installation

```bash
go install github.com/grokify/releasekit/cmd/releasekit@latest
```

## Commands

| Command | Description |
|---------|-------------|
| `validate` | Run pre-release checks (auto-detects language) |
| `check` | Validate version consistency between tags and CHANGELOG.json |
| `commits` | Parse and analyze git commits |
| `status` | Show git repository status |
| `modules` | List Go module dependencies |
| `version` | Print version information |

## Usage

### Validate a Project

Run pre-release validation checks:

```bash
# Validate current directory
releasekit validate

# Validate specific directory
releasekit validate /path/to/project

# Dry run (show what would be executed)
releasekit validate --dry-run

# Skip certain checks
releasekit validate --no-lint --no-test

# Include coverage check
releasekit validate --coverage
```

### Check Version Consistency

Validate that git tags and CHANGELOG.json are in sync:

```bash
releasekit check
releasekit check --changelog CHANGELOG.json
```

### Analyze Commits

Parse commits using conventional commit format:

```bash
releasekit commits
releasekit commits --since v1.0.0
```

### Output Formats

All commands support multiple output formats:

```bash
releasekit validate --format json
releasekit validate --format toon    # Default, token-efficient
releasekit validate --format text
```

## Validation Checks

The `validate` command runs language-specific checks:

### Go Projects

- `go build ./...` - Build all packages
- `go test ./...` - Run all tests
- `golangci-lint run` - Run linter (if installed)
- `go mod tidy` - Verify module tidiness
- Coverage check (optional)

### TypeScript/JavaScript Projects

- `npm run build` or `yarn build` - Build project
- `npm test` or `yarn test` - Run tests
- `npm run lint` or `yarn lint` - Run linter

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed (GO) |
| 1 | Error running checks |
| 2 | One or more checks failed (NO-GO) |

## Dependencies

- [multi-agent-spec](https://github.com/agentplexus/multi-agent-spec) - AgentResult schema for output
- [toon-go](https://github.com/toon-format/toon-go) - TOON format support
- [cobra](https://github.com/spf13/cobra) - CLI framework

## License

MIT License. See [LICENSE](LICENSE) for details.
