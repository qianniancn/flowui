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
	if bridge.Face != active.Typography.Typeface {
		t.Fatalf("material face = %q, want %q", bridge.Face, active.Typography.Typeface)
	}
	if bridge.Shaper == nil {
		t.Fatal("missing material text shaper")
	}
}

func TestThemesUseIndependentTextShapers(t *testing.T) {
	first := theme.DefaultTheme()
	second := theme.DefaultTheme()
	if theme.MaterialOf(&first).Shaper == theme.MaterialOf(&second).Shaper {
		t.Fatal("default themes share a text shaper")
	}

	copy := first
	theme.DetachMaterial(&copy)
	if theme.MaterialOf(&first).Shaper == theme.MaterialOf(&copy).Shaper {
		t.Fatal("detached themes share a text shaper")
	}
	theme.SyncMaterialTheme(&copy)
}

func TestSyncMaterialThemeUsesFontConfig(t *testing.T) {
	active := theme.DefaultTheme()
	active.Typography.Typeface = "FlowUI Sans"
	active.Fonts.SystemFonts = false
	theme.SyncMaterialTheme(&active)

	bridge := theme.MaterialOf(&active)
	if bridge.Face != "FlowUI Sans" {
		t.Fatalf("material face = %q, want FlowUI Sans", bridge.Face)
	}
	if bridge.Shaper == nil {
		t.Fatal("font configuration removed the text shaper")
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

func TestWorkbenchThemeFollowsLightAndDarkSurfaces(t *testing.T) {
	light := theme.DefaultTheme()
	dark := theme.DarkTheme()
	if light.Components.Workbench.SidebarWidth <= light.Components.Workbench.SidebarMinWidth || light.Components.Workbench.Density != 1 {
		t.Fatalf("light workbench sizing = %#v", light.Components.Workbench)
	}
	if dark.Components.Workbench.EditorBackground != dark.Palette.Background ||
		dark.Components.Workbench.SidebarBackground != dark.Palette.SurfaceSecondary {
		t.Fatalf("dark workbench surfaces = %#v", dark.Components.Workbench)
	}
	workbench := light.Components.Workbench
	workbench.Density = 1.5
	if workbench.EffectiveDensity() != 1.5 || workbench.Scale(10) != 15 {
		t.Fatalf("scaled workbench density = %v/%v", workbench.EffectiveDensity(), workbench.Scale(10))
	}
}

func TestNodeGraphThemeFollowsLightAndDarkPalettes(t *testing.T) {
	light := theme.DefaultTheme()
	dark := theme.DarkTheme()
	if light.Components.NodeGraph.CanvasBackground != light.Palette.Background ||
		light.Components.NodeGraph.CanvasBorder != light.Palette.Border ||
		light.Components.NodeGraph.CanvasRadius != 8 ||
		light.Components.NodeGraph.GridColor != light.Palette.MutedForeground ||
		light.Components.NodeGraph.GridOpacity != .35 ||
		light.Components.NodeGraph.NodeBackground != light.Palette.FieldBackground ||
		light.Components.NodeGraph.EdgeColor != light.Palette.Accent ||
		light.Components.NodeGraph.SelectedEdgeColor != light.Palette.AccentHover ||
		light.Components.NodeGraph.SelectionBorder != light.Palette.Accent {
		t.Fatalf("light node graph theme = %#v", light.Components.NodeGraph)
	}
	if dark.Components.NodeGraph.CanvasBackground != dark.Palette.Background ||
		dark.Components.NodeGraph.NodeBackground != dark.Palette.FieldBackground ||
		dark.Components.NodeGraph.GridColor != dark.Palette.MutedForeground ||
		dark.Components.NodeGraph.GridOpacity != .35 ||
		dark.Components.NodeGraph.EdgeColor != dark.Palette.Accent ||
		dark.Components.NodeGraph.SelectedEdgeColor != dark.Palette.AccentHover {
		t.Fatalf("dark node graph theme = %#v", dark.Components.NodeGraph)
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
