package theme_test

import (
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/qianniancn/flowui/internal/theme"
)

func TestDefaultMotionTheme(t *testing.T) {
	motion := theme.DefaultTheme().Motion
	if !motion.Enabled || motion.DefaultDuration != 200*time.Millisecond || motion.DurationScale != 1 {
		t.Fatalf("default motion theme = %#v", motion)
	}
}

func TestDefaultShadowProfiles(t *testing.T) {
	shadows := theme.DefaultTheme().Shadows
	if shadows.Surface.Layers[1].Blur != 4 || shadows.Overlay.Layers[1].Blur != 22 {
		t.Fatalf("surface shadows = %#v", shadows)
	}
	if shadows.Menu.Layers[2].Blur != 28 || shadows.Control.Layers[2].Blur != 4 || shadows.Checkbox.Layers[1].Blur != 2 || shadows.SwitchThumb.Layers[1].Blur != 8 {
		t.Fatalf("component shadows = %#v", shadows)
	}
}

func TestResolveMotionDuration(t *testing.T) {
	if got := theme.ResolveMotionDuration(theme.MotionTheme{Enabled: true, DurationScale: 0.5}, time.Second); got != 500*time.Millisecond {
		t.Fatalf("scaled duration = %v, want 500ms", got)
	}
	if got := theme.ResolveMotionDuration(theme.MotionTheme{Enabled: false, DurationScale: 1}, time.Second); got != 0 {
		t.Fatalf("disabled duration = %v, want 0", got)
	}
	for _, scale := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		motion := theme.MotionTheme{Enabled: true, DurationScale: scale}
		if theme.MotionEnabled(motion) || theme.ResolveMotionDuration(motion, time.Second) != 0 {
			t.Fatalf("non-finite scale %v enabled motion", scale)
		}
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
	active := theme.DefaultTheme()
	bridge := theme.MaterialOf(&active)
	if bridge == nil {
		t.Fatal("missing material bridge")
	}
	if bridge.Palette.ContrastBg != active.Palette.Accent {
		t.Fatalf("material accent = %#v, want %#v", bridge.Palette.ContrastBg, active.Palette.Accent)
	}
	if bridge.Palette.Fg != active.Palette.Foreground {
		t.Fatalf("material foreground = %#v, want %#v", bridge.Palette.Fg, active.Palette.Foreground)
	}
}

func TestDetachMaterialSeparatesBridge(t *testing.T) {
	source := theme.DefaultTheme()
	copy := source
	theme.DetachMaterial(&copy)
	if theme.MaterialOf(&source) == theme.MaterialOf(&copy) {
		t.Fatal("DetachMaterial retained the source bridge")
	}
}

func TestDarkThemeDefinesThemedSurfaceAndShadow(t *testing.T) {
	dark := theme.DarkTheme()
	if dark.Palette.Surface == theme.DefaultTheme().Palette.Surface {
		t.Fatal("dark theme did not override surface")
	}
	if dark.Palette.OverlayShadow == theme.DefaultTheme().Palette.OverlayShadow {
		t.Fatal("dark theme did not override overlay shadow")
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

func TestPalettePreservesTransparentSemanticColors(t *testing.T) {
	palette := theme.Palette{
		Surface:           color.NRGBA{R: 1, A: 0xff},
		Overlay:           color.NRGBA{},
		Default:           color.NRGBA{},
		FieldBackground:   color.NRGBA{},
		Separator:         color.NRGBA{},
		OverlayForeground: color.NRGBA{},
	}
	if palette.OverlayColor().A != 0 || palette.DefaultColor().A != 0 ||
		palette.FieldBackgroundColor().A != 0 || palette.SeparatorColor().A != 0 ||
		palette.OverlayForegroundColor().A != 0 {
		t.Fatalf("transparent semantic colors were replaced: %#v", palette)
	}
}
