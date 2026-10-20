package window

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/theme"
)

// TestCustomizeThemeOptionQueuesThemeMutation pins the surviving way to reshape
// a window's theme now that the design option is gone: options queue theme
// mutations that are applied in order when the window resolves.
func TestCustomizeThemeOptionQueuesThemeMutation(t *testing.T) {
	want := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	cfg := Resolve(CustomizeTheme(func(current *theme.Theme) {
		current.Palette.Accent = want
	}))
	if len(cfg.ThemeOps) != 1 {
		t.Fatalf("theme option count = %d, want 1", len(cfg.ThemeOps))
	}
	active := theme.NeutralTheme()
	for _, apply := range cfg.ThemeOps {
		apply(&active)
	}
	if active.Palette.Accent != want {
		t.Fatalf("accent = %v, want the option value %v", active.Palette.Accent, want)
	}
}
