package main

import (
	"encoding/json"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// generator turns one embedded JSON asset into one Go source file. T is left open
// because the assets genuinely disagree on shape: the registries decode into a
// list of constants, the mining speeds into a nested map.
type generator[T any] struct {
	name     string
	pkg      string
	scalar   string
	asset    string
	output   string
	template *template.Template
}

// source is what every template renders against
type source[T any] struct {
	Name    string
	Version string
	Package string
	Type    string
	Values  T
}

// the paths are relative on purpose: every generator runs through go:generate
// from the package that owns the file it writes
func (g generator[T]) command() *cobra.Command {
	return &cobra.Command{
		Use:   g.name + " <protocol>",
		Short: "Generate " + g.output + " for a codec version from its embedded assets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.render(args[0])
		},
	}
}

func (g generator[T]) render(protocol string) error {
	// the protocol only stamps the header, so checking it here is what stops that
	// header from claiming a version the project has never heard of
	_, err := releaseOf(protocol)
	if err != nil {
		return err
	}

	raw, err := os.ReadFile("../assets/" + g.asset)
	if err != nil {
		return err
	}

	var values T
	if err := json.Unmarshal(raw, &values); err != nil {
		return err
	}

	rendered := source[T]{
		Name:    g.name,
		Version: protocol,
		Package: g.pkg,
		Type:    g.scalar,
		Values:  values,
	}

	var out strings.Builder
	if err := g.template.Execute(&out, rendered); err != nil {
		return err
	}

	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return err
	}

	return os.WriteFile(g.output, formatted, 0o644)
}
