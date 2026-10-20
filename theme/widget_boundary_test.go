package theme_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWidgetsDoNotOwnAppearance guards the boundary the visual protocol exists
// for: widget packages describe behavior and layout, never concrete palette
// values or theme appearance reads. Composition packages pass that appearance
// down to the child widgets whose providers own it.
//
// code.go is the single exception: syntax highlighting reads the theme's
// CodeColors palette, which is theme data rather than a component surface.
func TestWidgetsDoNotOwnAppearance(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "widget", "*", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	allowedThemeReads := map[string]bool{
		"code.go": true,
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		if !allowedThemeReads[filepath.Base(path)] {
			for _, read := range []string{".Semantic()", ".Components."} {
				if strings.Contains(text, read) {
					t.Errorf("%s reads theme appearance (%s); expose it through a provider visual", path, read)
				}
			}
		}
		for _, literal := range []string{"style.RGB(", "style.RGBA("} {
			if strings.Contains(text, literal) {
				t.Errorf("%s hardcodes a palette value (%s); theme recipes own colors", path, literal)
			}
		}
	}
}
