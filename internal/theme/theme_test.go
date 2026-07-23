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

func TestDarkThemeMatchesHeroUI32Palette(t *testing.T) {
	dark := theme.DarkTheme()
	want := theme.Palette{
		Background:                 color.NRGBA{R: 0x06, G: 0x06, B: 0x07, A: 0xff},
		Surface:                    color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		SurfaceForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		SurfaceSecondary:           color.NRGBA{R: 0x23, G: 0x23, B: 0x25, A: 0xff},
		SurfaceSecondaryForeground: color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		SurfaceTertiary:            color.NRGBA{R: 0x26, G: 0x27, B: 0x28, A: 0xff},
		SurfaceTertiaryForeground:  color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		SurfaceHover:               color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff},
		SurfacePressed:             color.NRGBA{R: 0x2e, G: 0x2e, B: 0x31, A: 0xff},
		SurfaceRaised:              color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff},
		Overlay:                    color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		OverlayForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		Foreground:                 color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		MutedForeground:            color.NRGBA{R: 0x9f, G: 0x9f, B: 0xa9, A: 0xff},
		Border:                     color.NRGBA{R: 0x28, G: 0x28, B: 0x2c, A: 0xff},
		Separator:                  color.NRGBA{R: 0x21, G: 0x21, B: 0x24, A: 0xff},
		Default:                    color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff},
		DefaultForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		DefaultHover:               color.NRGBA{R: 0x2e, G: 0x2e, B: 0x31, A: 0xff},
		FieldBackground:            color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		FieldHover:                 color.NRGBA{R: 0x1c, G: 0x1c, B: 0x1e, A: 0xeb},
		FieldForeground:            color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		FieldPlaceholder:           color.NRGBA{R: 0x9f, G: 0x9f, B: 0xa9, A: 0xff},
		FieldFocus:                 color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		Segment:                    color.NRGBA{R: 0x46, G: 0x46, B: 0x4c, A: 0xff},
		SegmentForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		Accent:                     color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
		AccentHover:                color.NRGBA{R: 0x35, G: 0x92, B: 0xf9, A: 0xff},
		AccentPressed:              color.NRGBA{R: 0x00, G: 0x6f, B: 0xd8, A: 0xff},
		AccentForeground:           color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		AccentSoft:                 color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x1f},
		AccentSoftHover:            color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x29},
		AccentSoftForeground:       color.NRGBA{R: 0x61, G: 0xa8, B: 0xfb, A: 0xff},
		Success:                    color.NRGBA{R: 0x17, G: 0xc9, B: 0x64, A: 0xff},
		SuccessForeground:          color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		SuccessSoft:                color.NRGBA{R: 0x17, G: 0xc9, B: 0x64, A: 0x1f},
		SuccessSoftForeground:      color.NRGBA{R: 0x74, G: 0xd8, B: 0x8f, A: 0xff},
		Warning:                    color.NRGBA{R: 0xf7, G: 0xb7, B: 0x50, A: 0xff},
		WarningForeground:          color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		WarningSoft:                color.NRGBA{R: 0xf7, G: 0xb7, B: 0x50, A: 0x1f},
		WarningSoftForeground:      color.NRGBA{R: 0xf9, G: 0xcb, B: 0x86, A: 0xff},
		Danger:                     color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0xff},
		DangerHover:                color.NRGBA{R: 0xe1, G: 0x54, B: 0x51, A: 0xff},
		DangerPressed:              color.NRGBA{R: 0xc6, G: 0x2f, B: 0x33, A: 0xff},
		DangerForeground:           color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		DangerSoft:                 color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0x26},
		DangerSoftHover:            color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0x33},
		DangerSoftForeground:       color.NRGBA{R: 0xeb, G: 0x78, B: 0x72, A: 0xff},
		Focus:                      color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
		Selection:                  color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x58},
		OverlayShadow:              color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x4d},
	}
	if dark.Palette != want {
		t.Fatalf("dark palette = %#v, want %#v", dark.Palette, want)
	}
	wantOverlay := theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{{Blur: 1, Opacity: 1}}}
	if dark.Shadows.Overlay != wantOverlay {
		t.Fatalf("dark overlay shadow = %#v, want %#v", dark.Shadows.Overlay, wantOverlay)
	}
	if dark.Components.Modal.Backdrop != (color.NRGBA{A: 0x99}) || dark.Components.Modal.BlurBackdrop != (color.NRGBA{A: 0x99}) {
		t.Fatalf("dark modal backdrops = %#v %#v", dark.Components.Modal.Backdrop, dark.Components.Modal.BlurBackdrop)
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
