package parser

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

// targetLine matches a Makefile target name at the start of a line.
var targetLine = regexp.MustCompile(`^([a-zA-Z0-9_][a-zA-Z0-9_.-]*)\s*:`)

// MakefileParser extracts targets from a Makefile.
type MakefileParser struct{}

func (p *MakefileParser) Parse(path string) ([]Task, RunContext, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, RunContext{}, err
	}
	return p.parse(data)
}

func (p *MakefileParser) parse(data []byte) ([]Task, RunContext, error) {
	runCtx := RunContext{
		Binary:        "make",
		Subcmd:        "",
		ArgSeparator:  "",
		DisplayPrefix: []string{"make"},
	}

	type entry struct {
		name string
		desc string
	}

	var entries []entry
	hasDocumented := false

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines, recipes (tab-indented), comments, dot-prefixed targets.
		if len(line) == 0 || line[0] == '\t' || line[0] == '#' || line[0] == '.' {
			continue
		}

		// Skip variable assignments (:=, ::=, ?=, +=).
		if strings.Contains(line, ":=") || strings.Contains(line, "?=") || strings.Contains(line, "+=") {
			continue
		}

		m := targetLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		name := m[1]

		var desc string
		if i := strings.Index(line, "##"); i >= 0 {
			desc = strings.TrimSpace(line[i+2:])
			hasDocumented = true
		}

		entries = append(entries, entry{name, desc})
	}

	// If some targets have ## descriptions, only show those (curated public API).
	// If none have descriptions, show all targets.
	tasks := make([]Task, 0, len(entries))
	for _, e := range entries {
		if hasDocumented && e.desc == "" {
			continue
		}
		tasks = append(tasks, Task{
			Name:        e.name,
			Command:     "make " + e.name,
			Description: e.desc,
		})
	}

	sortTasks(tasks)

	return tasks, runCtx, nil
}
