package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pedromendonka/runtui/detector"
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

func TestSelectProjectDefault(t *testing.T) {
	projects := []detector.Project{
		{Type: detector.TypePackageJSON, Path: "package.json", Runner: "npm"},
		{Type: detector.TypeMakefile, Path: "Makefile", Runner: "make"},
	}
	p, err := selectProject(projects, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type != detector.TypePackageJSON {
		t.Errorf("type = %q, want %q (first by priority)", p.Type, detector.TypePackageJSON)
	}
}

func TestSelectProjectByType(t *testing.T) {
	projects := []detector.Project{
		{Type: detector.TypePackageJSON, Path: "package.json", Runner: "npm"},
		{Type: detector.TypeMakefile, Path: "Makefile", Runner: "make"},
	}
	p, err := selectProject(projects, "Makefile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type != detector.TypeMakefile {
		t.Errorf("type = %q, want %q", p.Type, detector.TypeMakefile)
	}
}

func TestSelectProjectCaseInsensitive(t *testing.T) {
	projects := []detector.Project{
		{Type: detector.TypeMakefile, Path: "Makefile", Runner: "make"},
	}
	p, err := selectProject(projects, "makefile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type != detector.TypeMakefile {
		t.Errorf("type = %q, want %q", p.Type, detector.TypeMakefile)
	}
}

func TestSelectProjectNotFound(t *testing.T) {
	projects := []detector.Project{
		{Type: detector.TypePackageJSON, Path: "package.json", Runner: "npm"},
	}
	_, err := selectProject(projects, "Makefile")
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !strings.Contains(err.Error(), "no \"Makefile\" project found") {
		t.Errorf("error = %q, want no-project-found message", err.Error())
	}
}

func TestRunTypeFlagSelectsMakefile(t *testing.T) {
	dir := t.TempDir()
	// package.json with scripts (would normally be selected first).
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"dev":"echo hi"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	// Empty Makefile (no targets → "no tasks found").
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	withCwd(t, dir, func() {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"--type=Makefile"}, &stdout, &stderr, "dev")

		if code != ExitError {
			t.Errorf("exit code = %d, want %d", code, ExitError)
		}
		// Should get "no tasks found" from the Makefile parser, not launch package.json TUI.
		if !strings.Contains(stderr.String(), "no tasks found") {
			t.Errorf("stderr = %q, want no-tasks message from Makefile", stderr.String())
		}
	})
}

func TestRunTypeFlagNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	withCwd(t, dir, func() {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"--type=Makefile"}, &stdout, &stderr, "dev")

		if code != ExitError {
			t.Errorf("exit code = %d, want %d", code, ExitError)
		}
		if !strings.Contains(stderr.String(), "no \"Makefile\" project found") {
			t.Errorf("stderr = %q, want type-not-found message", stderr.String())
		}
	})
}

func TestRunMultiProjectHint(t *testing.T) {
	dir := t.TempDir()
	// package.json with no scripts → "no tasks found", but should still hint about Makefile.
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build .\n"), 0644); err != nil {
		t.Fatal(err)
	}

	withCwd(t, dir, func() {
		var stdout, stderr bytes.Buffer
		Run(nil, &stdout, &stderr, "dev")

		if !strings.Contains(stderr.String(), "also detected Makefile") {
			t.Errorf("stderr = %q, want multi-project hint", stderr.String())
		}
	})
}

func TestParserForRegistered(t *testing.T) {
	p, err := parserFor(detector.TypePackageJSON, "pnpm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("parser is nil")
	}
}

func TestParserForUnregistered(t *testing.T) {
	_, err := parserFor(detector.ProjectType("Cargo.toml"), "")
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
	if !strings.Contains(err.Error(), "unsupported project type") {
		t.Errorf("error = %q, want unsupported project type", err.Error())
	}
}

func TestParserForMakefile(t *testing.T) {
	p, err := parserFor(detector.TypeMakefile, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("parser is nil")
	}
}
