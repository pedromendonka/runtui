package app

import (
	"fmt"

	"github.com/pedromendonka/runtui/detector"
	"github.com/pedromendonka/runtui/parser"
)

// parserFactory builds a Parser for a given detected project.
// The runner string is only meaningful for package-manager-based project
// types (npm/yarn/pnpm/bun) — other factories ignore it.
type parserFactory func(runner string) parser.Parser

// parserRegistry maps detector.ProjectType to the factory that builds the
// matching parser. Adding a new project type means registering it here
// (and adding the config file to detector.configFiles).
var parserRegistry = map[detector.ProjectType]parserFactory{
	detector.TypePackageJSON: func(runner string) parser.Parser {
		return parser.NewPackageJSON(runner)
	},
}

// parserFor returns the parser for the given project type, or an error
// if the type is not registered.
func parserFor(t detector.ProjectType, runner string) (parser.Parser, error) {
	factory, ok := parserRegistry[t]
	if !ok {
		return nil, fmt.Errorf("unsupported project type: %s", t)
	}
	return factory(runner), nil
}
