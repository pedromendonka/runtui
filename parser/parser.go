package parser

// Task represents a runnable task extracted from a project config file.
type Task struct {
	Name        string
	Command     string
	Description string
	Args        []ArgDef
}

// ArgDef defines a structured argument that a task accepts.
type ArgDef struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Hint     string `json:"hint,omitempty"`
	Default  string `json:"default,omitempty"`
}

// Parser extracts tasks from a project config file.
type Parser interface {
	Parse(path string) ([]Task, error)
}
