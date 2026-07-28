package runtime

import (
	"testing"

	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
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
		state style.StyleState
		want  uint16
	}{
		{
			name:  "empty",
			state: style.StyleState{},
			want:  0,
		},
		{
			name:  "hovered",
			state: style.StyleState{Hovered: true},
			want:  1 << 0,
		},
		{
			name:  "pressed",
			state: style.StyleState{Pressed: true},
			want:  1 << 1,
		},
		{
			name:  "hovered and pressed",
			state: style.StyleState{Hovered: true, Pressed: true},
			want:  (1 << 0) | (1 << 1),
		},
		{
			name:  "disabled",
			state: style.StyleState{Disabled: true},
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

	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	state := style.StyleState{Hovered: true}
	hovered := style.Style{}.When(style.Hovered, style.Style{}.Width(8))
	pressed := style.Style{}.When(style.Pressed, style.Style{}.Width(8))

	first := ResolveStatic(ctx, state, hovered, style.Style{}, style.Style{}, style.Style{})
	second := ResolveStatic(ctx, state, pressed, style.Style{}, style.Style{}, style.Style{})
	if first.Box == nil || first.Box.Width == nil {
		t.Fatal("hovered declaration did not resolve width")
	}
	if second.Box != nil && second.Box.Width != nil {
		t.Fatal("pressed declaration reused the hovered cache entry")
	}
}

func TestResolveStaticCacheTracksRuntimeThemeColors(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Background(style.TokenSurface)
	first := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})

	want := style.RGB(0x123456)
	activeTheme.Palette.Surface = want.Color
	frame.ReplaceTheme(ctx, activeTheme)
	second := ResolveStatic(ctx, style.StyleState{}, declaration, style.Style{}, style.Style{}, style.Style{})

	if got, _ := solidColor(first.Paint.Background); got == want.Color {
		t.Fatal("initial theme unexpectedly used replacement color")
	}
	if got, _ := solidColor(second.Paint.Background); got != want.Color {
		t.Fatalf("replacement surface = %#v, want %#v", got, want.Color)
	}
}

func TestResolvePartStaticCacheTracksInheritedStyles(t *testing.T) {
	previous := DisableCascadeCache
	DisableCascadeCache = false
	ClearCascadeCache()
	defer func() {
		DisableCascadeCache = previous
		ClearCascadeCache()
	}()

	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	want := style.RGB(0xfefefe)
	inherited := style.RGB(0x111111)
	defaults := style.Style{}.Part(style.PartIndicator, style.Style{}.TextColor(want))

	restore := frame.PushInheritedStyle(ctx, style.Style{}.TextColor(inherited))
	first := ResolvePartStatic(ctx, style.PartIndicator, style.StyleState{}, defaults, style.Style{}, style.Style{}, style.Style{})
	restore()
	second := ResolvePartStatic(ctx, style.PartIndicator, style.StyleState{}, defaults, style.Style{}, style.Style{}, style.Style{})

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
	state := style.StyleState{
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
	activeTheme := theme.DefaultTheme()
	state := style.StyleState{Hovered: true}
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	// Warm cache with direct call to uncached version
	layers := []style.Style{defaults, variant, size, custom}
	layers = resolveThemeMetrics(layers, &activeTheme)
	resolved := style.Cascade(state, layers...)
	resolveThemeColors(&resolved, &activeTheme)

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

	activeTheme := theme.DefaultTheme()
	state := style.StyleState{Hovered: true}
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		layers := []style.Style{defaults, variant, size, custom}
		layers = resolveThemeMetrics(layers, &activeTheme)
		resolved := style.Cascade(state, layers...)
		resolveThemeColors(&resolved, &activeTheme)
		_ = resolved
	}
}

func BenchmarkResolveStatic_MixedHitRate(b *testing.B) {
	DisableCascadeCache = false
	defer func() { DisableCascadeCache = false }()

	activeTheme := theme.DefaultTheme()
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	// Simulate realistic scenario with ~80% cache hit rate
	states := []style.StyleState{
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
		layers := []style.Style{defaults, variant, size, custom}
		layers = resolveThemeMetrics(layers, &activeTheme)
		resolved := style.Cascade(state, layers...)
		resolveThemeColors(&resolved, &activeTheme)
		_ = resolved
	}
}
