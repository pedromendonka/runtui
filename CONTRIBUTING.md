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
| `make lint` | Run vet + staticcheck (if installed) |
| `make fmt` | Format all Go files |
| `make clean` | Remove build artifacts |
| `make setup` | Install git hooks (required once after cloning) |
| `make release-dry` | Dry-run GoReleaser locally |

## Architecture

```
runtui/
  main.go                 Entry point, flag parsing, orchestration
  detector/
    detector.go            Scan cwd for config files + detect package manager from lockfiles
    detector_test.go
  parser/
    parser.go              Parser interface, Task and ArgDef types
    packagejson.go         Parse package.json scripts + runtui config block
    packagejson_test.go
  runner/
    runner.go              Build exec.Cmd from runner, subcmd, script, and args
    runner_test.go
  tui/
    model.go               Bubble Tea model, phase state machine, execution lifecycle
    update.go              Key event handling for list and args phases
    view.go                Render banner, task table, args form, exec results
    styles.go              Lip Gloss color palette and style definitions
  testdata/
    package.json           Sample project fixture with runtui config
```

### Data flow

```
main.go
  ├── detector.Detect()     → []Project (type, path, runner)
  ├── parser.Parse()        → []Task (name, command, description, args)
  └── tui.New() + tea.Run() → interactive loop
        ├── phaseList        → browse/filter tasks
        ├── phaseArgs        → collect arguments (simple or config-driven)
        └── tea.ExecProcess  → run command, resume TUI on completion
```

### Key design decisions

- **Bubble Tea Elm architecture** -- model is a value type, `Update` returns a new model, `View` is a pure function of state.
- **Interfaces at the package level** -- `parser.Parser` defines the contract; each project type (package.json, future Makefile, etc.) implements it.
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

1. **Create a parser** -- add `parser/newtype.go` implementing `parser.Parser`. The `Parse(path) ([]Task, error)` method reads the config file and returns tasks.
2. **Register in detector** -- add the config filename to `configFiles` in `detector/detector.go`.
3. **Wire in main.go** -- add a `case` to the project type switch.
4. **Add tests** -- add `parser/newtype_test.go` and a test fixture in `testdata/`.

Example for Makefile support:

```go
// parser/makefile.go
type Makefile struct{}

func (p *Makefile) Parse(path string) ([]Task, error) {
    // Parse make targets from Makefile
}
```

The runner subcmd for make would be `""` (targets are passed directly: `make build`).

## Code style

- **Go 1.26 idioms** -- use `errors.AsType[T]()` for type-safe error assertions, `slices.SortFunc` for sorting, `strings.Builder` for string construction.
- **No unnecessary abstractions** -- don't add helpers, wrappers, or interfaces until there's a second consumer.
- **Error wrapping** -- use `fmt.Errorf("context: %w", err)` to wrap errors with context.
- **No hardcoded strings** -- runner names are validated against `validRunners` in `main.go`.

## Testing

```bash
make test
```

Tests cover:
- `parser/` -- script parsing, runtui config, empty/malformed JSON, unknown script configs
- `runner/` -- command construction with/without args, with/without subcmd, different runners
- `detector/` -- project detection, lockfile-based runner detection, priority ordering

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
