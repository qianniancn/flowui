package runtime

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

func init() {
	// Disable cascade cache for tests to ensure test isolation.
	// In production, contexts are long-lived and per-context caching works well.
	// In tests, each test creates a new short-lived context, which can cause
	// cache key collisions when the same pointer address is reused.
	DisableCascadeCache = true
}

func TestResolveOrderAndThemeTokens(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	scope := style.Style{}.PaddingX(12).Background(style.TokenSurface)
	restore := frame.PushStyle(ctx, scope)
	defer restore()

	defaults := style.Style{}.Padding(4).Background(style.TokenBackground)
	custom := style.Style{}.PaddingX(20).TextColor(style.TokenAccentForeground).BorderColor(style.TokenAccent)
	resolved := Resolve(ctx, testContext(time.Time{}), "button", style.StyleState{}, defaults, style.Style{}, style.Style{}, custom)

	if resolved.Box == nil || resolved.Box.Padding == nil || resolved.Box.Padding.Top != 4 || resolved.Box.Padding.Left != 20 {
		t.Fatalf("padding = %#v, want top 4 and left 20", resolved.Box)
	}
	if got, _ := solidColor(resolved.Paint.Background); got != activeTheme.Palette.Surface {
		t.Fatalf("background = %#v, want surface token", got)
	}
	if got, _ := solidColor(resolved.Text.Color); got != activeTheme.Palette.AccentForeground {
		t.Fatalf("text = %#v, want accent foreground token", got)
	}
	if got, _ := solidColor(resolved.Paint.Border.Color); got != activeTheme.Palette.Accent {
		t.Fatalf("border = %#v, want accent token", got)
	}
}

// TestResolveStaticLayerOrder locks the public cascade contract:
// defaults → inherited text → variant → size → StyleScope → instance.
func TestResolveStaticLayerOrder(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)

	// Inherited text is pushed by layout hosts; inject it like a parent TextDeclaration.
	restoreInherited := frame.PushInheritedStyle(ctx, style.Style{}.
		TextColor(style.RGB(0x111111)).
		FontSize(10))
	defer restoreInherited()

	restoreScope := frame.PushStyle(ctx, style.Style{}.
		PaddingX(12).
		Background(style.RGB(0xaaaaaa)))
	defer restoreScope()

	defaults := style.Style{}.
		Padding(2).
		Background(style.RGB(0x010101)).
		FontSize(8)
	variant := style.Style{}.Background(style.RGB(0x020202))
	size := style.Style{}.Height(40)
	instance := style.Style{}.
		Width(100).
		PaddingLeft(20). // overrides scope PaddingX on the left only
		Background(style.RGB(0x030303)).
		TextColor(style.RGB(0xffffff)) // overrides inherited text

	resolved := ResolveStatic(ctx, style.StyleState{}, defaults, variant, size, instance)

	if resolved.Box == nil || resolved.Box.Width == nil || *resolved.Box.Width != 100 {
		t.Fatalf("instance width = %#v", resolved.Box)
	}
	if resolved.Box.Height == nil || *resolved.Box.Height != 40 {
		t.Fatalf("size height = %#v", resolved.Box)
	}
	// Padding: defaults 2 all sides, scope X=12, instance left=20 → L=20 R=12 T/B=2
	if resolved.Box.Padding == nil ||
		resolved.Box.Padding.Left != 20 || resolved.Box.Padding.Right != 12 ||
		resolved.Box.Padding.Top != 2 || resolved.Box.Padding.Bottom != 2 {
		t.Fatalf("padding = %#v, want L20 R12 T2 B2", resolved.Box.Padding)
	}
	if got, _ := solidColor(resolved.Paint.Background); got != style.RGB(0x030303).Color {
		t.Fatalf("background = %#v, want instance %#v", got, style.RGB(0x030303).Color)
	}
	if resolved.Text == nil || resolved.Text.Color == nil {
		t.Fatal("expected text style")
	}
	if got, _ := solidColor(resolved.Text.Color); got != style.RGB(0xffffff).Color {
		t.Fatalf("text color = %#v, want instance white", got)
	}
	// FontSize from inherited text (10) should win over defaults (8); instance did not set size.
	if resolved.Text.FontSize == nil || *resolved.Text.FontSize != 10 {
		t.Fatalf("font size = %#v, want inherited 10", resolved.Text.FontSize)
	}
}

func TestApplyOutlineOpacityCopiesResolvedStyle(t *testing.T) {
	col := color.NRGBA{R: 1, G: 2, B: 3, A: 200}
	resolved := style.Cascade(style.StyleState{}, style.Style{}.Outline(2, 1, style.SolidColor{Color: col}))
	faded := ApplyOutlineOpacity(resolved, .25)

	got, ok := Color(faded.Paint.Outline.Color)
	if !ok || got.A != 50 {
		t.Fatalf("faded outline = %#v, ok %v", got, ok)
	}
	original, _ := Color(resolved.Paint.Outline.Color)
	if original != col || faded.Paint == resolved.Paint || faded.Paint.Outline == resolved.Paint.Outline {
		t.Fatal("outline opacity mutated the input style")
	}
}

func TestResolveSemanticThemeTokens(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.AccentHover = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.AccentPressed = color.NRGBA{R: 2, A: 0xff}
	activeTheme.Palette.MutedForeground = color.NRGBA{R: 3, A: 0xff}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)

	tests := []struct {
		name  string
		token style.ThemeColor
		want  color.NRGBA
	}{
		{name: "accent hover", token: style.TokenAccentHover, want: activeTheme.Palette.AccentHover},
		{name: "accent pressed", token: style.TokenAccentPressed, want: activeTheme.Palette.AccentPressed},
		{name: "muted foreground", token: style.TokenMutedForeground, want: activeTheme.Palette.MutedForeground},
	}
	for _, test := range tests {
		resolved := ResolveStatic(ctx, style.StyleState{}, style.Style{}.Background(test.token), style.Style{}, style.Style{}, style.Style{})
		if got, _ := solidColor(resolved.Paint.Background); got != test.want {
			t.Errorf("%s = %#v, want %#v", test.name, got, test.want)
		}
	}
}

func TestResolveThemeColorWithAlpha(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Background(style.WithAlpha(style.TokenAccent, 0.2))

	resolved := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(resolved.Paint.Background); got != (color.NRGBA{R: 1, G: 2, B: 3, A: 51}) {
		t.Fatalf("alpha theme color = %#v", got)
	}
}

func TestResolveThemeTokensInsideGradient(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.Danger = color.NRGBA{R: 2, A: 0xff}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	gradient := style.LinearGradient(
		style.ColorStop(0, style.TokenAccent),
		style.ColorStop(1, style.TokenDanger),
	).Angle(90)

	resolved := ResolveStatic(ctx, style.StyleState{}, style.Style{}.Background(gradient), style.Style{}, style.Style{}, style.Style{})
	got := resolved.Paint.Background.(style.StyleGradient)
	first, _ := solidColor(got.Stops[0].Color)
	last, _ := solidColor(got.Stops[1].Color)
	if first != activeTheme.Palette.Accent || last != activeTheme.Palette.Danger || got.AngleDegrees != 90 {
		t.Fatalf("resolved gradient = %#v", got)
	}
}

func TestResolveThemeTokensInsideShadowAndOutline(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.SurfaceShadow = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.Focus = color.NRGBA{R: 2, A: 0xff}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.
		BoxShadow(0, 2, 4, 0, style.TokenSurfaceShadow).
		Outline(1, 2, style.TokenFocus)

	resolved := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	shadow, ok := resolved.Paint.Shadows[0].Color.(style.SolidColor)
	if !ok || shadow.Color != activeTheme.Palette.SurfaceShadow {
		t.Fatalf("shadow color = %#v", resolved.Paint.Shadows[0].Color)
	}
	outline, ok := resolved.Paint.Outline.Color.(style.SolidColor)
	if !ok || outline.Color != activeTheme.Palette.Focus {
		t.Fatalf("outline color = %#v", resolved.Paint.Outline.Color)
	}
}

func TestResolveThemeShadowProfile(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.SurfaceShadow = color.NRGBA{R: 1, A: 200}
	activeTheme.Shadows.Surface = theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
		{OffsetY: 2, Blur: 4, Spread: 1, Opacity: .5},
	}}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Shadow(style.ShadowSurface)

	resolved := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("theme shadows = %#v", resolved.Paint.Shadows)
	}
	shadow := resolved.Paint.Shadows[0]
	col, ok := shadow.Color.(style.SolidColor)
	if !ok || col.Color != (color.NRGBA{R: 1, A: 100}) || shadow.OffsetY != 2 || shadow.Blur != 4 || shadow.Spread != 1 || shadow.Profile != nil {
		t.Fatalf("resolved theme shadow = %#v", shadow)
	}
}

func TestResolveMenuShadowOpacity(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Menu.ShadowOpacity = 0.25
	activeTheme.Components.Menu.ShadowColor = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Shadows.Menu = theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
		{OffsetY: 2, Blur: 4, Opacity: 1},
	}}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Shadow(style.ShadowMenu)

	resolved := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("menu shadows = %#v", resolved.Paint.Shadows)
	}
	shadow, ok := resolved.Paint.Shadows[0].Color.(style.SolidColor)
	if !ok || shadow.Color != (color.NRGBA{R: 1, A: 64}) {
		t.Fatalf("menu shadow color = %#v, want alpha 64", resolved.Paint.Shadows[0].Color)
	}

	activeTheme.Components.Menu.ShadowOpacity = 0
	ctx = frame.New(nil, &activeTheme, locale.LanguageAuto)
	resolved = ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved.Paint.Shadows) != 0 {
		t.Fatalf("disabled menu shadows = %#v, want none", resolved.Paint.Shadows)
	}
}

func TestResolveThemeMetricTokensAndExplicitOverride(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Typography.ControlSize = 17
	activeTheme.Typography.SmallSize = 11
	activeTheme.Shape.ControlRadius = 9
	activeTheme.Spacing.ControlHeight = 41
	activeTheme.Spacing.ControlPaddingX = 13
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.
		Use(style.TokenControlHeight, style.TokenControlRadius, style.TokenControlPaddingX, style.TokenControlFontSize).
		Height(52).
		Part(style.PartLabel, style.Style{}.Use(style.TokenSmallFontSize))

	resolved := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if resolved.Box == nil || *resolved.Box.Height != 52 || resolved.Box.Padding.Left != 13 || *resolved.Paint.Radius != 9 || *resolved.Text.FontSize != 17 {
		t.Fatalf("metric tokens = %#v", resolved)
	}
	label := ResolvePartStatic(ctx, style.PartLabel, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if label.Text == nil || label.Text.FontSize == nil || *label.Text.FontSize != 11 {
		t.Fatalf("part metric token = %#v", label.Text)
	}
}

func TestResolvePartUsesScopeAndThemeTokens(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.Danger = color.NRGBA{R: 2, A: 0xff}
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	restore := frame.PushStyle(ctx, style.Style{}.Part(style.PartLabel, style.Style{}.TextColor(style.TokenAccent)))
	defer restore()
	custom := style.Style{}.
		Part(style.PartLabel, style.Style{}.
			When(style.Invalid, style.Style{}.TextColor(style.TokenDanger)))

	resolved := ResolvePartStatic(ctx, style.PartLabel, style.StyleState{Invalid: true}, style.Style{}, style.Style{}, style.Style{}, custom)
	got, ok := resolved.Text.Color.(style.SolidColor)
	if !ok || got.Color != activeTheme.Palette.Danger {
		t.Fatalf("part color = %#v", resolved.Text.Color)
	}
}

func TestResolvePartStateDoesNotCollideWithUserKey(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx := testContext(time.Unix(1, 0))
	declaration := style.Style{}.
		Part(style.PartLabel, style.Style{}.
			Opacity(1).
			Transition(style.PropOpacity, time.Second))

	ResolvePart(ctx, gtx, "button", style.PartLabel, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, gtx, "button/part/label", style.StyleState{}, style.Style{}.Opacity(1).Transition(style.PropOpacity, time.Second), style.Style{}, style.Style{}, style.Style{})

	if got := frame.StateLen(ctx); got != 2 {
		t.Fatalf("style runtime states = %d, want isolated root and part states", got)
	}
}

func TestSolidColorNilPointerDoesNotPanic(t *testing.T) {
	var nilSolid *style.SolidColor
	if _, ok := solidColor(nilSolid); ok {
		t.Fatal("nil *SolidColor should not report a color")
	}
}

func TestApplyPartTransitionsEmptyKeySoftSnaps(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx := testContext(time.Unix(1, 0))
	opacity := float32(0.5)
	resolved := style.ResolvedStyle{
		Paint: &style.PaintStyle{Opacity: &opacity},
		Transitions: []style.Transition{{
			Property: style.PropOpacity,
			Duration: time.Second,
		}},
	}
	out := ApplyPartTransitions(ctx, gtx, "", style.PartLabel, resolved)
	if out.Paint == nil || out.Paint.Opacity == nil || *out.Paint.Opacity != 0.5 {
		t.Fatalf("empty part key should soft-snap, got %#v", out.Paint)
	}
	if got := frame.StateLen(ctx); got != 0 {
		t.Fatalf("empty part key retained %d states, want 0", got)
	}
}

func TestApplyTransitionsEmptyKeySnapsWithoutState(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	declaration := style.Style{}.
		Opacity(0.5).
		Transition(style.PropOpacity, time.Second)

	resolved := Resolve(ctx, testContext(time.Unix(1, 0)), "", style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0.5 {
		t.Fatalf("empty key should soft-snap to target opacity, got %#v", resolved.Paint)
	}
	if got := frame.StateLen(ctx); got != 0 {
		t.Fatalf("empty key retained %d style states, want 0", got)
	}
}

func TestResolveTransitionsConditionalBackground(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	from := style.RGB(0x102030)
	to := style.RGB(0x8090a0)
	declaration := style.Style{}.
		Background(from).
		Transition(style.PropBackgroundColor, 100*time.Millisecond).
		When(style.Hovered, style.Style{}.Background(to))

	start := time.Unix(1, 0)

	normal := Resolve(ctx, testContext(start), "button", style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(normal.Paint.Background); got != from.Color {
		t.Fatalf("normal = %#v, want %#v", got, from.Color)
	}
	begin := Resolve(ctx, testContext(start), "button", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(begin.Paint.Background); got != from.Color {
		t.Fatalf("transition start = %#v, want %#v", got, from.Color)
	}
	middle := Resolve(ctx, testContext(start.Add(50*time.Millisecond)), "button", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	got, _ := solidColor(middle.Paint.Background)
	if got == from.Color || got == to.Color {
		t.Fatalf("transition midpoint = %#v", got)
	}
}

func TestResolveWithoutTransitionsIsStateless(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	frame.BeginFrame(ctx)
	Resolve(ctx, testContext(time.Unix(1, 0)), "static", style.StyleState{}, style.Style{}.Background(style.RGB(0x102030)), style.Style{}, style.Style{}, style.Style{})
	if got := frame.StateLen(ctx); got != 0 {
		t.Fatalf("static style retained %d runtime states", got)
	}
	frame.EndFrame(ctx)
}

func TestResolveTransitionsExplicitZeroValues(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	declaration := style.Style{}.
		Opacity(1).
		Radius(unit.Dp(8)).
		Transition(style.PropOpacity, time.Second).
		Transition(style.PropRadius, time.Second).
		When(style.Disabled, style.Style{}.Opacity(0).Radius(0))

	start := time.Unix(1, 0)
	Resolve(ctx, testContext(start), "button", style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, testContext(start), "button", style.StyleState{Disabled: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	resolved := Resolve(ctx, testContext(start.Add(time.Second)), "button", style.StyleState{Disabled: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if *resolved.Paint.Opacity != 0 || *resolved.Paint.Radius != 0 {
		t.Fatalf("resolved zero values = opacity %v radius %v", *resolved.Paint.Opacity, *resolved.Paint.Radius)
	}
}

func TestResolveTransitionDelayHoldsPreviousValue(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	from := style.RGB(0x102030)
	to := style.RGB(0x8090a0)
	declaration := style.Style{}.
		Background(from).
		Transition(style.PropBackgroundColor, 100*time.Millisecond, style.TransitionDelay(50*time.Millisecond)).
		When(style.Hovered, style.Style{}.Background(to))

	start := time.Unix(1, 0)

	Resolve(ctx, testContext(start), "delayed", style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, testContext(start), "delayed", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	delayed := Resolve(ctx, testContext(start.Add(25*time.Millisecond)), "delayed", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(delayed.Paint.Background); got != from.Color {
		t.Fatalf("background during delay = %#v, want %#v", got, from.Color)
	}
	afterDelay := Resolve(ctx, testContext(start.Add(100*time.Millisecond)), "delayed", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	got, _ := solidColor(afterDelay.Paint.Background)
	if got == from.Color || got == to.Color {
		t.Fatalf("background after delay = %#v, want a transition midpoint", got)
	}
}

func TestResolveTransitionsCompleteTransform(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	declaration := style.Style{}.
		Translate(0, 0).
		Rotate(0).
		Transition(style.PropTransform, time.Second, style.TransitionEase(func(value float32) float32 { return value })).
		When(style.Hovered, style.Style{}.Translate(10, 20).Rotate(1))

	start := time.Unix(1, 0)

	Resolve(ctx, testContext(start), "transform", style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, testContext(start), "transform", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	middle := Resolve(ctx, testContext(start.Add(500*time.Millisecond)), "transform", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if middle.Trans == nil || middle.Trans.TranslateX == nil || middle.Trans.TranslateY == nil || middle.Trans.Rotate == nil {
		t.Fatalf("transform is nil or incomplete: %#v", middle.Trans)
	}
	tx, ty, rot := *middle.Trans.TranslateX, *middle.Trans.TranslateY, *middle.Trans.Rotate
	if tx != 5 || ty != 10 || rot != .5 {
		t.Fatalf("transform midpoint = TX:%v TY:%v Rotate:%v, want TX:5 TY:10 Rotate:0.5", tx, ty, rot)
	}
}

func TestApplyOpacity(t *testing.T) {
	value := color.NRGBA{R: 1, G: 2, B: 3, A: 200}
	if got := ApplyOpacity(value, nil); got != value {
		t.Fatalf("nil opacity = %#v, want %#v", got, value)
	}
	opacity := float32(.25)
	if got := ApplyOpacity(value, &opacity); got.A != 50 {
		t.Fatalf("quarter opacity alpha = %d, want 50", got.A)
	}
	opacity = 2
	if got := ApplyOpacity(value, &opacity); got.A != value.A {
		t.Fatalf("clamped opacity alpha = %d, want %d", got.A, value.A)
	}
}

func testContext(now time.Time) layout.Context {
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         new(op.Ops),
		Now:         now,
	}
}

// Benchmarks for #5 and #6 optimizations

// BenchmarkResolveStatic_SimpleLayers benchmarks #5 - cascade without tokens/conditions
func BenchmarkResolveStatic_SimpleLayers(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)

	// Simple layers without tokens (should benefit from fast path)
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.Opacity(0.9)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveStatic(ctx, style.StyleState{}, defaults, variant, size, custom)
	}
}

// BenchmarkResolveStatic_WithTokens benchmarks the full expansion path with theme tokens
func BenchmarkResolveStatic_WithTokens(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)

	// Layers with color tokens (must expand fully)
	defaults := style.Style{}.Padding(8).Background(style.TokenSurface)
	variant := style.Style{}.PaddingX(16).TextColor(style.TokenForeground)
	size := style.Style{}.Height(36).BorderColor(style.TokenBorder)
	custom := style.Style{}.Opacity(0.9)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveStatic(ctx, style.StyleState{}, defaults, variant, size, custom)
	}
}

// BenchmarkResolve_CascadeLayers benchmarks the full cascade resolution with transitions
func BenchmarkResolve_CascadeLayers(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Now())

	// Multi-layer cascade (typical button scenario)
	defaults := style.Style{}.Padding(8).Background(style.TokenSurface)
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	instance := style.Style{}.TextColor(style.TokenAccentForeground)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resolve(ctx, gtx, "button", style.StyleState{}, defaults, variant, size, instance)
	}
}

// BenchmarkResolve_WithHover benchmarks resolution with hover state (triggers animations)
func BenchmarkResolve_WithHover(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Now())

	declaration := style.Style{}.
		Background(style.TokenSurface).
		When(style.Hovered, style.Style{}.Background(style.TokenSurfaceHover))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resolve(ctx, gtx, "button", style.StyleState{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	}
}

// BenchmarkActiveStyles_ReadOnly benchmarks #6 - zero-copy accessor
func BenchmarkActiveStyles_ReadOnly(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)

	// Push some styles to simulate real usage
	restore := frame.PushStyle(ctx, style.Style{}.Padding(12).Background(style.TokenSurface))
	defer restore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = frame.ActiveStylesReadOnly(ctx)
	}
}

// BenchmarkActiveInheritedStyles_ReadOnly benchmarks #6 - zero-copy inherited accessor
func BenchmarkActiveInheritedStyles_ReadOnly(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)

	// Push inherited text styles
	restore := frame.PushInheritedStyle(ctx, style.Style{}.FontSize(14).TextColor(style.TokenForeground))
	defer restore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = frame.ActiveInheritedStylesReadOnly(ctx)
	}
}
