package theme_test

import (
	"image/color"
	"testing"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestDisabledColorUsesOpacity(t *testing.T) {
	theme := theme.DefaultTheme()
	theme.DisabledOpacity = 0.25

	got := theme.DisabledColor(color.NRGBA{R: 1, G: 2, B: 3, A: 200})
	want := color.NRGBA{R: 1, G: 2, B: 3, A: 50}
	if got != want {
		t.Fatalf("disabled color = %#v, want %#v", got, want)
	}
}

func TestDisabledColorAllowsZeroOpacity(t *testing.T) {
	theme := theme.DefaultTheme()
	theme.DisabledOpacity = 0

	if got := theme.DisabledColor(color.NRGBA{A: 200}); got.A != 0 {
		t.Fatalf("disabled alpha = %d, want 0", got.A)
	}
}

func TestDisabledOpacityIsClamped(t *testing.T) {
	theme := theme.DefaultTheme()
	theme.DisabledOpacity = 2
	if got := theme.DisabledColor(color.NRGBA{A: 200}).A; got != 200 {
		t.Fatalf("disabled alpha = %d, want 200", got)
	}

	theme.DisabledOpacity = -1
	if got := theme.DisabledColor(color.NRGBA{A: 200}).A; got != 0 {
		t.Fatalf("disabled alpha = %d, want 0", got)
	}
}

func TestDefaultThemeSyncsMaterialBridge(t *testing.T) {
	theme := theme.DefaultTheme()
	if theme.Material == nil {
		t.Fatal("missing material bridge")
	}
	if theme.Material.Palette.ContrastBg != theme.Palette.Accent {
		t.Fatalf("material accent = %#v, want %#v", theme.Material.Palette.ContrastBg, theme.Palette.Accent)
	}
	if theme.Material.Palette.Fg != theme.Palette.Foreground {
		t.Fatalf("material foreground = %#v, want %#v", theme.Material.Palette.Fg, theme.Palette.Foreground)
	}
}

func TestDarkThemeDefinesThemedSurfaceAndShadow(t *testing.T) {
	dark := theme.DarkTheme()
	if dark.Palette.Surface == theme.DefaultTheme().Palette.Surface {
		t.Fatal("dark theme did not override surface")
	}
	if dark.Palette.Shadow == theme.DefaultTheme().Palette.Shadow {
		t.Fatal("dark theme did not override shadow")
	}
}
