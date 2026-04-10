BINARY := example
VERSION := 1.0.0

.PHONY: all build test lint clean deploy

all: build ## Build everything (default)

build: ## Compile the binary
	go build -o bin/$(BINARY) .

test: ## Run all tests
	go test ./...

lint: vet fmt-check ## Run all linters

vet:
	go vet ./...

fmt-check:
	gofmt -l .

clean: ## Remove build artifacts
	rm -rf bin/

deploy: ## Deploy the application
	./deploy.sh
