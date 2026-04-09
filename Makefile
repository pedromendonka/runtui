BINARY    := runtui
MODULE    := github.com/pedromendonka/runtui
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -s -w -X main.version=$(VERSION)
GOFLAGS   := -trimpath

.PHONY: all build install run run-info test lint vet fmt clean setup release-dry help

all: build ## Build the binary (default)

build: ## Compile the binary to bin/runtui with version embedded
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY) .

install: ## Install the binary globally to $GOPATH/bin so you can run 'runtui' from anywhere
	go install $(GOFLAGS) -ldflags '$(LDFLAGS)' .

run: build ## Build and launch the TUI against testdata/package.json (quick manual test)
	cd testdata && ../bin/$(BINARY)

run-info: build ## Same as 'run' but with --info flag, which shows full commands without truncation
	cd testdata && ../bin/$(BINARY) --info

test: ## Run all unit tests with colored, grouped output (falls back to go test if gotestsum is not installed)
	@which gotestsum > /dev/null 2>&1 \
		&& gotestsum --format testdox -- ./... \
		|| go test ./... -v

vet: ## Run go vet — catches common mistakes like wrong printf formats or unreachable code
	go vet ./...

lint: vet ## Run go vet + staticcheck — deeper analysis for bugs, performance, and style issues
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed — skipping (go install honnef.co/go/tools/cmd/staticcheck@latest)"

fmt: ## Auto-format all Go files with gofmt (simplifies code too, e.g. removes unnecessary parens)
	gofmt -s -w .

clean: ## Delete the bin/ directory and clear Go's build cache
	rm -rf bin/
	go clean

setup: ## Set up git hooks — required once after cloning, enforces fmt/vet/build on commit and tests on push
	git config core.hooksPath .githooks
	@echo "Git hooks installed from .githooks/"

release-dry: ## Simulate a full release locally — builds all OS/arch binaries without publishing
	goreleaser release --snapshot --clean

help: ## Show all available targets with descriptions
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
