# RUNTUI — Open-Source Refactoring Plan

> Created: 2026-04-10
> Goal: Refactor and improve runtui into a polished open-source project with Makefile support.

---

## Phase 1: Complete blocked Phase 2 — commit community health files

**Goal:** Land all the in-progress changes that got blocked (new templates, CI formatting check, Makefile `fmt-check` target, simplified pre-commit hook) and update CONTRIBUTING.md to match.

**Changes:**
- `.githooks/pre-commit` — already correct (simplified `go vet ./...`), just needs committing
- `.github/workflows/ci.yml` — already correct (added formatting check step), just needs committing
- `Makefile` — already correct (added `fmt-check` target, wired into `lint`), just needs committing
- `.github/ISSUE_TEMPLATE/bug_report.yml` — already staged, ready
- `.github/ISSUE_TEMPLATE/feature_request.yml` — already staged, ready
- `.github/ISSUE_TEMPLATE/config.yml` — already staged, ready
- `.github/PULL_REQUEST_TEMPLATE.md` — already staged, ready
- `CONTRIBUTING.md` — update Make targets table to add `fmt-check` and `tidy-check`; fix "Adding a new project type" section to match current `Parser` interface (returns `RunContext` now, not just `[]Task, error`); update architecture package list to include `app/`

**Depends on:** None
**Estimated effort:** S
**Tests:** `make lint && make test` pass
**Commit message:** `main | Community Health | Add issue/PR templates, CI format check, fmt-check target`

---

## Phase 2: Makefile parser

**Goal:** Implement a parser that extracts `make` targets and their `##` comments as descriptions, wired through detector and registry.

**Changes:**
- `detector/detector.go` — add `TypeMakefile` constant and `{"Makefile"}` entry to `configFiles`; refactor `detectRunner` so it only runs for package.json projects (Makefile projects don't need lockfile detection)
- `detector/detector_test.go` — add tests for Makefile detection, mixed project detection (both package.json + Makefile)
- `parser/makefile.go` — implement `MakefileParser` that reads a Makefile, extracts targets (lines matching `^target-name:`) and `## comment` descriptions; returns `RunContext{Binary: "make"}` with no subcmd/separator
- `parser/makefile_test.go` — test target extraction, `##` descriptions, `.PHONY` exclusion, empty Makefile, mixed real/phony targets
- `testdata/Makefile` — sample Makefile fixture with `##` comments
- `app/registry.go` — register `detector.TypeMakefile` → `MakefileParser` factory

**Depends on:** Phase 1
**Estimated effort:** M
**Tests:** New parser and detector tests pass; `make test` green
**Commit message:** `main | Makefile Parser | Detect and parse Makefile targets with ## descriptions`

---

## Phase 3: Multi-project selection

**Goal:** When both `package.json` and `Makefile` exist in the same directory, let the user pick which project type to use (instead of silently picking the first).

**Changes:**
- `app/app.go` — when `len(projects) > 1`, add a `--type` flag to force selection; when no flag given, launch a minimal selection prompt
- `app/app.go` — rethink `validRunners`: `--runner` only applies to npm-family projects; `make` doesn't need a runner override
- `app/app_test.go` — test multi-project selection, `--type` override, error on unknown type
- `detector/detector.go` — ensure ordering is deterministic (package.json first, then Makefile)

**Depends on:** Phase 2
**Estimated effort:** M
**Tests:** App tests cover multi-project detection; manual test in a dir with both files
**Commit message:** `main | Multi-Project | Add --type flag and project selection for mixed directories`

---

## Phase 4: GoReleaser Homebrew tap readiness

**Goal:** Configure GoReleaser so that tagging a release auto-publishes a Homebrew formula (once the `homebrew-tap` repo exists).

**Changes:**
- `.goreleaser.yml` — add `brews:` section pointing to `pedromendonka/homebrew-tap` with name, description, homepage, license
- `README.md` — add `--version` to CLI reference (implemented but not documented); update Features to mention Makefile support

**Depends on:** Phase 2
**Estimated effort:** S
**Tests:** `make release-dry` succeeds
**Commit message:** `main | Homebrew | Add brews config to GoReleaser for tap auto-publishing`

> **Note:** Phases 3 and 4 are independent of each other and can run in parallel after Phase 2.

---

## Phase 5: Docs refresh & continuation plan

**Goal:** Update all documentation to reflect the current state, mark completed milestones, and create a continuation roadmap for future work.

**Changes:**
- `PLAN.md` — replace `<username>` placeholders with `pedromendonka`; update interface examples to match current `Parser` signature; update milestone checklist (mark completed items); update project structure to include `app/`
- `README.md` — add Makefile to supported project types; update roadmap checklist (mark Makefile as done); add `--type` flag to CLI reference
- `CONTRIBUTING.md` — update "Adding a new project type" with a real second example (now that Makefile exists); update architecture diagram
- `ROADMAP.md` (new) — lightweight continuation plan for future phases:
  - Cargo.toml parser
  - pyproject.toml parser
  - docker-compose.yml service runner
  - Taskfile.yml / justfile support
  - Rerun cache (`runtui rerun`)
  - Shell completions
  - v0.1.0 tag + first release

**Depends on:** Phases 3 and 4
**Estimated effort:** S
**Tests:** No code changes — docs only
**Commit message:** `main | Docs | Refresh docs for open-source launch, add roadmap`

---

## Summary

| Phase | Description                         | Effort | Depends on |
|-------|-------------------------------------|--------|------------|
| 1     | Land blocked community health files | S      | None       |
| 2     | Makefile parser implementation      | M      | 1          |
| 3     | Multi-project selection (`--type`)  | M      | 2          |
| 4     | GoReleaser Homebrew tap config      | S      | 2          |
| 5     | Docs refresh + continuation plan    | S      | 3, 4       |

---

## How to resume

If this plan is interrupted, pick up from the first phase whose commit message is **not** in `git log --oneline`. All phases are independently committable — the codebase is in a working state after each one.
