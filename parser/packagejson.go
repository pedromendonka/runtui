package parser

import (
	"encoding/json"
	"fmt"
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
// The runner is the package manager command (npm, yarn, pnpm, bun) that
// should be used to execute scripts — typically chosen by the detector
// from the presence of a lockfile.
type PackageJSON struct {
	runner string
}

// NewPackageJSON returns a parser configured for the given package manager.
func NewPackageJSON(runner string) *PackageJSON {
	return &PackageJSON{runner: runner}
}

func (p *PackageJSON) Parse(path string) ([]Task, RunContext, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, RunContext{}, err
	}
	return p.parse(data)
}

func (p *PackageJSON) parse(data []byte) ([]Task, RunContext, error) {
	runCtx := RunContext{
		Binary:        p.runner,
		Subcmd:        "run",
		ArgSeparator:  "--",
		DisplayPrefix: []string{p.runner, "run"},
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, runCtx, fmt.Errorf("parsing package.json: %w", err)
	}

	if len(pkg.Scripts) == 0 {
		return nil, runCtx, nil
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

	sortTasks(tasks)

	return tasks, runCtx, nil
}
