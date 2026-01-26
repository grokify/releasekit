# ReleaseKit - Product Requirements Document

## Overview

ReleaseKit is a Go toolkit and CLI for accelerating release management of individual repositories. It consolidates git operations, commit analysis, release readiness checks, and workflow orchestration into a single, composable library with LLM-optimized (TOON) output.

## Problem Statement

Release preparation for Go (and multi-language) projects involves repetitive, error-prone manual steps:

1. **Scattered tooling** - Git operations, commit parsing, validation checks, and changelog generation are implemented independently across projects, leading to duplication and inconsistency.
2. **Slow context gathering** - Understanding "what changed since the last release" requires multiple git commands, mental cross-referencing of tags vs changelogs, and manual verification.
3. **No single source of truth** - Projects like `structured-changelog` and `release-agent-team` each maintain their own git wrappers, conventional commit parsers, and output formatters.
4. **LLM inefficiency** - Raw git output is verbose; AI-assisted release workflows waste tokens on noise.

## Target Users

1. **Developers** preparing releases for Go projects (primary)
2. **AI agents** (Claude Code, release-agent-team) automating release workflows
3. **CI/CD pipelines** validating release readiness
4. **Multi-repo maintainers** needing consistent release tooling across projects

## Key Features

### P0 - Core (MVP)

#### Git Operations Library (`git/`)

- Execute git commands with structured error handling
- Tag management: list, create, verify, delete
- Commit log with range queries (`since..until`, `--last N`, `--path`)
- Working tree status (staged, unstaged, untracked)
- Remote URL parsing (SSH/HTTPS normalization)
- Branch tracking state (ahead/behind)

#### Commit Analysis (`commits/`)

- Parse raw git log output into structured commits
- Conventional commit parsing (type, scope, subject, breaking)
- Category suggestion with confidence scoring
- Issue/PR reference extraction

#### Release Status CLI (`releasekit status`)

- Show latest tag, commits since, uncommitted changes
- Nested Go module awareness (tag prefix filtering)
- TOON output by default for LLM consumption

#### Version Consistency Check (`releasekit check`)

- Compare tag list against CHANGELOG.json versions
- Detect mismatches (changelog describes unreleased content as released)
- Validate semantic version format

### P1 - Validation

#### Pre-Push Checks (`checks/`)

- Go: build, test, lint, format, mod tidy, no local replace
- TypeScript: eslint, prettier, tsc, test
- Pluggable Checker interface for custom checks

#### Language Detection (`detect/`)

- Auto-detect project languages from manifest files
- Support monorepos with multiple languages

#### CI Status (`git/ci.go`)

- GitHub Actions status via `gh` CLI
- Polling with configurable timeout
- Pass/fail/pending state

### P2 - Workflow

#### Release Workflow (`workflow/`)

- Step-based orchestration engine
- Standard 9-step release workflow
- Dry-run mode
- Interactive approval for mutating steps

#### Nested Module Support (`releasekit modules`)

- Detect Go submodules in a repo
- Show per-module tag and delta status
- Coordinate multi-module releases

### P3 - Integration

#### Output Formats (`output/`)

- TOON (default) - token-optimized for LLMs
- JSON - structured for programmatic use
- Text - human-readable CLI output

#### Changelog Integration

- Interface with `schangelog` for changelog generation
- Validate changelog entries against actual commits

## Non-Goals

- **Multi-repo orchestration** - That's `releasehub`
- **Changelog IR/rendering** - That stays in `structured-changelog`
- **Agent definitions/specs** - That stays in `release-agent-team`
- **Git hosting provider APIs** (GitHub releases, etc.) - Future consideration

## Success Metrics

1. Both `structured-changelog` and `release-agent-team` import releasekit for git operations (zero duplication)
2. `releasekit status` replaces 5+ manual git commands in a single invocation
3. TOON output is ~8x more token-efficient than raw git output for LLM consumers
4. All existing tests in downstream projects continue to pass after migration

## Dependencies

- Go 1.21+
- `git` CLI
- `gh` CLI (optional, for CI status)
- `github.com/spf13/cobra` (CLI framework)
- `github.com/toon-format/toon-go` (TOON output)

## Relationship to Other Projects

```
releasekit (this project)
    ^               ^
    |               |
structured-     release-agent-
changelog       team
    ^               ^
    |               |
releasehub (orchestrates across repos)
```

- **releasekit** provides composable git/release primitives
- **structured-changelog** uses releasekit for git ops, adds changelog IR/rendering
- **release-agent-team** uses releasekit for git ops, adds agent orchestration
- **releasehub** orchestrates releasekit across 100s of repos on a schedule
