# Releasing

One-time setup, then every release is a single tag push.

## One-time setup (before first public release)

Do these in order:

1. **Create the Homebrew tap repo** (public, can be empty — GoReleaser pushes the formula):

   ```bash
   gh repo create pedromendonka/homebrew-tap --public --description "Homebrew formulas for tools"
   ```

2. **Set `TAP_GITHUB_TOKEN` secret** — a GitHub PAT with `repo` scope on `homebrew-tap` (fine-grained: Contents read/write on that repo only):

   ```bash
   gh secret set TAP_GITHUB_TOKEN -R pedromendonka/runtui
   ```

3. **Set `NPM_TOKEN` secret** — npm automation token (npmjs.com → Access Tokens → Granular, publish rights):

   ```bash
   gh secret set NPM_TOKEN -R pedromendonka/runtui
   ```

4. **Flip the repo public** — required for `go install`, brew downloads, and the npm postinstall (all pull from public GitHub):

   ```bash
   gh repo edit pedromendonka/runtui --visibility public
   ```

## Every release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The `Release` workflow then:

1. **goreleaser** — builds darwin/linux/windows (amd64+arm64, no windows/arm64), uploads archives + checksums to GitHub Releases, pushes the brew formula to `pedromendonka/homebrew-tap`.
2. **npm-publish** — stamps `npm/package.json` with the tag version, publishes `runtui` to npm. Users' `postinstall` downloads the matching release archive.

## Post-release verification

```bash
# Go
go install github.com/pedromendonka/runtui@latest && runtui --version

# Homebrew
brew tap pedromendonka/tap && brew install runtui && runtui --version

# npm
npm install -g runtui && runtui --version
```

## Notes

- Version comes from the git tag only — `main.go` `version` var is stamped via ldflags by GoReleaser; `npm/package.json` stays `0.0.0` in git and is stamped in CI.
- npm publish needs the GitHub release assets to already exist (`needs: goreleaser` enforces the order).
- Supported npm platforms: darwin/linux (x64, arm64), windows x64. Others get an error pointing at `go install`.
