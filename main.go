// Command runtui is an interactive TUI for running project tasks.
//
// All orchestration lives in the app package. main.go is intentionally a
// thin shim so Run can be exercised from tests without touching os.Exit.
package main

import (
	"os"

	"github.com/pedromendonka/runtui/app"
)

// version is injected at build time via -ldflags '-X main.version=...'.
var version = "dev"

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr, version))
}
