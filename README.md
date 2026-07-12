# runtui

> One TUI to run them all.

**runtui** is an interactive terminal UI that auto-detects your project type, lists all available tasks, and lets you select, configure arguments, and run them -- without memorizing command names.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Distributed as a single binary.

```
╭───────────────────────────────────╮
│  ╦═╗  ╦ ╦  ╔╗ ╔  ╔╦╗  ╦ ╦  ╦    │
│  ╠╦╝  ║ ║  ║╚╗║   ║   ║ ║  ║    │
│  ║╚╗  ║ ║  ║ ╚║   ║   ║ ║  ║    │
│  ╩ ╚  ╚═╝  ╝  ╚   ╩   ╚═╝  ╩    │
│  One TUI to run them all.        │
╰───────────────────────────────────╯
  package.json (npm)

    Task             Description                     Command
    ─────────────────────────────────────────────────────────────
  → dev              Start development server        next dev
    build                                            tsc && node build.js
    test             Run test suite                   jest
    env:set          Set an environment variable      dotenvx set
    env:filter       Filter environment variables     node scripts/filter.js
    deploy           Deploy the application           ./deploy.sh

  / │
  ↑↓ navigate  ·  / filter  ·  enter run  ·  q quit
```

---

## Features

- **Auto-detect project type** -- scans for `package.json` and `Makefile`
- **Auto-detect package manager** -- picks `npm`, `yarn`, `pnpm`, or `bun` from your lockfile
- **Tabular task list** -- tasks displayed in aligned columns with descriptions and commands
- **Filterable** -- type `/` to filter tasks by substring
- **Argument prompting** -- every task gets an argument prompt before execution
  - **Simple mode**: free-form `Arguments:` input for any task
  - **Config-driven mode**: structured fields with labels, required markers, and hints
- **Full terminal passthrough** -- the TUI suspends while your command runs, then resumes
- **Execution feedback** -- success/failure badges with exit codes after each run
- **Run multiple tasks** -- after a task finishes, you're back in the list to pick another
- **Single binary** -- no runtime dependencies, runs anywhere

---

## Install

### Homebrew (macOS/Linux)

```bash
brew tap pedromendonka/tap
brew install runtui
```

### npm

```bash
npm install -g runtui
# or run without installing:
npx runtui
```

### Go

```bash
go install github.com/pedromendonka/runtui@latest
```

### Binary download

Grab the latest release from [GitHub Releases](https://github.com/pedromendonka/runtui/releases) and place it in your `PATH`.

---

## Usage

```bash
# Launch in current directory
runtui

# Show full commands without truncation
runtui --info

# Override detected package manager
runtui --runner=pnpm

# Force project type when both package.json and Makefile exist
runtui --type=Makefile
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `up` / `down` | Navigate task list |
| Type any character | Filter tasks by name |
| `Backspace` | Delete filter character |
| `Esc` | Clear filter (or quit if filter is empty) |
| `Enter` | Select task / confirm arguments / run |
| `Tab` / `Shift+Tab` | Navigate between argument fields |
| `q` | Quit (when filter is empty) |
| `Ctrl+C` | Quit immediately |

### Flow

1. **runtui** scans the current directory for project config files
2. Parses tasks and displays them in a tabular list
3. You select a task and enter arguments (or skip)
4. The command runs with full terminal access
5. A success/failure banner shows the result
6. You're back in the list to run another task

---

## Configuration

Add a `runtui` key to your `package.json` to enhance any script with descriptions and structured arguments.

### Basic example

```json
{
  "scripts": {
    "dev": "next dev",
    "build": "tsc && node build.js",
    "test": "jest"
  },
  "runtui": {
    "dev": {
      "description": "Start development server"
    },
    "test": {
      "description": "Run test suite"
    }
  }
}
```

Descriptions appear in their own column in the task list.

### Structured arguments

Define arguments per-script with names, required flags, hints, and defaults:

```json
{
  "scripts": {
    "env:set": "dotenvx set",
    "deploy": "./deploy.sh"
  },
  "runtui": {
    "env:set": {
      "description": "Set an environment variable",
      "args": [
        { "name": "KEY", "required": true, "hint": "e.g. DATABASE_URL" },
        { "name": "VALUE", "required": true }
      ]
    },
    "deploy": {
      "description": "Deploy the application",
      "args": [
        { "name": "ENVIRONMENT", "required": true, "hint": "staging or production" },
        { "name": "TAG", "required": false, "default": "latest" }
      ]
    }
  }
}
```

When you select `env:set`, runtui shows a structured form:

```
╭───────────────────────────────────╮
│  ╦═╗  ╦ ╦  ╔╗ ╔  ╔╦╗  ╦ ╦  ╦    │
│  ╠╦╝  ║ ║  ║╚╗║   ║   ║ ║  ║    │
│  ║╚╗  ║ ║  ║ ╚║   ║   ║ ║  ║    │
│  ╩ ╚  ╚═╝  ╝  ╚   ╩   ╚═╝  ╩    │
╰───────────────────────────────────╯
  env:set  dotenvx set

  KEY   *  DATABASE_URL│
           e.g. DATABASE_URL

  VALUE *  │

  tab next  ·  enter run  ·  esc back
```

### Argument fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Argument label shown in the prompt |
| `required` | bool | If `true`, must be filled before running |
| `hint` | string | Help text shown below the field |
| `default` | string | Pre-filled value |

Tasks **without** a `runtui.args` config get a simple free-form prompt:

```
  Arguments: --coverage --watch│

  enter run (empty to skip)  ·  esc back
```

---

## Makefile support

runtui parses Makefile targets and uses `## comment` annotations as descriptions — the same self-documenting pattern used by many projects:

```makefile
build: ## Compile the binary
	go build -o bin/app .

test: ## Run all tests
	go test ./...

lint: vet fmt-check ## Run all linters
```

If any targets have `##` descriptions, only those are shown (the curated public API). If none have descriptions, all targets are listed.

When both `package.json` and `Makefile` exist in the same directory, runtui picks `package.json` by default and prints a hint:

```
runtui: also detected Makefile — use --type=Makefile to switch
```

---

## Execution feedback

After a task runs, a status banner shows the result:

```
  ✓ DONE   dev
           npm run dev
```

If a task fails, the exit code is shown:

```
  ✗ FAIL   test  exit 1
           npm run test -- --coverage
```

---

## Package manager detection

runtui auto-detects your package manager by checking for lockfiles:

| Lockfile | Runner |
|----------|--------|
| `bun.lockb` / `bun.lock` | `bun` |
| `pnpm-lock.yaml` | `pnpm` |
| `yarn.lock` | `yarn` |
| `package-lock.json` | `npm` |
| *(none found)* | `npm` (default) |

Override with `--runner`:

```bash
runtui --runner=yarn
```

---

## CLI reference

```
Usage: runtui [flags]

Flags:
  --info           Show full commands without truncation
  --runner=NAME    Override detected package manager (npm, yarn, pnpm, bun)
  --type=TYPE      Project type to use when multiple are detected (e.g. package.json, Makefile)
  --version        Print version and exit
  --help           Show help
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture, code style, how to add new project types, and release process.

Quick start:

```bash
git clone https://github.com/pedromendonka/runtui.git
cd runtui
make build    # build to bin/runtui
make run      # run against testdata/
make test     # run all tests
make help     # see all targets
```

---

## Roadmap

- [x] Makefile target parser
- [ ] Cargo.toml support
- [ ] pyproject.toml (poetry/pdm) support
- [ ] docker-compose.yml service runner
- [ ] Taskfile.yml / justfile support
- [ ] Rerun last command (`runtui rerun`)
- [ ] Multi-select to run tasks in sequence
- [ ] Shell completions

---

## Compared to

| Tool | Language | Limitations |
|------|----------|-------------|
| [ntl](https://github.com/ruyadorno/ntl) | Node.js | No argument prompting, Node-only, appears dormant |
| [lazynpm](https://github.com/jesseduffield/lazygit) | Go | npm-only, no arg prompting |
| [nrun](https://github.com/nicolo-ribaudo/nrun) | Node.js | Not a TUI, no interactive selection |

**runtui** adds structured argument prompting, multi-ecosystem support, and a single Go binary with no runtime dependencies.

---

## License

[MIT](LICENSE)
