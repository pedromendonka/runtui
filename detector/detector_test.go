package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	projects, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != TypePackageJSON {
		t.Errorf("type = %q, want %q", projects[0].Type, TypePackageJSON)
	}
	if projects[0].Runner != "npm" {
		t.Errorf("runner = %q, want npm (default)", projects[0].Runner)
	}
}

func TestDetectNoConfigFiles(t *testing.T) {
	dir := t.TempDir()

	projects, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDetectRunnerFromLockfiles(t *testing.T) {
	tests := []struct {
		lockfile string
		want     string
	}{
		{"package-lock.json", "npm"},
		{"yarn.lock", "yarn"},
		{"pnpm-lock.yaml", "pnpm"},
		{"bun.lockb", "bun"},
		{"bun.lock", "bun"},
	}

	for _, tt := range tests {
		t.Run(tt.lockfile, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, tt.lockfile), []byte{}, 0644); err != nil {
				t.Fatal(err)
			}

			projects, err := Detect(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(projects) != 1 {
				t.Fatalf("expected 1 project, got %d", len(projects))
			}
			if projects[0].Runner != tt.want {
				t.Errorf("runner = %q, want %q", projects[0].Runner, tt.want)
			}
		})
	}
}

func TestDetectLockfilePriority(t *testing.T) {
	dir := t.TempDir()
	// Create package.json + both bun and npm lockfiles — bun should win (checked first).
	for _, f := range []string{"package.json", "bun.lockb", "package-lock.json"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(`{}`), 0644); err != nil {
			t.Fatal(err)
		}
	}

	projects, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projects[0].Runner != "bun" {
		t.Errorf("runner = %q, want bun (higher priority)", projects[0].Runner)
	}
}
