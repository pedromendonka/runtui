# RUNTUI — Project Plan

> **One TUI to run them all.**
> Auto-detects your project type. Pick a task, add args, run.

---

## What is RUNTUI?

A Go-based terminal user interface (TUI) that reads task definitions from project config files (package.json, Makefile, Cargo.toml, etc.), presents them in an interactive list, and lets you select, configure arguments, and run them — all without memorizing command names.

Built with **Bubble Tea** (TUI framework) and distributed as a single binary.

---

## Core Features (MVP)

### 1. Auto-detect project type
- Scan current directory for known config files
- If multiple found (e.g. `package.json` + `Makefile`), let user pick which to use

### 2. Read & list tasks
- Parse scripts/tasks from the detected config file
- Display in a navigable, filterable list
- Show the actual command next to each task name

### 3. Argument prompting (key differentiator)
- After selecting a task, prompt for arguments before execution
- Two modes working together:
  - **Simple mode (default):** single "Arguments (enter to skip)" prompt for any task
  - **Config-driven mode:** structured prompts per task via project config

#### Config-driven args example (package.json):
```json
{
  "runtui": {
    "descriptions": {
      "deploy": "Deploy the application",
      "test": "Run test suite"
    },
    "args": {
      "env:set": [
        { "name": "KEY", "required": true, "hint": "e.g. DATABASE_URL" },
        { "name": "VALUE", "required": true }
      ],
      "env:filter": [
        { "name": "PARTIAL_KEY", "required": true, "hint": "e.g. STRIPE" }
      ]
    }
  }
}
```

### 4. Fuzzy search / filter
- Type to filter tasks as the list is displayed
- Essential for projects with many scripts

### 5. Rerun last command
- Cache the last executed task + args per project
- Shorthand command to instantly repeat it

### 6. Multi-runner support
- Support npm, yarn, pnpm, bun (for Node projects)
- Auto-detect which is in use (lockfile detection) or allow manual config

---

## Supported Project Types

| Status | Config File | Task Source | Runner |
|--------|------------|-------------|--------|
| Done | `package.json` | `scripts` object | npm / yarn / pnpm / bun |
| Done | `Makefile` | make targets | make |
| Planned | `Cargo.toml` | cargo commands | cargo |
| Planned | `pyproject.toml` | poetry/pdm scripts | poetry / pdm |
| Planned | `docker-compose.yml` | services | docker compose |
| Future | `Taskfile.yml` | tasks | task |
| Future | `justfile` | recipes | just |

---

## Architecture

### Tech Stack
- **Language:** Go
- **TUI framework:** Bubble Tea (charmbracelet/bubbletea)
- **Styling:** Lip Gloss (charmbracelet/lipgloss)
- **CLI flags:** Go standard `flag` package (no Cobra needed for now)

### Project Structure
```
runtui/
├── main.go                 # Thin entry point, delegates to app.Run
├── go.mod
├── go.sum
├── app/
│   ├── app.go              # Orchestration: flags, detection, parser selection, TUI boot
│   ├── app_test.go
│   └── registry.go         # Parser factory registry (maps project types to parsers)
├── detector/
│   ├── detector.go         # Auto-detect project type + package manager from lockfiles
│   └── detector_test.go
├── parser/
│   ├── parser.go           # Parser interface, Task, ArgDef, RunContext types
│   ├── packagejson.go      # Parse package.json scripts + runtui config
│   ├── packagejson_test.go
│   ├── makefile.go         # Parse Makefile targets + ## descriptions
│   └── makefile_test.go
├── runner/
│   ├── runner.go           # Build exec.Cmd from RunContext, script, and args
│   └── runner_test.go
├── tui/
│   ├── model.go            # Bubble Tea model (state)
│   ├── update.go           # Bubble Tea update (handle events)
│   ├── update_test.go
│   ├── view.go             # Bubble Tea view (render UI)
│   ├── model_test.go
│   └── styles.go           # Lip Gloss styles
├── testdata/
│   ├── package.json        # Sample package.json fixture with runtui config
│   └── Makefile            # Sample Makefile fixture with ## descriptions
├── .goreleaser.yml
├── .github/
│   └── workflows/
│       ├── ci.yml          # CI: test, lint, vet across OS matrix
│       └── release.yml     # GoReleaser on version tags
├── PLAN.md                 # This file
├── README.md
├── CONTRIBUTING.md
└── LICENSE
```

### Key Interfaces
```go
// Every project type implements this
type Parser interface {
    Parse(path string) ([]Task, RunContext, error)
}

type Task struct {
    Name        string
    Command     string
    Description string
    Args        []ArgDef  // nil = use simple prompt
}

type ArgDef struct {
    Name     string
    Required bool
    Hint     string
    Default  string
}

// RunContext describes how to invoke a task (so TUI/runner don't need
// to know whether the project is npm, make, cargo, etc.)
type RunContext struct {
    Binary        string   // e.g. "npm", "make"
    Subcmd        string   // e.g. "run" for npm, "" for make
    ArgSeparator  string   // e.g. "--" for npm, "" for make
    DisplayPrefix []string // e.g. ["npm", "run"] or ["make"]
}
```

---

## CLI Usage

```bash
runtui                      # Launch TUI in current dir
runtui --info               # Show full commands next to task names
runtui --runner=pnpm        # Override detected runner
runtui rerun                # Rerun last executed task
runtui --type=makefile      # Force project type detection
```

---

## Distribution Plan (in priority order)

### Step 1: GitHub Repository
- Create `github.com/pedromendonka/runtui`
- Module path in `go.mod` must match: `module github.com/pedromendonka/runtui`
- `main.go` at repo root so `go install` works

### Step 2: GoReleaser + GitHub Actions
This is the linchpin — powers all other distribution methods.

- Add `.goreleaser.yml` to project root
- Builds binaries for: macOS (Intel + Apple Silicon), Linux (amd64 + arm64), Windows
- GitHub Action triggers on new tags (`v*`)
- Creates GitHub Release with all binaries attached

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Step 3: go install
Once repo is public, this works immediately:
```bash
go install github.com/pedromendonka/runtui@latest
```
No extra setup needed.

### Step 4: Homebrew Tap
- Create a separate repo: `github.com/pedromendonka/homebrew-runtui`
- Add `brews` section to `.goreleaser.yml`:

```yaml
brews:
  - name: runtui
    homepage: https://github.com/pedromendonka/runtui
    description: "Interactive TUI for running project tasks"
    license: "MIT"
    repository:
      owner: pedromendonka
      name: homebrew-runtui
```

GoReleaser auto-updates the formula on each release. Users install with:
```bash
brew tap pedromendonka/runtui
brew install runtui
```

### Step 5: npm Wrapper (optional)
Thin npm package that downloads the correct Go binary on `postinstall`.

```json
{
  "name": "runtui",
  "version": "1.0.0",
  "bin": { "runtui": "./bin/runtui" },
  "scripts": {
    "postinstall": "node install.js"
  }
}
```

`install.js` detects OS/arch → downloads matching binary from GitHub Release → places in `./bin/`. Libraries like `binary-install` handle this pattern. This is how esbuild and turbo distribute.

Publish to npm under your existing account.

---

## TUI Flow

```
┌─────────────────────────────────────────┐
│  runtui — package.json (npm)            │
│                                         │
│  Filter: _                              │
│                                         │
│  ❯ dev           dotenvx run -- ...     │
│    build         tsc && node bui...     │
│    test          jest                   │
│    env:set       dotenvx set            │
│    env:filter    node scripts/fi...     │
│    deploy        ./deploy.sh            │
│                                         │
│  ↑/↓ navigate · type to filter          │
│  enter to run · q to quit               │
└─────────────────────────────────────────┘

→ User selects: env:set

┌─────────────────────────────────────────┐
│  runtui — env:set (dotenvx set)         │
│                                         │
│  KEY:   DATABASE_URL_                   │
│         hint: e.g. DATABASE_URL         │
│                                         │
│  VALUE: postgres://localhost:5432/mydb_  │
│                                         │
│  ▸ Run   ○ Cancel                       │
└─────────────────────────────────────────┘

→ Executing: npm run env:set DATABASE_URL postgres://localhost:5432/mydb
```

---

## Existing Tools & How RUNTUI Differs

| Tool | Language | What it does | RUNTUI advantage |
|------|----------|-------------|------------------|
| **ntl** | Node.js | Interactive npm script picker | No arg prompting, Node-only, appears dormant |
| **lazynpm** | Go | Full npm TUI (deps, packages, scripts) | No arg prompting, npm-only, less maintained |
| **nrun** | Node.js | Script runner with direct args | Not a TUI, no interactive selection |

**RUNTUI's differentiators:**
1. Argument prompting — structured per-task arg definitions
2. Multi-ecosystem — package.json, Makefile, Cargo.toml, etc.
3. Single Go binary — no runtime dependencies
4. Homebrew + npm + go install distribution

---

## Development Milestones

### v0.1.0 — MVP (done)
- [x] package.json parser (scripts + descriptions)
- [x] Bubble Tea list with arrow navigation
- [x] Fuzzy search/filter
- [x] Simple "Arguments (enter to skip)" prompt
- [x] Execute selected task via os/exec
- [x] GoReleaser + GitHub Actions setup

### v0.2.0 — Args & Config (done)
- [x] Config-driven argument definitions (`runtui.args` in package.json)
- [x] Structured multi-field argument prompts
- [ ] Rerun cache + `runtui rerun` shorthand
- [x] `--info` flag to show full commands

### v0.3.0 — Makefile Support (done)
- [x] Makefile target parser (with `##` description extraction)
- [x] Auto-detection (package.json vs Makefile)
- [x] `--type` flag for explicit project selection
- [x] Homebrew tap config (GoReleaser brews section)

### v0.4.0 — Distribution
- [ ] Homebrew tap repo (`homebrew-tap`) + first tagged release
- [ ] npm wrapper package
- [ ] README with demo GIF

### v0.5.0 — Ecosystem Expansion
- [ ] Cargo.toml support
- [ ] pyproject.toml support
- [ ] docker-compose.yml support

### v1.0.0 — Stable
- [x] Multi-runner auto-detection (npm/yarn/pnpm/bun via lockfile)
- [ ] Multi-select to run several tasks in sequence
- [ ] Shell completions
- [ ] Rerun cache + `runtui rerun`
- [ ] Polished error handling and edge cases

---

## References

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — Pre-built TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [GoReleaser](https://goreleaser.com/) — Build & release automation
- [ntl](https://github.com/ruyadorno/ntl) — Inspiration for core UX
- [binary-install](https://www.npmjs.com/package/binary-install) — npm wrapper pattern

---

## License

MIT
