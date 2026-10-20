package computed

import (
	"image"
	"image/color"
	"reflect"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func init() {
	// Disable cascade cache for tests to ensure test isolation.
	// In production, contexts are long-lived and per-context caching works well.
	// In tests, each test creates a new short-lived context, which can cause
	// cache key collisions when the same pointer address is reused.
	DisableCascadeCache = true
}

func TestResolveOrderAndThemeValues(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	scope := style.Style{}.PaddingX(12).Background(style.Solid(activeTheme.Palette.Surface))
	restore := core.PushStyle(ctx, scope)
	defer restore()

	defaults := style.Style{}.Padding(4).Background(style.Solid(activeTheme.Palette.Background))
	custom := style.Style{}.PaddingX(20).TextColor(style.Solid(activeTheme.Palette.AccentForeground)).BorderColor(style.Solid(activeTheme.Palette.Accent))
	resolved := Resolve(ctx, testContext(time.Time{}), "button", style.State{}, defaults, style.Style{}, style.Style{}, custom)

	if resolved.Box == nil || resolved.Box.Padding == nil || resolved.Box.Padding.Top != 4 || resolved.Box.Padding.Left != 20 {
		t.Fatalf("padding = %#v, want top 4 and left 20", resolved.Box)
	}
	if got, _ := solidColor(resolved.Paint.Background); got != activeTheme.Palette.Surface {
		t.Fatalf("background = %#v, want surface value", got)
	}
	if got, _ := solidColor(resolved.Text.Color); got != activeTheme.Palette.AccentForeground {
		t.Fatalf("text = %#v, want accent foreground value", got)
	}
	if got, _ := solidColor(resolved.Paint.Border.Color); got != activeTheme.Palette.Accent {
		t.Fatalf("border = %#v, want accent value", got)
	}
}

func TestResolvePartsMatchesIndividualResolution(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Time{})
	state := style.State{Hovered: true, Focused: true, Invalid: true}
	defaults := style.Style{}.
		TextColor(style.RGB(0x112233)).
		Part(style.PartContent, style.Style{}.Padding(4).Background(style.RGB(0x223344))).
		Part(style.PartIndicator, style.Style{}.Width(2)).
		On(style.Hovered, style.Style{}.Part(style.PartContent, style.Style{}.Padding(8)))
	variant := style.Style{}.Part(style.PartPrefix, style.Style{}.Width(5))
	size := style.Style{}.Part(style.PartSuffix, style.Style{}.Width(6))
	custom := style.Style{}.Part(style.PartIndicator, style.Style{}.Width(3))

	restore := core.PushInheritedStyle(ctx, style.Style{}.TextColor(style.RGB(0xabcdef)))
	defer restore()
	parts := []style.Part{
		style.PartRoot,
		style.PartContent,
		style.PartPrefix,
		style.PartSuffix,
		style.PartIndicator,
	}
	batched := ResolveParts(ctx, gtx, "batch", state, parts, defaults, variant, size, custom)
	want := []style.ResolvedStyle{
		Resolve(ctx, gtx, "individual-root", state, defaults, variant, size, custom),
		ResolvePart(ctx, gtx, "individual-content", style.PartContent, state, defaults, variant, size, custom),
		ResolvePart(ctx, gtx, "individual-prefix", style.PartPrefix, state, defaults, variant, size, custom),
		ResolvePart(ctx, gtx, "individual-suffix", style.PartSuffix, state, defaults, variant, size, custom),
		ResolvePart(ctx, gtx, "individual-indicator", style.PartIndicator, state, defaults, variant, size, custom),
	}
	if !reflect.DeepEqual(batched, want) {
		t.Fatalf("batch resolution differs from individual resolution:\n got  %#v\n want %#v", batched, want)
	}
}

func TestResolvePartsStaticMatchesIndividualResolution(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	state := style.State{Hovered: true, Disabled: true}
	defaults := style.Style{}.
		TextColor(style.RGB(0x112233)).
		Part(style.PartContent, style.Style{}.Padding(4).Background(style.RGB(0x223344))).
		On(style.Hovered, style.Style{}.Part(style.PartContent, style.Style{}.Padding(8)))
	variant := style.Style{}.Part(style.PartPrefix, style.Style{}.Width(5))
	size := style.Style{}.Part(style.PartSuffix, style.Style{}.Width(6))
	custom := style.Style{}.Part(style.PartIndicator, style.Style{}.Width(3))

	parts := []style.Part{
		style.PartRoot,
		style.PartContent,
		style.PartPrefix,
		style.PartSuffix,
		style.PartIndicator,
	}
	batched := ResolvePartsStatic(ctx, state, parts, defaults, variant, size, custom)
	var into [5]style.ResolvedStyle
	ResolvePartsStaticInto(ctx, state, parts, into[:], defaults, variant, size, custom)
	want := []style.ResolvedStyle{
		ResolveStatic(ctx, state, defaults, variant, size, custom),
		ResolvePartStatic(ctx, style.PartContent, state, defaults, variant, size, custom),
		ResolvePartStatic(ctx, style.PartPrefix, state, defaults, variant, size, custom),
		ResolvePartStatic(ctx, style.PartSuffix, state, defaults, variant, size, custom),
		ResolvePartStatic(ctx, style.PartIndicator, state, defaults, variant, size, custom),
	}
	if !reflect.DeepEqual(batched, want) {
		t.Fatalf("static batch resolution differs from individual resolution:\n got  %#v\n want %#v", batched, want)
	}
	if !reflect.DeepEqual(into[:], want) {
		t.Fatalf("static into resolution differs from individual resolution:\n got  %#v\n want %#v", into, want)
	}
}

func TestApplyTransitionsWithSharedKeyDoesNotMutateSharedStyle(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.
		Part(style.PartContent, style.Style{}.
			Background(style.RGB(0x102030)).
			Transition(style.PropBackgroundColor, time.Second))
	static := ResolvePartStaticShared(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if static.Paint == nil || len(static.Transitions) == 0 {
		t.Fatal("shared resolve did not retain the transition-bearing paint")
	}
	if got, ok := static.Paint.Background.(style.SolidColor); !ok || got.Color != (color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}) {
		t.Fatalf("shared static background = %#v, want target color", static.Paint.Background)
	}

	animated := ApplyTransitionsWithSharedKey(ctx, testContext(time.Now()), "shared-transition", static)
	if animated.Paint == static.Paint {
		t.Fatal("active shared transition mutated the cache-owned paint")
	}
	if got, ok := static.Paint.Background.(style.SolidColor); !ok || got.Color != (color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}) {
		t.Fatalf("shared static background changed after animation = %#v", static.Paint.Background)
	}
}

func TestResolvePartsPreservesRootAndPartTransitions(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Time{})
	declaration := style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, 120*time.Millisecond).
		Part(style.PartContent, style.Style{}.
			TextColor(style.RGB(0x112233)).
			Transition(style.PropTextColor, 120*time.Millisecond))

	resolved := ResolveParts(ctx, gtx, "transition-group", style.State{}, []style.Part{style.PartRoot, style.PartContent}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved) != 2 {
		t.Fatalf("resolved parts = %d, want 2", len(resolved))
	}
	if len(resolved[0].Transitions) == 0 || len(resolved[1].Transitions) == 0 {
		t.Fatalf("batch transitions = root %d/content %d, want both non-empty", len(resolved[0].Transitions), len(resolved[1].Transitions))
	}
}

// The theme hash is a cache key, so two themes that differ only in the gradient
// must not share one: the cached declaration would paint the other's brand.
func TestThemeHashSeparatesTheGradient(t *testing.T) {
	flat := theme.NeutralTheme()
	flat.UseSemantic(theme.SemanticFromPalette(flat.Palette))

	stop := func(r uint8) *theme.Gradient {
		return &theme.Gradient{
			AngleDegrees: 135,
			Stops: []theme.GradientStop{
				{Offset: 0, Color: color.NRGBA{R: r, A: 0xff}},
				{Offset: 1, Color: color.NRGBA{R: 0xff - r, A: 0xff}},
			},
		}
	}
	withGradient := theme.NeutralTheme()
	semanticWithGradient := theme.SemanticFromPalette(withGradient.Palette)
	semanticWithGradient.PrimaryGradient = stop(0x11)
	withGradient.UseSemantic(semanticWithGradient)

	otherGradient := theme.NeutralTheme()
	semanticWithOtherGradient := theme.SemanticFromPalette(otherGradient.Palette)
	semanticWithOtherGradient.PrimaryGradient = stop(0x22)
	otherGradient.UseSemantic(semanticWithOtherGradient)

	flatHash := computeThemeHash(&flat)
	if computeThemeHash(&withGradient) == flatHash {
		t.Error("a gradient went unnoticed by the theme hash")
	}
	if computeThemeHash(&otherGradient) == computeThemeHash(&withGradient) {
		t.Error("a changed gradient went unnoticed by the theme hash")
	}
}

// TestResolveStaticLayerOrder locks the cascade contract:
// defaults -> inherited styles -> variant -> size -> active scope -> instance.
func TestResolveStaticLayerOrder(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)

	// Inherited text is pushed by layout hosts; inject it like a parent TextDeclaration.
	restoreInherited := core.PushInheritedStyle(ctx, style.Style{}.
		TextColor(style.RGB(0x111111)).
		FontSize(10))
	defer restoreInherited()

	restoreScope := core.PushStyle(ctx, style.Style{}.
		PaddingX(12).
		Background(style.RGB(0xaaaaaa)))
	defer restoreScope()

	defaults := style.Style{}.
		Padding(2).
		Background(style.RGB(0x010101)).
		FontSize(8)
	variant := style.Style{}.Background(style.RGB(0x020202))
	size := style.Style{}.
		Height(40).
		Background(style.RGB(0x040404))
	instance := style.Style{}.
		Width(100).
		PaddingLeft(20). // overrides scope PaddingX on the left only
		Background(style.RGB(0x030303)).
		TextColor(style.RGB(0xffffff)) // overrides inherited text

	scoped := ResolveStatic(ctx, style.State{}, defaults, variant, size, style.Style{})
	if got, _ := solidColor(scoped.Paint.Background); got != style.RGB(0xaaaaaa).Color {
		t.Fatalf("scope background = %#v, want scope to override variant and size", got)
	}

	resolved := ResolveStatic(ctx, style.State{}, defaults, variant, size, instance)

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
	resolved := style.Cascade(style.State{}, style.Style{}.Outline(2, 1, style.SolidColor{Color: col}))
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

func TestResolveColorWithAlpha(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	activeTheme = paletteTheme(activeTheme)
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Background(style.WithAlpha(activeTheme.Palette.Accent, 0.2))

	resolved := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(resolved.Paint.Background); got != (color.NRGBA{R: 1, G: 2, B: 3, A: 51}) {
		t.Fatalf("alpha color = %#v", got)
	}
}

func TestResolveColorsInsideGradient(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.Danger = color.NRGBA{R: 2, A: 0xff}
	activeTheme = paletteTheme(activeTheme)
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gradient := style.LinearGradient(
		style.ColorStop(0, style.Solid(activeTheme.Palette.Accent)),
		style.ColorStop(1, style.Solid(activeTheme.Palette.Danger)),
	).Angle(90)

	resolved := ResolveStatic(ctx, style.State{}, style.Style{}.Background(gradient), style.Style{}, style.Style{}, style.Style{})
	got := resolved.Paint.Background.(style.Gradient)
	first, _ := solidColor(got.Stops[0].Color)
	last, _ := solidColor(got.Stops[1].Color)
	if first != activeTheme.Palette.Accent || last != activeTheme.Palette.Danger || got.AngleDegrees != 90 {
		t.Fatalf("resolved gradient = %#v", got)
	}
}

func TestResolveColorsInsideShadowAndOutline(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Palette.SurfaceShadow = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.Focus = color.NRGBA{R: 2, A: 0xff}
	activeTheme = paletteTheme(activeTheme)
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.
		BoxShadow(0, 2, 4, 0, style.Solid(activeTheme.Palette.SurfaceShadow)).
		Outline(1, 2, style.Solid(activeTheme.Palette.Focus))

	resolved := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
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
	activeTheme := theme.NeutralTheme()
	activeTheme.Palette.SurfaceShadow = color.NRGBA{R: 1, A: 200}
	activeTheme = paletteTheme(activeTheme)
	activeTheme.Shadows.Surface = theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
		{OffsetY: 2, Blur: 4, Spread: 1, Opacity: .5},
	}}
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Shadow(style.ShadowSurface)

	resolved := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("theme shadows = %#v", resolved.Paint.Shadows)
	}
	shadow := resolved.Paint.Shadows[0]
	col, ok := shadow.Color.(style.SolidColor)
	if !ok || col.Color != (color.NRGBA{R: 1, A: 100}) || shadow.OffsetY != 2 || shadow.Blur != 4 || shadow.Spread != 1 || shadow.Profile != nil {
		t.Fatalf("resolved theme shadow = %#v", shadow)
	}
}

func TestResolveThemeShadowWithoutThemeStaysNeutral(t *testing.T) {
	ctx := core.New(nil, nil, locale.LanguageAuto)
	declaration := style.Style{}.Shadow(style.ShadowSurface)

	resolved := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if resolved.Paint == nil {
		t.Fatal("resolved paint is nil")
	}
	if len(resolved.Paint.Shadows) != 0 {
		t.Fatalf("theme-less shadow profile resolved to %#v, want no visual shadow", resolved.Paint.Shadows)
	}
}

func TestResolveMenuShadowOpacity(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Components.Menu.ShadowOpacity = 0.25
	activeTheme.Components.Menu.ShadowColor = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Shadows.Menu = theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
		{OffsetY: 2, Blur: 4, Opacity: 1},
	}}
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Shadow(style.ShadowMenu)

	resolved := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("menu shadows = %#v", resolved.Paint.Shadows)
	}
	shadow, ok := resolved.Paint.Shadows[0].Color.(style.SolidColor)
	if !ok || shadow.Color != (color.NRGBA{R: 1, A: 64}) {
		t.Fatalf("menu shadow color = %#v, want alpha 64", resolved.Paint.Shadows[0].Color)
	}

	activeTheme.Components.Menu.ShadowOpacity = 0
	ctx = core.New(nil, &activeTheme, locale.LanguageAuto)
	resolved = ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if len(resolved.Paint.Shadows) != 0 {
		t.Fatalf("disabled menu shadows = %#v, want none", resolved.Paint.Shadows)
	}
}

func TestResolvePartUsesScopeAndThemeValues(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.Danger = color.NRGBA{R: 2, A: 0xff}
	activeTheme = paletteTheme(activeTheme)
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	restore := core.PushStyle(ctx, style.Style{}.Part(style.PartLabel, style.Style{}.TextColor(style.Solid(activeTheme.Palette.Accent))))
	defer restore()
	custom := style.Style{}.
		Part(style.PartLabel, style.Style{}.
			On(style.Invalid, style.Style{}.TextColor(style.Solid(activeTheme.Palette.Danger))))

	resolved := ResolvePartStatic(ctx, style.PartLabel, style.State{Invalid: true}, style.Style{}, style.Style{}, style.Style{}, custom)
	got, ok := resolved.Text.Color.(style.SolidColor)
	if !ok || got.Color != activeTheme.Palette.Danger {
		t.Fatalf("part color = %#v", resolved.Text.Color)
	}
}

func TestResolvePartStateDoesNotCollideWithUserKey(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Unix(1, 0))
	declaration := style.Style{}.
		Part(style.PartLabel, style.Style{}.
			Opacity(1).
			Transition(style.PropOpacity, time.Second))

	ResolvePart(ctx, gtx, "button", style.PartLabel, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, gtx, "button/part/label", style.State{}, style.Style{}.Opacity(1).Transition(style.PropOpacity, time.Second), style.Style{}, style.Style{}, style.Style{})

	if got := core.StateLen(ctx); got != 2 {
		t.Fatalf("state slots = %d, want isolated root and part states", got)
	}
}

func TestTransitionsWithoutThemeStayAtTargetWithoutState(t *testing.T) {
	ctx := core.New(nil, nil, locale.LanguageAuto)
	gtx := testContext(time.Unix(1, 0))
	declaration := style.Style{}.
		Opacity(0.4).
		Transition(style.PropOpacity, time.Second)

	resolved := Resolve(ctx, gtx, "button", style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0.4 {
		t.Fatalf("theme-less transition result = %#v, want target opacity", resolved.Paint)
	}
	if got := core.StateLen(ctx); got != 0 {
		t.Fatalf("theme-less transition retained %d animation states, want 0", got)
	}
}

// TestDimMatchesDisabledColorArithmetic locks the equivalence that lets a
// rule table replace the fluent disabled declaration field by field:
// Dim(DisabledOpacityValue()) must produce the very byte DisabledColor
// produces, truncation included.
func TestDimMatchesDisabledColorArithmetic(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.DisabledOpacity = 0.3
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	base := style.RGB(0x336699)
	declaration := style.Style{}.
		Background(base).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.Dim(activeTheme.DisabledOpacityValue()),
		})

	resolved := ResolveStatic(ctx, style.State{Disabled: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	want := activeTheme.DisabledColor(base.Color)
	if got, _ := solidColor(resolved.Paint.Background); got != want {
		t.Fatalf("dimmed background = %#v, want %#v", got, want)
	}
}

func TestSolidColorNilPointerDoesNotPanic(t *testing.T) {
	var nilSolid *style.SolidColor
	if _, ok := solidColor(nilSolid); ok {
		t.Fatal("nil *SolidColor should not report a color")
	}
}

func TestApplyPartTransitionsEmptyKeySoftSnaps(t *testing.T) {
	ctx := core.New(nil, nil, locale.LanguageAuto)
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
	if got := core.StateLen(ctx); got != 0 {
		t.Fatalf("empty part key retained %d states, want 0", got)
	}
}

func TestApplyTransitionsEmptyKeySnapsWithoutState(t *testing.T) {
	ctx := core.New(nil, nil, locale.LanguageAuto)
	declaration := style.Style{}.
		Opacity(0.5).
		Transition(style.PropOpacity, time.Second)

	resolved := Resolve(ctx, testContext(time.Unix(1, 0)), "", style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0.5 {
		t.Fatalf("empty key should soft-snap to target opacity, got %#v", resolved.Paint)
	}
	if got := core.StateLen(ctx); got != 0 {
		t.Fatalf("empty key retained %d style states, want 0", got)
	}
}

func TestApplyTransitionsWithKeyReusesSettledStyle(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Unix(1, 0))
	value := float32(0.5)
	declaration := style.ResolvedStyle{
		Paint: &style.PaintStyle{Opacity: &value},
		Transitions: []style.Transition{{
			Property: style.PropOpacity,
			Duration: time.Second,
		}},
	}

	first := ApplyTransitionsWithKey(ctx, gtx, "item-transition", declaration)
	secondValue := value
	secondInput := style.ResolvedStyle{
		Paint:       &style.PaintStyle{Opacity: &secondValue},
		Transitions: declaration.Transitions,
	}
	second := ApplyTransitionsWithKey(ctx, gtx, "item-transition", secondInput)
	if second.Paint != secondInput.Paint {
		t.Fatal("settled animation copied the resolved paint")
	}
	if first.Paint == nil || second.Paint == nil || *first.Paint.Opacity != value || *second.Paint.Opacity != value {
		t.Fatalf("settled animation changed target: first=%#v second=%#v", first.Paint, second.Paint)
	}
}

func TestApplyTransitionsWithKeyRetainsStateAcrossFrames(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Unix(1, 0))
	value := float32(0.5)
	input := func() style.ResolvedStyle {
		return style.ResolvedStyle{
			Paint: &style.PaintStyle{Opacity: &value},
			Transitions: []style.Transition{{
				Property: style.PropOpacity,
				Duration: time.Second,
			}},
		}
	}

	core.BeginFrame(ctx)
	firstInput := input()
	first := ApplyTransitionsWithKey(ctx, gtx, "retained-animation", firstInput)
	firstState, ok := core.PeekState[engineState](ctx, "retained-animation", stateSlot)
	if !ok {
		t.Fatal("first frame did not create animation state")
	}
	core.EndFrame(ctx)

	core.BeginFrame(ctx)
	secondInput := input()
	second := ApplyTransitionsWithKey(ctx, gtx, "retained-animation", secondInput)
	secondState, ok := core.PeekState[engineState](ctx, "retained-animation", stateSlot)
	if !ok || secondState != firstState {
		t.Fatal("settled keyed animation state was recreated across frames")
	}
	if first.Paint == nil || second.Paint != secondInput.Paint {
		t.Fatal("settled keyed animation did not preserve the shared resolved style")
	}
	core.EndFrame(ctx)
}

func TestResolveTransitionsConditionalBackground(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	from := style.RGB(0x102030)
	to := style.RGB(0x8090a0)
	declaration := style.Style{}.
		Background(from).
		Transition(style.PropBackgroundColor, 100*time.Millisecond).
		On(style.Hovered, style.Style{}.Background(to))

	start := time.Unix(1, 0)

	normal := Resolve(ctx, testContext(start), "button", style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(normal.Paint.Background); got != from.Color {
		t.Fatalf("normal = %#v, want %#v", got, from.Color)
	}
	begin := Resolve(ctx, testContext(start), "button", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(begin.Paint.Background); got != from.Color {
		t.Fatalf("transition start = %#v, want %#v", got, from.Color)
	}
	middle := Resolve(ctx, testContext(start.Add(50*time.Millisecond)), "button", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	got, _ := solidColor(middle.Paint.Background)
	if got == from.Color || got == to.Color {
		t.Fatalf("transition midpoint = %#v", got)
	}
}

func TestResolveWithoutTransitionsIsStateless(t *testing.T) {
	ctx := core.New(nil, nil, locale.LanguageAuto)
	core.BeginFrame(ctx)
	Resolve(ctx, testContext(time.Unix(1, 0)), "static", style.State{}, style.Style{}.Background(style.RGB(0x102030)), style.Style{}, style.Style{}, style.Style{})
	if got := core.StateLen(ctx); got != 0 {
		t.Fatalf("static style retained %d runtime states", got)
	}
	core.EndFrame(ctx)
}

func TestResolveTransitionsExplicitZeroValues(t *testing.T) {
	ctx := core.New(nil, nil, locale.LanguageAuto)
	declaration := style.Style{}.
		Opacity(1).
		Radius(unit.Dp(8)).
		Transition(style.PropOpacity, time.Second).
		Transition(style.PropRadius, time.Second).
		On(style.Disabled, style.Style{}.Opacity(0).Radius(0))

	start := time.Unix(1, 0)
	Resolve(ctx, testContext(start), "button", style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, testContext(start), "button", style.State{Disabled: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	resolved := Resolve(ctx, testContext(start.Add(time.Second)), "button", style.State{Disabled: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if *resolved.Paint.Opacity != 0 || *resolved.Paint.Radius != 0 {
		t.Fatalf("resolved zero values = opacity %v radius %v", *resolved.Paint.Opacity, *resolved.Paint.Radius)
	}
}

func TestResolveTransitionDelayHoldsPreviousValue(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	from := style.RGB(0x102030)
	to := style.RGB(0x8090a0)
	declaration := style.Style{}.
		Background(from).
		Transition(style.PropBackgroundColor, 100*time.Millisecond, style.TransitionDelay(50*time.Millisecond)).
		On(style.Hovered, style.Style{}.Background(to))

	start := time.Unix(1, 0)

	Resolve(ctx, testContext(start), "delayed", style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, testContext(start), "delayed", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	delayed := Resolve(ctx, testContext(start.Add(25*time.Millisecond)), "delayed", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	if got, _ := solidColor(delayed.Paint.Background); got != from.Color {
		t.Fatalf("background during delay = %#v, want %#v", got, from.Color)
	}
	afterDelay := Resolve(ctx, testContext(start.Add(100*time.Millisecond)), "delayed", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	got, _ := solidColor(afterDelay.Paint.Background)
	if got == from.Color || got == to.Color {
		t.Fatalf("background after delay = %#v, want a transition midpoint", got)
	}
}

func TestResolveTransitionsCompleteTransform(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.
		Translate(0, 0).
		Rotate(0).
		Transition(style.PropTransform, time.Second, style.TransitionEase(func(value float32) float32 { return value })).
		On(style.Hovered, style.Style{}.Translate(10, 20).Rotate(1))

	start := time.Unix(1, 0)

	Resolve(ctx, testContext(start), "transform", style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	Resolve(ctx, testContext(start), "transform", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	middle := Resolve(ctx, testContext(start.Add(500*time.Millisecond)), "transform", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
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

// BenchmarkResolveStatic_SimpleLayers benchmarks #5 - the cascade fast path without conditions
func BenchmarkResolveStatic_SimpleLayers(b *testing.B) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)

	// Simple layers without tokens (should benefit from fast path)
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.Opacity(0.9)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveStatic(ctx, style.State{}, defaults, variant, size, custom)
	}
}

// BenchmarkResolve_CascadeLayers benchmarks the full cascade resolution with transitions
func BenchmarkResolve_CascadeLayers(b *testing.B) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Now())

	// Multi-layer cascade (typical button scenario)
	defaults := style.Style{}.Padding(8).Background(style.Solid(activeTheme.Palette.Surface))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	instance := style.Style{}.TextColor(style.Solid(activeTheme.Palette.AccentForeground))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resolve(ctx, gtx, "button", style.State{}, defaults, variant, size, instance)
	}
}

// BenchmarkResolve_WithHover benchmarks resolution with hover state (triggers animations)
func BenchmarkResolve_WithHover(b *testing.B) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testContext(time.Now())

	declaration := style.Style{}.
		Background(style.Solid(activeTheme.Palette.Surface)).
		On(style.Hovered, style.Style{}.Background(style.Solid(activeTheme.Palette.SurfaceHover)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resolve(ctx, gtx, "button", style.State{Hovered: true}, declaration, style.Style{}, style.Style{}, style.Style{})
	}
}

// BenchmarkActiveStyles_ReadOnly benchmarks #6 - zero-copy accessor
func BenchmarkActiveStyles_ReadOnly(b *testing.B) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)

	// Push some styles to simulate real usage
	restore := core.PushStyle(ctx, style.Style{}.Padding(12).Background(style.Solid(activeTheme.Palette.Surface)))
	defer restore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = core.ActiveStylesReadOnly(ctx)
	}
}

// BenchmarkActiveInheritedStyles_ReadOnly benchmarks #6 - zero-copy inherited accessor
// paletteTheme installs a theme's v1 palette as its colour source, so edits to
// the palette reach the widgets: a widget reads the semantic layer, and a theme
// built from a Seed carries one that no longer follows the palette.
func paletteTheme(t theme.Theme) theme.Theme {
	t.UseSemantic(theme.SemanticFromPalette(t.Palette))
	return t
}

func BenchmarkActiveInheritedStyles_ReadOnly(b *testing.B) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)

	// Push inherited text styles
	restore := core.PushInheritedStyle(ctx, style.Style{}.FontSize(14).TextColor(style.Solid(activeTheme.Palette.Foreground)))
	defer restore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = core.ActiveInheritedStylesReadOnly(ctx)
	}
}
