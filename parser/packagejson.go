package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

// packageJSON mirrors the relevant fields of a package.json file.
type packageJSON struct {
	Scripts map[string]string           `json:"scripts"`
	RunTUI  map[string]runtuiTaskConfig `json:"runtui"`
}

// runtuiTaskConfig holds per-script metadata defined under the "runtui" key.
type runtuiTaskConfig struct {
	Description string   `json:"description"`
	Args        []ArgDef `json:"args"`
}

// PackageJSON parses tasks from a package.json file.
type PackageJSON struct{}

func (p *PackageJSON) Parse(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return p.parse(data)
}

func (p *PackageJSON) parse(data []byte) ([]Task, error) {
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parsing package.json: %w", err)
	}

	if len(pkg.Scripts) == 0 {
		return nil, nil
	}

	tasks := make([]Task, 0, len(pkg.Scripts))
	for name, cmd := range pkg.Scripts {
		task := Task{
			Name:    name,
			Command: cmd,
		}
		if cfg, ok := pkg.RunTUI[name]; ok {
			task.Description = cfg.Description
			task.Args = cfg.Args
		}
		tasks = append(tasks, task)
	}

	slices.SortFunc(tasks, func(a, b Task) int {
		return strings.Compare(a.Name, b.Name)
	})

	return tasks, nil
}
