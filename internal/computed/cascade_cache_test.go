package computed

import (
	"image/color"
	"testing"
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestCascadeCache_Basic(t *testing.T) {
	cache := newCascadeCache(10)

	key := cascadeCacheKey{
		componentKey: "test-button",
		stateHash:    0,
		themeHash:    12345,
		customHash:   67890,
	}

	// Cache miss
	if _, found := cache.get(key); found {
		t.Error("expected cache miss on empty cache")
	}

	// Put and retrieve
	resolved := style.ResolvedStyle{}
	cache.put(key, resolved)

	if _, found := cache.get(key); !found {
		t.Error("expected cache hit after put")
	}

	stats := cache.stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Errorf("expected 1 hit and 1 miss, got hits=%d misses=%d", stats.Hits, stats.Misses)
	}
}

func TestCascadeCachePutOwnsStoredStyle(t *testing.T) {
	cache := newCascadeCache(1)
	key := cascadeCacheKey{componentKey: "menu-item"}

	initialColor := style.RGB(0x102030)
	initial := style.ResolvedStyle{Paint: &style.PaintStyle{Background: initialColor}}
	cache.put(key, initial)
	initial.Paint.Background = style.RGB(0x405060)
	got, _ := cache.get(key)
	if got.Paint == nil || got.Paint.Background != initialColor {
		t.Fatalf("inserted background = %#v, want %#v", got.Paint, initialColor)
	}

	updatedColor := style.RGB(0x708090)
	updated := style.ResolvedStyle{Paint: &style.PaintStyle{Background: updatedColor}}
	cache.put(key, updated)
	updated.Paint.Background = style.RGB(0xa0b0c0)
	got, _ = cache.get(key)
	if got.Paint == nil || got.Paint.Background != updatedColor {
		t.Fatalf("updated background = %#v, want %#v", got.Paint, updatedColor)
	}
}

func TestCascadeCacheGetOwnsNestedStyleData(t *testing.T) {
	cache := newCascadeCache(1)
	key := cascadeCacheKey{componentKey: "nested-style"}
	width := unit.Dp(1)
	profile := style.ShadowMenu
	gradient := style.Gradient{
		Stops: []style.GradientStop{{
			Offset: 0.25,
			Color:  &style.SolidColor{Color: color.NRGBA{R: 0x10, A: 0xff}},
		}},
	}
	cache.put(key, style.ResolvedStyle{Paint: &style.PaintStyle{
		Background: gradient,
		Border:     &style.BorderStyle{Width: &width},
		Shadows: []style.Shadow{{
			Color:   &style.SolidColor{Color: color.NRGBA{G: 0x20, A: 0xff}},
			Profile: &profile,
		}},
	}})

	got, ok := cache.get(key)
	if !ok || got.Paint == nil {
		t.Fatal("missing cached paint")
	}
	gotGradient := got.Paint.Background.(style.Gradient)
	gotGradient.Stops[0].Offset = 0.75
	gotGradient.Stops[0].Color.(*style.SolidColor).Color.B = 0x30
	*got.Paint.Border.Width = 2
	got.Paint.Shadows[0].Color.(*style.SolidColor).Color.B = 0x40
	*got.Paint.Shadows[0].Profile = style.ShadowOverlay

	again, ok := cache.get(key)
	if !ok || again.Paint == nil {
		t.Fatal("cached paint disappeared")
	}
	againGradient := again.Paint.Background.(style.Gradient)
	if againGradient.Stops[0].Offset != 0.25 || againGradient.Stops[0].Color.(*style.SolidColor).Color.B != 0 {
		t.Fatalf("cached gradient was aliased: %#v", againGradient)
	}
	if *again.Paint.Border.Width != 1 {
		t.Fatalf("cached border width was aliased: %v", *again.Paint.Border.Width)
	}
	if again.Paint.Shadows[0].Color.(*style.SolidColor).Color.B != 0 || *again.Paint.Shadows[0].Profile != style.ShadowMenu {
		t.Fatalf("cached shadow was aliased: %#v", again.Paint.Shadows[0])
	}
}

func TestResolveStaticCacheSharesTransitionFreeStyles(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declaration := style.Style{}.Background(style.RGB(0x102030)).Padding(8)

	// The first call populates the cache and returns the uncached result.
	ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	first := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	second := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if first.Paint == nil || second.Paint == nil || first.Paint != second.Paint {
		t.Fatal("transition-free cache hits did not share immutable paint")
	}
	if first.Box == nil || second.Box == nil || first.Box != second.Box {
		t.Fatal("transition-free cache hits did not share immutable box")
	}
}

func TestResolveStaticCacheCopiesStylesWithTransitions(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declaration := style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, 150*time.Millisecond)

	ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	first := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	second := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if first.Paint == nil || second.Paint == nil || first.Paint == second.Paint {
		t.Fatal("transition cache hits shared mutable paint")
	}
	if first.Paint.Opacity == nil || second.Paint.Opacity == nil {
		t.Fatal("transition opacity was dropped")
	}
	*first.Paint.Opacity = 0
	if *second.Paint.Opacity != 1 {
		t.Fatalf("transition cache hit was aliased: got %v", *second.Paint.Opacity)
	}
}

func TestResolvePartStaticCacheSharesTransitionFreeStyles(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declaration := style.Style{}.
		Part(style.PartContent, style.Style{}.Background(style.RGB(0x102030)).Padding(8))

	ResolvePartStatic(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	first := ResolvePartStatic(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	second := ResolvePartStatic(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if first.Paint == nil || second.Paint == nil || first.Paint != second.Paint {
		t.Fatal("transition-free part cache hits did not share immutable paint")
	}
	if first.Box == nil || second.Box == nil || first.Box != second.Box {
		t.Fatal("transition-free part cache hits did not share immutable box")
	}
}

func TestResolvePartStaticCacheCopiesStylesWithTransitions(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declaration := style.Style{}.
		Part(style.PartContent, style.Style{}.
			Opacity(1).
			Transition(style.PropOpacity, 150*time.Millisecond))

	ResolvePartStatic(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	first := ResolvePartStatic(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	second := ResolvePartStatic(ctx, style.PartContent, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if first.Paint == nil || second.Paint == nil || first.Paint == second.Paint {
		t.Fatal("transition part cache hits shared mutable paint")
	}
	if first.Paint.Opacity == nil || second.Paint.Opacity == nil {
		t.Fatal("transition part opacity was dropped")
	}
	*first.Paint.Opacity = 0
	if *second.Paint.Opacity != 1 {
		t.Fatalf("transition part cache hit was aliased: got %v", *second.Paint.Opacity)
	}
}

func TestCloneResolvedStyleForAnimationCopiesOnlyMutableBranches(t *testing.T) {
	source := style.Cascade(style.State{}, style.Style{}.
		Padding(8).
		Background(style.RGB(0x102030)).
		BorderWidth(1).
		BorderColor(style.RGB(0x203040)).
		Outline(2, 1, style.RGB(0x304050)).
		Radius(4).
		RadiusTopLeft(6).
		TextColor(style.RGB(0x405060)).
		Scale(.9).
		Transition(style.PropBackgroundColor, 100*time.Millisecond))

	clone := cloneResolvedStyleForAnimation(source)
	if clone.Box != source.Box {
		t.Fatal("animation clone copied immutable box")
	}
	if clone.Paint == source.Paint || clone.Paint.Border == source.Paint.Border || clone.Paint.Outline == source.Paint.Outline || clone.Paint.Radii == source.Paint.Radii {
		t.Fatal("animation clone shared mutable paint state")
	}
	if clone.Text == source.Text {
		t.Fatal("animation clone shared mutable text state")
	}
	if clone.Trans == source.Trans {
		t.Fatal("animation clone shared mutable transform state")
	}

	clone.Paint.Background = style.RGB(0xffffff)
	clone.Paint.Border.Color = style.RGB(0xffffff)
	clone.Paint.Outline.Color = style.RGB(0xffffff)
	clone.Paint.Radii.TopLeft = 12
	clone.Text.Color = style.RGB(0xffffff)
	scale := float32(.5)
	clone.Trans.ScaleX = &scale

	background, ok := Brush(source.Paint.Background)
	if !ok || background.ColorAt(.5) != (color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}) {
		t.Fatalf("source background mutated: %#v", background)
	}
	if source.Paint.Radii.TopLeft != unit.Dp(6) || *source.Trans.ScaleX != .9 {
		t.Fatal("source radius or transform mutated")
	}
}

func TestSharedStaticStyleSurvivesCopyOnWriteHelpers(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declaration := style.Style{}.Outline(2, 1, style.RGB(0x204060))
	ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	shared := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	ApplyOutlineOpacity(shared, .25)
	again := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if again.Paint == nil || again.Paint.Outline == nil {
		t.Fatal("cached outline disappeared")
	}
	got, ok := Color(again.Paint.Outline.Color)
	if !ok || got.A != 0xff {
		t.Fatalf("copy-on-write helper mutated shared style: %#v", got)
	}
}

func TestCascadeCache_LRU(t *testing.T) {
	cache := newCascadeCache(3)

	// Fill cache
	for i := range 3 {
		key := cascadeCacheKey{
			componentKey: string(rune('a' + i)),
			stateHash:    0,
			themeHash:    0,
			customHash:   0,
		}
		cache.put(key, style.ResolvedStyle{})
	}

	if cache.size != 3 {
		t.Errorf("expected size 3, got %d", cache.size)
	}

	// Add one more - should evict LRU (first entry)
	key4 := cascadeCacheKey{
		componentKey: "d",
		stateHash:    0,
		themeHash:    0,
		customHash:   0,
	}
	cache.put(key4, style.ResolvedStyle{})

	if cache.size != 3 {
		t.Errorf("expected size 3 after eviction, got %d", cache.size)
	}

	// First entry should be evicted
	key1 := cascadeCacheKey{
		componentKey: "a",
		stateHash:    0,
		themeHash:    0,
		customHash:   0,
	}
	if _, found := cache.get(key1); found {
		t.Error("expected first entry to be evicted")
	}

	// Other entries should still be present
	key2 := cascadeCacheKey{
		componentKey: "b",
		stateHash:    0,
		themeHash:    0,
		customHash:   0,
	}
	if _, found := cache.get(key2); !found {
		t.Error("expected second entry to still be cached")
	}
}

func TestCascadeCache_Clear(t *testing.T) {
	cache := newCascadeCache(10)

	// Add entries
	for i := range 5 {
		key := cascadeCacheKey{
			componentKey: string(rune('a' + i)),
			stateHash:    0,
			themeHash:    0,
			customHash:   0,
		}
		cache.put(key, style.ResolvedStyle{})
	}

	if cache.size != 5 {
		t.Errorf("expected size 5, got %d", cache.size)
	}

	// Clear
	cache.clear()

	if cache.size != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.size)
	}

	stats := cache.stats()
	if stats.Size != 0 {
		t.Errorf("expected stats size 0 after clear, got %d", stats.Size)
	}
}

func TestComputeStyleStateHash(t *testing.T) {
	tests := []struct {
		name  string
		state style.State
		want  uint16
	}{
		{
			name:  "empty",
			state: style.State{},
			want:  0,
		},
		{
			name:  "hovered",
			state: style.State{Hovered: true},
			want:  1 << 0,
		},
		{
			name:  "pressed",
			state: style.State{Pressed: true},
			want:  1 << 1,
		},
		{
			name:  "hovered and pressed",
			state: style.State{Hovered: true, Pressed: true},
			want:  (1 << 0) | (1 << 1),
		},
		{
			name:  "disabled",
			state: style.State{Disabled: true},
			want:  1 << 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeStyleStateHash(tt.state)
			if got != tt.want {
				t.Errorf("computeStyleStateHash() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCustomStyleHash(t *testing.T) {
	cache := newCascadeCache(10)

	// Same style should produce same hash
	s1 := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	s2 := style.Style{}.Padding(8).Background(style.RGB(0xffffff))

	hash1 := cache.computeCustomStyleHash(s1)
	hash2 := cache.computeCustomStyleHash(s2)

	if hash1 != hash2 {
		t.Error("expected identical styles to produce same hash")
	}

	// Different style should produce different hash
	s3 := style.Style{}.Padding(16).Background(style.RGB(0x000000))
	hash3 := cache.computeCustomStyleHash(s3)

	if hash1 == hash3 {
		t.Error("expected different styles to produce different hash")
	}
}

func TestCustomStyleHashDistinguishesPropertyIdentity(t *testing.T) {
	cache := newCascadeCache(10)
	width := style.Style{}.Width(8)
	height := style.Style{}.Height(8)

	if cache.computeCustomStyleHash(width) == cache.computeCustomStyleHash(height) {
		t.Fatal("width and height declarations produced the same hash")
	}
}

func TestResolveStaticCacheDistinguishesConditionalPredicates(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	state := style.State{Hovered: true}
	hovered := style.Style{}.On(style.Hovered, style.Style{}.Width(8))
	pressed := style.Style{}.On(style.Pressed, style.Style{}.Width(8))

	first := ResolveStatic(ctx, state, hovered, style.Style{}, style.Style{}, style.Style{})
	second := ResolveStatic(ctx, state, pressed, style.Style{}, style.Style{}, style.Style{})
	if first.Box == nil || first.Box.Width == nil {
		t.Fatal("hovered declaration did not resolve width")
	}
	if second.Box != nil && second.Box.Width != nil {
		t.Fatal("pressed declaration reused the hovered cache entry")
	}
}

func TestResolveStaticCacheTracksRuntimeThemeShadows(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Shadow(style.ShadowSurface)
	first := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	activeTheme.Shadows.Surface = theme.ShadowTheme{
		Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{{OffsetY: 7, Blur: 9, Opacity: 1}},
	}
	core.ReplaceTheme(ctx, activeTheme)
	second := ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if first.Paint == nil || second.Paint == nil || first.Paint == second.Paint {
		t.Fatal("replacement shadow profile reused the cached entry")
	}
	for _, shadow := range second.Paint.Shadows {
		if shadow.OffsetY == 7 && shadow.Blur == 9 {
			return
		}
	}
	t.Fatalf("replacement surface shadow = %#v, want offsetY 7 blur 9", second.Paint.Shadows)
}

func TestResolvePartStaticCacheTracksInheritedStyles(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	want := style.RGB(0xfefefe)
	inherited := style.RGB(0x111111)
	defaults := style.Style{}.Part(style.PartIndicator, style.Style{}.TextColor(want))

	restore := core.PushInheritedStyle(ctx, style.Style{}.TextColor(inherited))
	first := ResolvePartStatic(ctx, style.PartIndicator, style.State{}, defaults, style.Style{}, style.Style{}, style.Style{})
	restore()
	second := ResolvePartStatic(ctx, style.PartIndicator, style.State{}, defaults, style.Style{}, style.Style{}, style.Style{})

	if got, _ := solidColor(first.Text.Color); got != inherited.Color {
		t.Fatalf("inherited color = %#v, want %#v", got, inherited.Color)
	}
	if got, _ := solidColor(second.Text.Color); got != want.Color {
		t.Fatalf("unscoped color = %#v, want %#v", got, want.Color)
	}
}

// Benchmark cascade cache operations

func BenchmarkCascadeCache_Hit(b *testing.B) {
	cache := newCascadeCache(512)

	key := cascadeCacheKey{
		componentKey: "test-button",
		stateHash:    0,
		themeHash:    12345,
		customHash:   67890,
	}

	resolved := style.ResolvedStyle{}
	cache.put(key, resolved)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.get(key)
	}
}

func BenchmarkCascadeCache_Miss(b *testing.B) {
	cache := newCascadeCache(512)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := cascadeCacheKey{
			componentKey: "test-button",
			stateHash:    uint16(i),
			themeHash:    12345,
			customHash:   67890,
		}
		_, _ = cache.get(key)
	}
}

func BenchmarkCascadeCache_Put(b *testing.B) {
	cache := newCascadeCache(512)
	resolved := style.ResolvedStyle{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := cascadeCacheKey{
			componentKey: "test-button",
			stateHash:    uint16(i % 100),
			themeHash:    12345,
			customHash:   67890,
		}
		cache.put(key, resolved)
	}
}

func BenchmarkComputeStyleStateHash(b *testing.B) {
	state := style.State{
		Hovered:  true,
		Pressed:  false,
		Focused:  true,
		Disabled: false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = computeStyleStateHash(state)
	}
}

func BenchmarkComputeCustomStyleHash(b *testing.B) {
	cache := newCascadeCache(512)
	s := style.Style{}.Padding(8).Background(style.RGB(0xffffff))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = cache.computeCustomStyleHash(s)
	}
}

// Integration benchmarks comparing cached vs uncached resolution

func BenchmarkResolveStatic_WithCache(b *testing.B) {
	DisableCascadeCache = false
	defer func() { DisableCascadeCache = false }()

	// Create minimal test context - just need theme
	activeTheme := theme.NeutralTheme()
	state := style.State{Hovered: true}
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	// Warm cache with direct call to uncached version
	resolved := style.Cascade(state, defaults, variant, size, custom)

	// Cache the result
	themeHash := computeThemeHash(&activeTheme)
	stateHash := computeStyleStateHash(state)
	customHash := globalCascadeCache.computeCustomStyleHash(custom)
	cacheKey := cascadeCacheKey{
		componentKey: "",
		stateHash:    stateHash,
		themeHash:    themeHash,
		customHash:   customHash,
	}
	globalCascadeCache.put(cacheKey, resolved)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Direct cache lookup
		_, _ = globalCascadeCache.get(cacheKey)
	}
}

func BenchmarkResolveStatic_WithoutCache(b *testing.B) {
	DisableCascadeCache = true
	defer func() { DisableCascadeCache = false }()

	state := style.State{Hovered: true}
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resolved := style.Cascade(state, defaults, variant, size, custom)
		_ = resolved
	}
}

func BenchmarkResolveStatic_MixedHitRate(b *testing.B) {
	DisableCascadeCache = false
	defer func() { DisableCascadeCache = false }()

	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	// Simulate realistic scenario with ~80% cache hit rate
	states := []style.State{
		{},               // Normal
		{Hovered: true},  // Hover
		{Pressed: true},  // Pressed
		{Disabled: true}, // Disabled
		{Focused: true},  // Focused
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		state := states[i%len(states)]
		resolved := style.Cascade(state, defaults, variant, size, custom)
		_ = resolved
	}
}
