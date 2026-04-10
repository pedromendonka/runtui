# Contributing to runtui

Thanks for your interest in contributing! This guide covers everything you need to get started.

## Prerequisites

- Go 1.26+
- A terminal with true-color support (for the full visual experience)

### Optional dev tools

These are not required but improve the development experience:

```bash
go install gotest.tools/gotestsum@latest              # colored, grouped test output (used by make test)
go install honnef.co/go/tools/cmd/staticcheck@latest   # deeper static analysis (used by make lint)
```

`make test` and `make lint` fall back to standard Go tools if these are not installed.

## Getting started

```bash
git clone https://github.com/pedromendonka/runtui.git
cd runtui
make setup    # install git hooks (required once)
make build
```

Run against the included test fixture:

```bash
make run
```

## Git hooks

After cloning, run `make setup` to activate the project's git hooks. This is a one-time step.

| Hook | When it runs | What it checks |
|------|-------------|----------------|
| **pre-commit** | Every `git commit` | `gofmt` format check, `go vet`, `go build` |
| **pre-push** | Every `git push` | Full test suite (`go test ./...`) |

The hooks live in `.githooks/` and are tracked in the repo. `make setup` tells git to use them via `core.hooksPath`.

If a hook fails, the commit or push is blocked. Fix the issue and retry. The pre-commit hook will tell you what to run (e.g. `make fmt`).

## Make targets

Run `make help` to see all available commands:

| Target | Description |
|--------|-------------|
| `make build` | Build binary to `bin/runtui` |
| `make install` | Install to `$GOPATH/bin` |
| `make run` | Build and run against `testdata/` |
| `make run-info` | Build and run with `--info` flag |
| `make test` | Run all tests |
| `make vet` | Run `go vet` |
| `make lint` | Run vet + format check + staticcheck + tidy check |
| `make fmt` | Format all Go files |
| `make fmt-check` | Fail if any Go files are not formatted |
| `make tidy` | Run `go mod tidy` and verify integrity |
| `make tidy-check` | Fail if `go.mod`/`go.sum` would change after tidy |
| `make clean` | Remove build artifacts |
| `make setup` | Install git hooks (required once after cloning) |
| `make release-dry` | Dry-run GoReleaser locally |

## Architecture

```
runtui/
  main.go                 Thin entry point, delegates to app.Run
  app/
    app.go                 Orchestration: flag parsing, detection, parser selection, TUI boot
    app_test.go            End-to-end behavior tests
    registry.go            Parser factory registry (maps project types to parsers)
  detector/
    detector.go            Scan cwd for config files + detect package manager from lockfiles
    detector_test.go
  parser/
    parser.go              Parser interface, Task, ArgDef, and RunContext types
    packagejson.go         Parse package.json scripts + runtui config block
    packagejson_test.go
    makefile.go            Parse Makefile targets + ## descriptions
    makefile_test.go
  runner/
    runner.go              Build exec.Cmd from RunContext, script, and args
    runner_test.go
  tui/
    model.go               Bubble Tea model, phase state machine, execution lifecycle
    update.go              Key event handling for list and args phases
    view.go                Render banner, task table, args form, exec results
    styles.go              Lip Gloss color palette and style definitions
  testdata/
    package.json           Sample package.json fixture with runtui config
    Makefile               Sample Makefile fixture with ## descriptions
```

### Data flow

```
main.go
  └── app.Run()               orchestration, flag parsing
        ├── detector.Detect()     → []Project (type, path, runner)
        ├── parserFor()           → parser.Parser (via registry)
        ├── parser.Parse()        → []Task, RunContext
        └── tui.New() + tea.Run() → interactive loop
              ├── phaseList        → browse/filter tasks
              ├── phaseArgs        → collect arguments (simple or config-driven)
              └── tea.ExecProcess  → run command, resume TUI on completion
```

### Key design decisions

- **Bubble Tea Elm architecture** -- model is a value type, `Update` returns a new model, `View` is a pure function of state.
- **Interfaces at the package level** -- `parser.Parser` defines the contract; each project type (package.json, Makefile, etc.) implements it.
- **`runner.BuildCmd` takes a subcmd parameter** -- `"run"` for npm/yarn/pnpm/bun, `""` for make, making it extensible to future project types.
- **`tea.ExecProcess` for terminal passthrough** -- the TUI suspends, gives the subprocess full stdin/stdout/stderr, then resumes.
- **`--runner` validated against an allow-list** -- prevents arbitrary command execution.

### TUI phases

The TUI model has two phases:

1. **`phaseList`** -- task table with column headers, arrow navigation, substring filter, execution result display.
2. **`phaseArgs`** -- argument collection. Two modes:
   - *Simple*: single free-form input (tasks without `runtui.args` config)
   - *Config-driven*: structured fields with labels, required markers, hints (tasks with `runtui.args`)

### Color palette

Purple-to-teal gradient (`#7C3AED` -> `#6366F1` -> `#14B8A6`) with slate grays for muted/dim text. Defined in `tui/styles.go`. All colors use hex values for true-color terminals.

## Adding a new project type

1. **Create a parser** -- add `parser/newtype.go` implementing `parser.Parser`. The `Parse(path string) ([]Task, RunContext, error)` method reads the config file and returns tasks plus a `RunContext` describing how to invoke them.
2. **Register in detector** -- add the config filename and a `ProjectType` constant to `detector/detector.go`.
3. **Wire in registry** -- add a `parserFactory` entry in `app/registry.go` mapping the new `ProjectType` to the parser constructor.
4. **Add tests** -- add `parser/newtype_test.go` and a test fixture in `testdata/`.

For a real example, see how Makefile support was added: `parser/makefile.go`, `detector/detector.go` (`TypeMakefile`), and `app/registry.go`.

Example for a hypothetical Cargo.toml parser:

```go
// parser/cargo.go
type CargoParser struct{}

func (p *CargoParser) Parse(path string) ([]Task, RunContext, error) {
    // Parse [scripts] or known cargo commands from Cargo.toml
    // RunContext: Binary="cargo", Subcmd="", ArgSeparator="--"
}
```

```go
// detector/detector.go — add the project type
const TypeCargoToml ProjectType = "Cargo.toml"

// configFiles — add the entry with default runner
{TypeCargoToml, "Cargo.toml", "cargo"},
```

```go
// app/registry.go — register the parser
detector.TypeCargoToml: func(_ string) parser.Parser { return &parser.CargoParser{} },
```

## Code style

- **Go 1.26 idioms** -- use `errors.AsType[T]()` for type-safe error assertions, `slices.SortFunc` for sorting, `strings.Builder` for string construction.
- **No unnecessary abstractions** -- don't add helpers, wrappers, or interfaces until there's a second consumer.
- **Error wrapping** -- use `fmt.Errorf("context: %w", err)` to wrap errors with context.
- **No hardcoded strings** -- runner names are validated against `validRunners` in `app/app.go`.

## Testing

```bash
make test
```

Tests cover:
- `app/` -- flag parsing, project selection, `--type` flag, registry wiring, error paths
- `parser/` -- package.json scripts + runtui config, Makefile targets + `##` descriptions, edge cases
- `runner/` -- command construction with/without args, with/without subcmd, different runners
- `detector/` -- project detection, lockfile-based runner detection, multi-project, priority ordering
- `tui/` -- navigation, filtering, argument collection (simple + config-driven), state transitions

## Releasing

Releases are automated via GoReleaser + GitHub Actions:

```bash
git tag v0.1.0
git push --tags
```

This triggers `.github/workflows/release.yml`, which builds binaries for:
- macOS (Intel + Apple Silicon)
- Linux (amd64 + arm64)
- Windows (amd64)

Version is injected via `-ldflags -X main.version={{.Version}}`.

To dry-run locally:

```bash
make release-dry
```

## Commit format

```
BRANCH-ID | Feature Description | Phase Description
```

Extract the branch ID from the current git branch name.
