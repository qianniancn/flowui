package computed

import (
	"image/color"
	"testing"
	"time"

	giotext "gioui.org/text"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func benchmarkResolvedStyleCases() []struct {
	name  string
	value style.ResolvedStyle
} {
	empty := style.ResolvedStyle{}
	typical := style.ResolvedStyle{
		Box: &style.BoxStyle{
			Width:   benchmarkPtr(unit.Dp(160)),
			Padding: &style.Insets{Top: 8, Right: 12, Bottom: 8, Left: 12},
		},
		Paint: &style.PaintStyle{
			Background: style.SolidColor{Color: color.NRGBA{R: 0x20, G: 0x60, B: 0xa0, A: 0xff}},
			Border: &style.BorderStyle{
				Color: style.SolidColor{Color: color.NRGBA{A: 0xff}},
				Width: benchmarkPtr(unit.Dp(1)),
			},
			Radius: benchmarkPtr(unit.Dp(8)),
		},
		Text: &style.TextStyle{
			Color:    style.SolidColor{Color: color.NRGBA{A: 0xff}},
			FontSize: benchmarkPtr(unit.Sp(14)),
		},
		Transitions: []style.Transition{{
			Property: style.PropBackgroundColor,
			Duration: 150 * time.Millisecond,
		}},
	}
	full := typical
	full.Box = &style.BoxStyle{
		Width:       benchmarkPtr(unit.Dp(160)),
		Height:      benchmarkPtr(unit.Dp(36)),
		MinWidth:    benchmarkPtr(unit.Dp(80)),
		MaxWidth:    benchmarkPtr(unit.Dp(320)),
		MinHeight:   benchmarkPtr(unit.Dp(24)),
		MaxHeight:   benchmarkPtr(unit.Dp(64)),
		FillWidth:   new(true),
		FillHeight:  new(false),
		AspectRatio: new(float32(2)),
		Padding:     &style.Insets{Top: 8, Right: 12, Bottom: 8, Left: 12},
		Margin:      &style.Insets{Top: 2, Right: 3, Bottom: 4, Left: 5},
		Overflow:    benchmarkPtr(style.OverflowHidden),
	}
	full.Paint = &style.PaintStyle{
		Background: style.Gradient{AngleDegrees: 90, Stops: []style.GradientStop{{Offset: 0, Color: style.SolidColor{Color: color.NRGBA{A: 0xff}}}}},
		Border: &style.BorderStyle{
			Color: style.SolidColor{Color: color.NRGBA{A: 0xff}},
			Width: benchmarkPtr(unit.Dp(1)),
		},
		BorderBottom: &style.BorderStyle{
			Color: style.SolidColor{Color: color.NRGBA{A: 0xff}},
			Width: benchmarkPtr(unit.Dp(1)),
		},
		Radius: benchmarkPtr(unit.Dp(8)),
		Radii:  &style.CornerRadii{TopLeft: 8, TopRight: 7, BottomRight: 6, BottomLeft: 5},
		Shadows: []style.Shadow{{
			OffsetX: 1,
			OffsetY: 2,
			Blur:    4,
			Spread:  1,
			Color:   style.SolidColor{Color: color.NRGBA{A: 0x40}},
		}},
		Outline: &style.OutlineStyle{
			Width:  1,
			Offset: 2,
			Color:  style.SolidColor{Color: color.NRGBA{A: 0xff}},
		},
		VisualOutset: &style.Insets{Top: 2, Right: 2, Bottom: 2, Left: 2},
		Opacity:      new(float32(.9)),
	}
	full.Text = &style.TextStyle{
		Color:           style.SolidColor{Color: color.NRGBA{A: 0xff}},
		FontSize:        benchmarkPtr(unit.Sp(14)),
		FontWeight:      new(600),
		LineHeight:      benchmarkPtr(unit.Sp(20)),
		LineHeightScale: new(float32(1.2)),
		MaxLines:        new(2),
		Align:           benchmarkPtr(style.TextAlignCenter),
		Wrap:            benchmarkPtr(giotext.WrapGraphemes),
		Truncator:       new("..."),
	}
	full.Trans = &style.TransformStyle{
		TranslateX: benchmarkPtr(unit.Dp(1)),
		TranslateY: benchmarkPtr(unit.Dp(2)),
		ScaleX:     new(float32(1)),
		ScaleY:     new(float32(1)),
		Rotate:     new(float32(.1)),
	}
	full.Transitions = []style.Transition{
		{Property: style.PropBackgroundColor, Duration: 150 * time.Millisecond},
		{Property: style.PropTextColor, Duration: 100 * time.Millisecond},
		{Property: style.PropTransform, Duration: 200 * time.Millisecond},
	}
	return []struct {
		name  string
		value style.ResolvedStyle
	}{
		{name: "Empty", value: empty},
		{name: "Typical", value: typical},
		{name: "Full", value: full},
	}
}

//go:fix inline
func benchmarkPtr[T any](value T) *T {
	return new(value)
}

func BenchmarkCloneResolvedStyle(b *testing.B) {
	for _, benchmark := range benchmarkResolvedStyleCases() {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			var result style.ResolvedStyle
			for b.Loop() {
				result = cloneResolvedStyle(benchmark.value)
			}
			if result.Box == nil && benchmark.value.Box != nil {
				b.Fatal("clone dropped box")
			}
		})
	}
}

func BenchmarkCascadeCacheGetClone(b *testing.B) {
	for _, benchmark := range benchmarkResolvedStyleCases() {
		b.Run(benchmark.name, func(b *testing.B) {
			cache := newCascadeCache(1)
			key := cascadeCacheKey{componentKey: benchmark.name}
			cache.put(key, benchmark.value)
			b.ReportAllocs()
			var result style.ResolvedStyle
			for b.Loop() {
				result, _ = cache.get(key)
			}
			if result.Box == nil && benchmark.value.Box != nil {
				b.Fatal("cache get dropped box")
			}
		})
	}
}

func BenchmarkCascadeCacheGetShared(b *testing.B) {
	for _, benchmark := range benchmarkResolvedStyleCases() {
		b.Run(benchmark.name, func(b *testing.B) {
			cache := newCascadeCache(1)
			key := cascadeCacheKey{componentKey: benchmark.name}
			cache.put(key, benchmark.value)
			b.ReportAllocs()
			var result style.ResolvedStyle
			for b.Loop() {
				result, _ = cache.getShared(key)
			}
			if result.Box == nil && benchmark.value.Box != nil {
				b.Fatal("cache get dropped box")
			}
		})
	}
}

func BenchmarkResolveStaticCacheOwnership(b *testing.B) {
	for _, benchmark := range []struct {
		name        string
		declaration style.Style
	}{
		{name: "SharedWithoutTransitions", declaration: style.Style{}.Padding(8).Background(style.RGB(0xffffff))},
		{name: "OwnedWithTransitions", declaration: style.Style{}.
			Padding(8).
			Background(style.RGB(0xffffff)).
			Transition(style.PropBackgroundColor, 150*time.Millisecond)},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			DisableCascadeCache = false
			defer func() { DisableCascadeCache = false }()
			ClearCascadeCache()

			activeTheme := theme.NeutralTheme()
			ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
			resolved := ResolveStatic(ctx, style.State{}, benchmark.declaration, style.Style{}, style.Style{}, style.Style{})
			if resolved.Paint == nil {
				b.Fatal("warm-up resolve returned no paint")
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				resolved = ResolveStatic(ctx, style.State{}, benchmark.declaration, style.Style{}, style.Style{}, style.Style{})
			}
			if resolved.Paint == nil {
				b.Fatal("resolve returned no paint")
			}
		})
	}
}
