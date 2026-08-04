package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

const (
	jsxFactory  = "__element"
	jsxFragment = "__fragment"

	apiModule = "gocraft"
)

var sourceKinds = [...]string{".ts", ".tsx", ".js", ".jsx"}

type Source struct {
	Name string
	Path string
	Code string
}

func Collect(dir string) ([]Source, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var found []Source
	for _, entry := range entries {
		if entry.IsDir() || !plugable(entry.Name()) {
			continue
		}

		compiled, err := Compile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		found = append(found, compiled)
	}

	return found, nil
}

func plugable(name string) bool {
	if strings.HasSuffix(name, ".d.ts") {
		return false
	}

	return slices.Contains(sourceKinds[:], filepath.Ext(name))
}

func Compile(path string) (Source, error) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	options := api.BuildOptions{
		EntryPoints: []string{path},
		Bundle:      true,
		Write:       false,
		Format:      api.FormatCommonJS,
		Platform:    api.PlatformNeutral,
		Target:      api.ES2020,
		JSXFactory:  jsxFactory,
		JSXFragment: jsxFragment,

		External: []string{apiModule},
	}

	built := api.Build(options)
	if len(built.Errors) > 0 {
		return Source{}, fmt.Errorf("plugin %s: %s", name, describe(built.Errors))
	}
	if len(built.OutputFiles) == 0 {
		return Source{}, fmt.Errorf("plugin %s: compiled to nothing", name)
	}

	return Source{
		Name: name,
		Path: path,
		Code: string(built.OutputFiles[0].Contents),
	}, nil
}

func describe(failures []api.Message) string {
	lines := make([]string, 0, len(failures))
	for _, failure := range failures {
		if failure.Location == nil {
			lines = append(lines, failure.Text)

			continue
		}

		lines = append(lines, fmt.Sprintf("%s:%d: %s", failure.Location.File, failure.Location.Line, failure.Text))
	}

	return strings.Join(lines, "; ")
}
