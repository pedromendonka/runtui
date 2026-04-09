package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withCwd runs f inside dir, restoring the previous working directory
// afterwards. Needed because app.Run uses os.Getwd to find the project.
func withCwd(t *testing.T, dir string, f func()) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	f()
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--version"}, &stdout, &stderr, "1.2.3")

	if code != ExitOK {
		t.Errorf("exit code = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(stdout.String(), "runtui 1.2.3") {
		t.Errorf("stdout = %q, want to contain %q", stdout.String(), "runtui 1.2.3")
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunInvalidRunner(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--runner=invalid"}, &stdout, &stderr, "dev")

	if code != ExitError {
		t.Errorf("exit code = %d, want %d", code, ExitError)
	}
	if !strings.Contains(stderr.String(), "unsupported runner") {
		t.Errorf("stderr = %q, want to contain %q", stderr.String(), "unsupported runner")
	}
}

func TestRunNoProjectFiles(t *testing.T) {
	dir := t.TempDir()

	withCwd(t, dir, func() {
		var stdout, stderr bytes.Buffer
		code := Run(nil, &stdout, &stderr, "dev")

		if code != ExitError {
			t.Errorf("exit code = %d, want %d", code, ExitError)
		}
		if !strings.Contains(stderr.String(), "no supported project files found") {
			t.Errorf("stderr = %q, want no-supported-project message", stderr.String())
		}
	})
}

func TestRunMalformedPackageJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	withCwd(t, dir, func() {
		var stdout, stderr bytes.Buffer
		code := Run(nil, &stdout, &stderr, "dev")

		if code != ExitError {
			t.Errorf("exit code = %d, want %d", code, ExitError)
		}
		if !strings.Contains(stderr.String(), "parsing package.json") {
			t.Errorf("stderr = %q, want parse error", stderr.String())
		}
	})
}

func TestRunNoTasks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "empty"}`), 0644); err != nil {
		t.Fatal(err)
	}

	withCwd(t, dir, func() {
		var stdout, stderr bytes.Buffer
		code := Run(nil, &stdout, &stderr, "dev")

		if code != ExitError {
			t.Errorf("exit code = %d, want %d", code, ExitError)
		}
		if !strings.Contains(stderr.String(), "no tasks found") {
			t.Errorf("stderr = %q, want no-tasks message", stderr.String())
		}
	})
}

func TestRunUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--does-not-exist"}, &stdout, &stderr, "dev")

	if code != ExitError {
		t.Errorf("exit code = %d, want %d", code, ExitError)
	}
}

func TestParserForRegistered(t *testing.T) {
	p, err := parserFor("package.json", "pnpm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("parser is nil")
	}
}

func TestParserForUnregistered(t *testing.T) {
	_, err := parserFor("Makefile", "")
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
	if !strings.Contains(err.Error(), "unsupported project type") {
		t.Errorf("error = %q, want unsupported project type", err.Error())
	}
}
