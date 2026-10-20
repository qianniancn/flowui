package button

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type buttonBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
	values []Widget
}

func newButtonBenchmarkFrame(maximum image.Point) *buttonBenchmarkFrame {
	activeTheme := *testtheme.Behavior()
	frame := &buttonBenchmarkFrame{ctx: core.New(nil, &activeTheme, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func newButtonComponentTreeFrame(count int) *buttonBenchmarkFrame {
	frame := newButtonBenchmarkFrame(image.Pt(640, 480))
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New("Save changes").Key("benchmark-button-" + strconv.Itoa(index)).Variant(Primary)
	}
	return frame
}

func (frame *buttonBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	dimensions := value.Layout(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
	return dimensions
}

func (frame *buttonBenchmarkFrame) layoutTree() layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	children := make([]layout.FlexChild, len(frame.values))
	for index := range frame.values {
		value := frame.values[index]
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(frame.ctx, gtx)
		})
	}
	dimensions := layout.Flex{Axis: layout.Vertical}.Layout(frame.gtx, children...)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
	return dimensions
}

func (frame *buttonBenchmarkFrame) cacheStats() ButtonDeclarationCacheStats {
	var total ButtonDeclarationCacheStats
	for index := range frame.values {
		state, ok := core.PeekState[buttonDeclarationCache](frame.ctx, frame.values[index].key, stateSlotStyle)
		if !ok || state == nil {
			continue
		}
		stats := state.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func buttonCacheStatsDelta(after, before ButtonDeclarationCacheStats) ButtonDeclarationCacheStats {
	return ButtonDeclarationCacheStats{
		Hits:               after.Hits - before.Hits,
		Misses:             after.Misses - before.Misses,
		ThemeInvalidations: after.ThemeInvalidations - before.ThemeInvalidations,
		KeyInvalidations:   after.KeyInvalidations - before.KeyInvalidations,
	}
}

func reportButtonCacheStats(b *testing.B, stats ButtonDeclarationCacheStats) {
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "button-declaration-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "button-declaration-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "button-declaration-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "button-declaration-cache-key-invalidations/op")
}

func BenchmarkButtonLayoutSteadyState(b *testing.B) {
	frame := newButtonBenchmarkFrame(image.Pt(320, 80))
	value := New("Save changes").Variant(Primary)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkButtonGroupLayoutSteadyState(b *testing.B) {
	frame := newButtonBenchmarkFrame(image.Pt(320, 80))
	value := Group(
		New("Save").Variant(Primary),
		New("Cancel").Variant(Secondary),
	).Separators(true)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

// BenchmarkButtonComponentTreeSteadyState measures the same provider and
// cascade path over a realistic repeated control tree rather than one button.
func BenchmarkButtonComponentTreeSteadyState(b *testing.B) {
	frame := newButtonComponentTreeFrame(128)
	frame.layoutTree()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkButtonComponentTreeVisualCacheHit(b *testing.B) {
	frame := newButtonComponentTreeFrame(128)
	frame.layoutTree()
	baseline := frame.cacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
	b.StopTimer()
	reportButtonCacheStats(b, buttonCacheStatsDelta(frame.cacheStats(), baseline))
}

func BenchmarkButtonComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newButtonComponentTreeFrame(128)
	frame.layoutTree()
	baseline := frame.cacheStats()
	light := *testtheme.Behavior()
	dark := *testtheme.Behavior()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, light)
		} else {
			core.ReplaceTheme(frame.ctx, dark)
		}
		frame.layoutTree()
	}
	b.StopTimer()
	reportButtonCacheStats(b, buttonCacheStatsDelta(frame.cacheStats(), baseline))
}

func BenchmarkStyledVerticalButtonGroupLayoutSteadyState(b *testing.B) {
	frame := newButtonBenchmarkFrame(image.Pt(320, 160))
	value := (Group(
		New("Save").Variant(Primary),
		New("Cancel").Variant(Secondary),
	).Orientation(GroupVertical)).Style(style.Style{}.PaddingX(8))
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkButtonProviderStyleDeclarations(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	value := New("").Content(nil).Variant(Primary).Size(Medium)

	b.Run("cache-hit", func(b *testing.B) {
		cache := buttonDeclarationCache{}
		value.declarationsForOverride(&activeTheme, &cache, nil)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			value.declarationsForOverride(&activeTheme, &cache, nil)
		}
		b.StopTimer()
		b.ReportMetric(float64(cache.Stats().Hits)/float64(b.N), "cache-hits/op")
	})

	b.Run("cache-miss", func(b *testing.B) {
		cache := buttonDeclarationCache{}
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			variant := theme.ButtonPrimary
			if index&1 != 0 {
				variant = theme.ButtonSecondary
			}
			value.Variant(variant).declarationsForOverride(&activeTheme, &cache, nil)
		}
		b.StopTimer()
		b.ReportMetric(float64(cache.Stats().Misses)/float64(b.N), "cache-misses/op")
		b.ReportMetric(float64(cache.Stats().KeyInvalidations)/float64(b.N), "key-invalidations/op")
	})

	b.Run("override-cache-hit", func(b *testing.B) {
		cache := buttonDeclarationCache{}
		override := &theme.ButtonOverride{Declaration: style.Style{}.PaddingX(8)}
		value.declarationsForOverride(&activeTheme, &cache, override)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			value.declarationsForOverride(&activeTheme, &cache, override)
		}
		b.StopTimer()
		b.ReportMetric(float64(cache.Stats().Hits)/float64(b.N), "cache-hits/op")
	})

	b.Run("override-cache-miss", func(b *testing.B) {
		cache := buttonDeclarationCache{}
		first := &theme.ButtonOverride{Declaration: style.Style{}.PaddingX(8)}
		second := &theme.ButtonOverride{Declaration: style.Style{}.PaddingX(12)}
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			override := first
			if index&1 != 0 {
				override = second
			}
			value.declarationsForOverride(&activeTheme, &cache, override)
		}
		b.StopTimer()
		b.ReportMetric(float64(cache.Stats().Misses)/float64(b.N), "cache-misses/op")
		b.ReportMetric(float64(cache.Stats().KeyInvalidations)/float64(b.N), "key-invalidations/op")
	})
}
