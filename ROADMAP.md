# RUNTUI — Roadmap

> What's next after the open-source refactoring.

---

## Next up

### Cargo.toml parser
- Parse `[package.metadata.scripts]` or common cargo subcommands
- RunContext: `Binary="cargo"`, `ArgSeparator="--"`
- Detect from `Cargo.toml` presence

### pyproject.toml parser
- Parse `[tool.poetry.scripts]` and `[tool.pdm.scripts]`
- Auto-detect runner from `poetry.lock` vs `pdm.lock`

### docker-compose.yml service runner
- Parse service names from `docker-compose.yml` / `compose.yml`
- RunContext: `Binary="docker"`, `Subcmd="compose run"`

### Taskfile.yml / justfile
- Parse task/recipe definitions
- Lower priority — niche but requested

---

## Features

### Rerun cache
- Cache last executed task + args per project directory
- `runtui rerun` shorthand to instantly repeat
- Store in `~/.cache/runtui/` or XDG cache dir

### Multi-select
- Select multiple tasks to run in sequence
- Toggle selection with space, run all with enter

### Shell completions
- Generate bash/zsh/fish completions via `runtui completion <shell>`
- Register task names for tab-completion

---

## Distribution

### First tagged release (v0.1.0)
- Create `pedromendonka/homebrew-tap` repo on GitHub
- Add `TAP_GITHUB_TOKEN` secret (PAT with `repo` scope) to runtui repo
- Tag and push: `git tag v0.1.0 && git push --tags`
- Verify: GitHub Release created, Homebrew formula pushed, `go install` works

### npm wrapper (optional)
- Thin npm package that downloads the correct Go binary on `postinstall`
- Follow the pattern used by esbuild and turbo

---

## Quality

### Coverage reporting
- Add Codecov or Coveralls integration to CI
- Badge in README

### Benchmarks
- Startup time benchmark (detect + parse + TUI init)
- Parser benchmarks for large Makefiles / package.json files

### Demo GIF
- Record a terminal session with VHS or asciinema
- Add to README hero section
