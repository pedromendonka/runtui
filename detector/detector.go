package detector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ProjectType identifies the kind of project config detected.
type ProjectType string

const (
	TypePackageJSON ProjectType = "package.json"
)

// Project represents a detected project configuration file.
type Project struct {
	Type   ProjectType
	Path   string
	Runner string // auto-detected package manager
}

// configFiles maps each project type to its expected filename.
var configFiles = []struct {
	Type     ProjectType
	Filename string
}{
	{TypePackageJSON, "package.json"},
}

// lockfileRunners maps lockfile names to their package manager command.
// Order matters: first match wins.
var lockfileRunners = []struct {
	Filename string
	Runner   string
}{
	{"bun.lockb", "bun"},
	{"bun.lock", "bun"},
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "yarn"},
	{"package-lock.json", "npm"},
}

// Detect scans dir for known project config files and returns all matches.
func Detect(dir string) ([]Project, error) {
	var projects []Project

	for _, cf := range configFiles {
		path := filepath.Join(dir, cf.Filename)
		if _, err := os.Stat(path); err == nil {
			projects = append(projects, Project{
				Type:   cf.Type,
				Path:   path,
				Runner: detectRunner(dir),
			})
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("checking %s: %w", path, err)
		}
	}

	return projects, nil
}

func detectRunner(dir string) string {
	for _, lf := range lockfileRunners {
		if _, err := os.Stat(filepath.Join(dir, lf.Filename)); err == nil {
			return lf.Runner
		}
	}
	return "npm"
}
