# ReleaseKit - Technical Requirements Document

## Architecture

ReleaseKit provides **two interfaces** to the same underlying functionality:

1. **Go library** — composable packages importable by downstream Go programs (e.g., `schangelog`, `release-agent-team`). All logic lives here.
2. **CLI** — a thin wrapper (`cmd/releasekit`) that calls library functions, formats output, and exits with appropriate codes. Used by agents, CI pipelines, and humans at the terminal.

The library is the primary interface. The CLI adds no logic beyond argument parsing, output formatting, and exit codes. Any capability available via CLI must also be available as a direct Go API call.

```
github.com/grokify/releasekit
├── run/           # Command execution abstraction (all external commands)
├── git/           # Git command execution and parsing
├── commits/       # Conventional commit analysis
├── checks/        # Validation check framework
├── detect/        # Language/project detection
├── workflow/      # Release workflow orchestration
├── output/        # Output formatting (TOON, JSON, text)
└── cmd/releasekit # CLI application (thin wrapper)
```

### Interface Design Principle

Every CLI command maps 1:1 to a library function. The CLI is purely:

```
parse flags → call library function → format result → exit
```

This ensures:

- **Library consumers** (schangelog, custom tooling) get full programmatic access with typed inputs/outputs
- **CLI consumers** (agents, CI, humans) get the same capabilities with TOON/JSON/text output
- **No feature drift** — the CLI cannot do anything the library cannot

### Library Usage Example (schangelog importing releasekit)

```go
import (
    "github.com/grokify/releasekit/git"
    "github.com/grokify/releasekit/commits"
    "github.com/grokify/releasekit/run"
)

func generateChangelog(dir, sinceTag string) (*Changelog, error) {
    runner := run.NewRunner()
    repo := git.NewRepo(dir, runner)

    // Get commits since last tag — returns typed structs, not CLI text
    infos, err := repo.CommitsSince(sinceTag)
    if err != nil {
        return nil, err
    }

    // Parse conventional commits
    for _, info := range infos {
        cc, err := commits.ParseConventional(info.Subject)
        if err != nil {
            continue
        }
        suggestion := commits.SuggestCategory(cc)
        // ... build changelog entries from typed data
    }
    return changelog, nil
}
```

### CLI Usage Example (agent or human)

```bash
# Same operation, but via CLI with TOON output for an LLM agent
$ releasekit commits --since=v1.2.0 --format=toon

# Or JSON for CI pipeline consumption
$ releasekit commits --since=v1.2.0 --format=json
```

Both paths execute the same underlying `git.Repo.CommitsSince()` + `commits.ParseConventional()` logic.

## Package Design

### `run/` - Command Execution

Shared command execution layer for all external commands (git, go, ls, gh, linters, etc.). All packages that run external processes use this layer rather than calling `os/exec` directly.

#### Key Types

```go
// Command represents any executable command for release analysis.
type Command struct {
    Name    string        // Human-readable label (e.g., "go build", "git status")
    Args    []string      // Full command and arguments (e.g., ["git", "status", "--porcelain"])
    Dir     string        // Working directory (empty = cwd)
    Env     []string      // Additional environment variables (KEY=VALUE)
    Timeout time.Duration // Per-command timeout (0 = use context deadline)
}

// Result captures the outcome of running a command.
type Result struct {
    Command  Command
    ExitCode int
    Stdout   string
    Stderr   string
    Duration time.Duration
    Err      error // Non-nil if command failed to start or timed out
}

// Passed returns true if the command exited with code 0.
func (r *Result) Passed() bool

// Output returns Stdout if non-empty, otherwise Stderr.
func (r *Result) Output() string

// Runner executes commands. Implementations support real execution, dry-run,
// and recording for audit/replay.
type Runner interface {
    Run(ctx context.Context, cmd Command) *Result
}

// RunSet represents a batch of commands run together (e.g., all checks).
type RunSet struct {
    Commands []Command
    Results  []*Result
}

// Passed returns true if all results passed.
func (rs *RunSet) Passed() bool

// Summary returns a pass/fail/skip breakdown.
func (rs *RunSet) Summary() Summary

type Summary struct {
    Total   int
    Passed  int
    Failed  int
    Skipped int
}
```

#### Runners

```go
// NewRunner returns a real command executor.
func NewRunner() Runner

// NewDryRunner returns a runner that records commands without executing them.
// Useful for previewing what a workflow or check suite would do.
func NewDryRunner() *DryRunner

type DryRunner struct {
    Recorded []Command
}

func (d *DryRunner) Run(ctx context.Context, cmd Command) *Result

// NewRecordingRunner wraps a real runner and records all commands + results.
// Useful for audit trails and LLM-optimized output.
func NewRecordingRunner(inner Runner) *RecordingRunner

type RecordingRunner struct {
    Inner   Runner
    History []RunRecord
}

type RunRecord struct {
    Command Command
    Result  *Result
}
```

#### Helpers

```go
// Git returns a Command configured for a git subcommand.
func Git(dir string, args ...string) Command

// Go returns a Command configured for a go subcommand.
func Go(dir string, args ...string) Command

// Shell returns a Command for an arbitrary shell command (ls, test, etc.).
func Shell(dir string, args ...string) Command

// RunAll executes a slice of commands sequentially, stopping on first failure
// if stopOnFail is true.
func RunAll(ctx context.Context, r Runner, cmds []Command, stopOnFail bool) *RunSet
```

#### Design Rationale

- **Unified execution** — git, go, ls, gh, and linter commands all flow through one layer
- **Dry-run for free** — any operation can preview what it would execute without side effects
- **Auditability** — RecordingRunner captures every command + result for LLM or CI consumption
- **Consistent timeouts** — per-command timeouts plus context-level cancellation
- **Composability** — workflows and check suites build on `Command` and `RunSet` values
- **Testability** — inject DryRunner in tests to verify commands without executing them

### `git/` - Git Operations

Core abstraction over git CLI commands with structured output. Uses `run.Runner` for all command execution.

#### Key Types

```go
// Repo represents a git repository context.
type Repo struct {
    Dir    string     // Working directory (empty = cwd)
    Runner run.Runner // Command executor (nil = default real runner)
}

// Tag represents a git tag with metadata.
type Tag struct {
    Name   string
    Commit string
    Date   time.Time
}

// CommitInfo represents a parsed git commit.
type CommitInfo struct {
    Hash      string
    ShortHash string
    Author    string
    Email     string
    Date      time.Time
    Subject   string
    Body      string
    Files     []FileStat
}

// FileStat represents file-level change statistics.
type FileStat struct {
    Path       string
    Insertions int
    Deletions  int
}

// Status represents working tree state.
type Status struct {
    Staged    []FileStatus
    Unstaged  []FileStatus
    Untracked []string
}

// FileStatus represents a file's git status.
type FileStatus struct {
    Path   string
    Status string // M, A, D, R, C
}

// RemoteInfo represents parsed remote URL.
type RemoteInfo struct {
    Host  string // github.com
    Owner string // grokify
    Repo  string // releasekit
    URL   string // https://github.com/grokify/releasekit
}

// CIStatus represents CI pipeline state.
type CIStatus struct {
    State      string // success, failure, pending
    Checks     []Check
    TotalCount int
}
```

#### Key Functions

```go
func NewRepo(dir string, runner run.Runner) *Repo

// Tags
func (r *Repo) Tags() ([]Tag, error)
func (r *Repo) LatestTag(prefix string) (*Tag, error)
func (r *Repo) CreateTag(name string) error
func (r *Repo) PushTag(name string) error

// Log
func (r *Repo) Log(opts LogOptions) ([]CommitInfo, error)
func (r *Repo) CommitsSince(tag string) ([]CommitInfo, error)
func (r *Repo) CommitsSincePath(tag, path string) ([]CommitInfo, error)

// Status
func (r *Repo) Status() (*Status, error)
func (r *Repo) IsClean() (bool, error)
func (r *Repo) BranchAheadBehind() (ahead, behind int, err error)

// Remote
func (r *Repo) RemoteURL(name string) (string, error)
func (r *Repo) ParseRemote(name string) (*RemoteInfo, error)
func ParseRemoteURL(rawURL string) (*RemoteInfo, error)

// CI (requires gh CLI)
func (r *Repo) CIStatus() (*CIStatus, error)
func (r *Repo) WaitForCI(timeout time.Duration, poll time.Duration) (*CIStatus, error)
```

#### LogOptions

```go
type LogOptions struct {
    Since    string // Start ref (tag or commit)
    Until    string // End ref (default: HEAD)
    Path     string // Filter by path
    Last     int    // Limit to N most recent
    NoMerges bool   // Exclude merge commits
    NumStat  bool   // Include file statistics
}
```

### `commits/` - Commit Analysis

Parses conventional commits and suggests changelog categories.

#### Key Types

```go
// ConventionalCommit represents a parsed conventional commit.
type ConventionalCommit struct {
    Type     string // feat, fix, docs, etc.
    Scope    string // optional scope
    Subject  string // description
    Breaking bool   // breaking change flag
    Issues   []int  // referenced issues
    PRs      []int  // referenced PRs
}

// CategorySuggestion represents a suggested changelog category.
type CategorySuggestion struct {
    Category   string  // Added, Fixed, Changed, etc.
    Tier       string  // core, standard, extended, optional
    Confidence float64 // 0.0 - 1.0
    Reasoning  string  // Why this category was chosen
}
```

#### Key Functions

```go
func ParseConventional(subject string) (*ConventionalCommit, error)
func SuggestCategory(commit *ConventionalCommit) *CategorySuggestion
func ExtractIssueRefs(text string) []int
func ExtractPRRefs(text string) []int
func IsBreakingChange(subject, body string) bool
```

### `checks/` - Validation Framework

Pluggable validation checks with consistent result reporting. All checkers use `run.Runner` for command execution, enabling dry-run previews and audit trails.

#### Key Types

```go
// Result represents a check outcome.
type Result struct {
    Name    string
    Status  Status // Pass, Fail, Warn, Skip
    Message string
    Detail  string      // Extended output (e.g., test failures)
    Cmd     run.Command // The command that was executed
}

// Status represents check outcome state.
type Status string

const (
    StatusPass Status = "pass"
    StatusFail Status = "fail"
    StatusWarn Status = "warn"
    StatusSkip Status = "skip"
)

// Checker runs validation checks for a language/area.
type Checker interface {
    Name() string
    Commands(dir string) []run.Command // Returns commands that would be run
    Check(ctx context.Context, r run.Runner, dir string) ([]Result, error)
}
```

The `Commands` method allows callers to inspect what a checker would execute before running it (e.g., for dry-run or user confirmation).

#### Built-in Checkers

```go
// Go checks
func NewGoChecker(opts GoCheckOptions) Checker

type GoCheckOptions struct {
    Build           bool
    Test            bool
    Lint            bool
    Format          bool
    ModTidy         bool
    Coverage        bool
    ExcludeCoverage string
}

// TypeScript checks
func NewTypeScriptChecker(opts TSCheckOptions) Checker

// File existence checks (ls-based)
func NewFileChecker(files ...string) Checker

// Release readiness checks (combines git, file, and language checks)
func NewReleaseChecker() Checker
```

#### Example: GoChecker Commands

When `NewGoChecker(GoCheckOptions{Build: true, Test: true, Lint: true, ModTidy: true})` is created, calling `Commands(dir)` returns:

```go
[]run.Command{
    run.Go(dir, "build", "./..."),
    run.Go(dir, "test", "./..."),
    run.Shell(dir, "golangci-lint", "run"),
    run.Go(dir, "mod", "tidy"),
}
```

### `detect/` - Language Detection

Detects project languages from manifest files.

```go
type Language string

const (
    LangGo         Language = "go"
    LangTypeScript Language = "typescript"
    LangJavaScript Language = "javascript"
    LangPython     Language = "python"
    LangRust       Language = "rust"
    LangSwift      Language = "swift"
)

// DetectedLanguage includes the language and its root path.
type DetectedLanguage struct {
    Language Language
    Path     string // Relative path to manifest
}

func Detect(dir string) ([]DetectedLanguage, error)
```

### `workflow/` - Release Orchestration

Step-based workflow engine for release processes. Each step defines its commands via `run.Command` values, and the workflow engine executes them through a `run.Runner`. This gives workflows automatic dry-run support and full audit trails.

```go
// Step represents a single step in a workflow.
type Step struct {
    Name     string
    Commands func(ctx context.Context) []run.Command // Commands this step will execute
    Run      func(ctx context.Context, r run.Runner) error
    Confirm  bool // Require user confirmation before executing
}

// Workflow orchestrates a sequence of steps.
type Workflow struct {
    Steps  []Step
    Runner run.Runner
}

func (w *Workflow) Execute(ctx context.Context, opts ExecuteOptions) error

// Preview returns all commands that would be executed without running them.
func (w *Workflow) Preview(ctx context.Context) []run.Command

type ExecuteOptions struct {
    DryRun      bool        // Use DryRunner instead of real runner
    Interactive bool        // Prompt for confirmation on Confirm steps
    Logger      *slog.Logger
}
```

### `output/` - Formatting

```go
type Format string

const (
    FormatTOON Format = "toon"
    FormatJSON Format = "json"
    FormatText Format = "text"
)

// Formatter writes structured data in the specified format.
type Formatter interface {
    Format(v any) ([]byte, error)
}

func NewFormatter(format Format) Formatter
```

## CLI Commands

Each CLI command is a thin wrapper: parse flags → call library → format output → exit. The "Library call" annotation shows the equivalent Go API.

### `releasekit status`

Combines tag inventory, commit delta, and working tree state.

**Library call:** `git.NewRepo(dir, runner).Status()`, `git.NewRepo(dir, runner).LatestTag(prefix)`, `git.NewRepo(dir, runner).CommitsSince(tag)`

```
Usage: releasekit status [flags]

Flags:
  --path string      Path prefix for nested module (e.g., sdk/go)
  --format string    Output format: toon, json, text (default: toon)
```

### `releasekit check`

Validates version consistency between tags and changelog.

**Library call:** `checks.NewReleaseChecker().Check(ctx, runner, dir)`

```
Usage: releasekit check [flags]

Flags:
  --changelog string   Path to CHANGELOG.json (default: CHANGELOG.json)
  --format string      Output format: toon, json, text (default: toon)
```

### `releasekit modules`

Shows status of all Go submodules in a repo.

**Library call:** `detect.Detect(dir)` + per-module `git.NewRepo(subdir, runner).CommitsSince(tag)`

```
Usage: releasekit modules [flags]

Flags:
  --format string    Output format: toon, json, text (default: toon)
```

### `releasekit validate`

Runs pre-push validation checks.

**Library call:** `detect.Detect(dir)` → `checks.NewGoChecker(opts).Check(ctx, runner, dir)`

```
Usage: releasekit validate [dir] [flags]

Flags:
  --lang string      Language to check (auto-detected if empty)
  --format string    Output format: toon, json, text (default: toon)
  --dry-run          Show commands that would be executed without running them
```

### `releasekit commits`

Parses commits in a range with conventional commit analysis.

**Library call:** `git.NewRepo(dir, runner).Log(opts)` → `commits.ParseConventional(subject)`

```
Usage: releasekit commits [flags]

Flags:
  --since string     Start ref (tag or commit)
  --until string     End ref (default: HEAD)
  --last int         Limit to N most recent
  --path string      Filter by path
  --no-merges        Exclude merge commits
  --format string    Output format: toon, json (default: toon)
```

### `releasekit release`

Executes the full release workflow.

**Library call:** `workflow.Workflow{...}.Execute(ctx, opts)`

```
Usage: releasekit release [flags]

Flags:
  --dry-run          Preview all commands without executing
  --format string    Output format: toon, json, text (default: toon)
```

## Module Dependencies

```
require (
    github.com/spf13/cobra v1.8+
    github.com/toon-format/toon-go v0.1+
)
```

No other external dependencies. All external command execution (git, go, linters, shell) flows through the `run/` package which uses `os/exec` internally.

## Error Handling

Follow the priority order:

1. **Return errors** - All library functions return errors to callers
2. **Wrap with context** - Use `fmt.Errorf("operation %s: %w", arg, err)`
3. **Structured logging** - CLI layer logs via `*slog.Logger`
4. **Exit codes** - CLI returns non-zero on failure

## Testing Strategy

- Unit tests for all parsing logic (commits, conventional, category)
- Integration tests for git operations (using `t.TempDir()` with real git repos)
- Table-driven tests for all parsers
- `run.DryRunner` for unit-testing command composition without executing commands
- Integration tests use real `run.Runner` with temporary repos
- `run.RecordingRunner` for verifying exact commands issued by checkers/workflows

## Migration Path

### Phase 1: Extract

1. Copy `gitlog/` from structured-changelog into `commits/`
2. Copy `pkg/git/` from release-agent-team into `git/`
3. Adapt interfaces, remove project-specific coupling
4. Add new `status`, `check`, `modules` commands

### Phase 2: Integrate

1. Update `structured-changelog` to import `releasekit/git` and `releasekit/commits`
2. Update `release-agent-team` to import `releasekit/git`, `releasekit/checks`, etc.
3. Remove duplicated code from both projects
4. Verify all existing tests pass

### Phase 3: Extend

1. Add `releasekit validate` with multi-language checks
2. Add `releasekit release` workflow command
3. Add changelog integration hooks
