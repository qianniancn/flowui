// Command cataloggen regenerates catalog_gen.go from the installed Lucide
// icon module. It reads the generated var declarations because the icon module
// intentionally exports icons as individual values rather than a runtime list.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const module = "github.com/qianniancn/flowui-icons-lucide"

var declaration = regexp.MustCompile(`(?m)^var ([A-Za-z0-9_]+) = Data\("`)

func main() {
	dir, err := output("go", "list", "-f", "{{.Dir}}", module)
	if err != nil {
		fail(err)
	}
	version, err := output("go", "list", "-m", "-f", "{{.Version}}", module)
	if err != nil {
		fail(err)
	}

	entries, err := filepath.Glob(filepath.Join(dir, "icons_*_gen.go"))
	if err != nil {
		fail(err)
	}
	names := make([]string, 0, 1800)
	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			fail(err)
		}
		for _, match := range declaration.FindAllSubmatch(data, -1) {
			names = append(names, string(match[1]))
		}
	}
	sort.Slice(names, func(i, j int) bool {
		left, right := strings.ToLower(names[i]), strings.ToLower(names[j])
		if left == right {
			return names[i] < names[j]
		}
		return left < right
	})

	var builder strings.Builder
	fmt.Fprintf(&builder, "// Code generated from %s %s; DO NOT EDIT.\n\n", module, version)
	builder.WriteString("package main\n\n")
	builder.WriteString("import lucide \"github.com/qianniancn/flowui-icons-lucide\"\n\n")
	builder.WriteString("type catalogIcon struct {\n\tname string\n\tdata lucide.Data\n}\n\n")
	builder.WriteString("var catalog = []catalogIcon{\n")
	for _, name := range names {
		fmt.Fprintf(&builder, "\t{%q, lucide.%s},\n", name, name)
	}
	builder.WriteString("}\n")

	if err := os.WriteFile("catalog_gen.go", []byte(builder.String()), 0o644); err != nil {
		fail(err)
	}
}

func output(name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	data, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(data)), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "cataloggen:", err)
	os.Exit(1)
}
