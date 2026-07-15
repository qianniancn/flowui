package theme_test

import (
	"image/color"
	"testing"
	"time"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestDefaultMotionTheme(t *testing.T) {
	motion := theme.DefaultTheme().Motion
	if !motion.Enabled || motion.DefaultDuration != 200*time.Millisecond || motion.DurationScale != 1 {
		t.Fatalf("default motion theme = %#v", motion)
	}
}

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

func TestMenuHeroUI32SemanticColors(t *testing.T) {
	light := theme.DefaultTheme()
	if light.Palette.Default != (color.NRGBA{R: 0xeb, G: 0xeb, B: 0xec, A: 0xff}) ||
		light.Palette.DefaultForeground != (color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}) ||
		light.Palette.Separator != (color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff}) ||
		light.Components.Menu.DangerColor != (color.NRGBA{R: 0xff, G: 0x38, B: 0x3c, A: 0xff}) {
		t.Fatalf("light Menu semantic colors = palette %#v tokens %#v", light.Palette, light.Components.Menu)
	}
	dark := theme.DarkTheme()
	if dark.Palette.Default != (color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff}) ||
		dark.Palette.Separator != (color.NRGBA{R: 0x21, G: 0x21, B: 0x24, A: 0xff}) ||
		dark.Components.Menu.BackgroundColor != (color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}) ||
		dark.Components.Menu.DangerColor != (color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0xff}) {
		t.Fatalf("dark Menu semantic colors = palette %#v tokens %#v", dark.Palette, dark.Components.Menu)
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

func TestSelectionColorsMatchTheme(t *testing.T) {
	light := color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x50}
	dark := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x58}
	if got := theme.DefaultTheme().Palette.Selection; got != light {
		t.Fatalf("light selection = %#v, want %#v", got, light)
	}
	if got := theme.DarkTheme().Palette.Selection; got != dark {
		t.Fatalf("dark selection = %#v, want %#v", got, dark)
	}
}
