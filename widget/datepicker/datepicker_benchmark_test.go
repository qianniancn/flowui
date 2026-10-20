package datepicker

import (
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkDatePickerProviderDeclaration(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	provider, err := theme.ComponentProvider(&activeTheme)
	if err != nil {
		b.Fatal(err)
	}
	request := theme.DatePickerRequest{Variant: theme.InputPrimary}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = provider.DatePicker(&activeTheme, request)
	}
}

func BenchmarkDatePickerPartDefaultsCache(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	base := style.Style{}

	b.Run("cell-cache-hit", func(b *testing.B) {
		cache := datePickerPartDefaultsCache{}
		context := CellStyleContext{Hovered: true, Today: true, Active: true, Range: true}
		cache.cellDefaults(&activeTheme, base, context)

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = cache.cellDefaults(&activeTheme, base, context)
		}
	})

	b.Run("range-cache-hit", func(b *testing.B) {
		cache := datePickerPartDefaultsCache{}
		cache.rangeDefaults(&activeTheme, base, false)

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = cache.rangeDefaults(&activeTheme, base, false)
		}
	})

}

// BenchmarkDatePickerOpenSteadyState measures repeated calendar popup frames
// after the picker state, calendar cells, and style declarations are warm.
func BenchmarkDatePickerOpenSteadyState(b *testing.B) {
	now := testDate(2026, 7, 1)
	picker := New("bench-open", testDate(2026, 7, 9))
	activeTheme := *testtheme.Behavior()
	ctx := newContextWithThemeAndLanguage(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	layoutDatePickerFrameAt(ctx, router, picker, now)
	state := testComponentState[datePickerState](ctx, "bench-open", stateSlotDatePicker)
	if state == nil {
		b.Fatal("datepicker state was not initialized")
	}
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.openCalendar()
	layoutDatePickerFrameAt(ctx, router, picker, now.Add(time.Millisecond))
	startStats := state.resolvedParts.stats()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layoutDatePickerFrameAt(ctx, router, picker, now.Add(time.Duration(i+2)*time.Millisecond))
	}
	endStats := state.resolvedParts.stats()
	for _, metric := range []struct {
		name  string
		value uint64
	}{
		{name: "datepicker-resolved-cache-hits/op", value: endStats.Hits - startStats.Hits},
		{name: "datepicker-resolved-cache-misses/op", value: endStats.Misses - startStats.Misses},
		{name: "datepicker-resolved-cache-evictions/op", value: endStats.Evictions - startStats.Evictions},
		{name: "datepicker-resolved-cache-theme-invalidations/op", value: endStats.ThemeInvalidations - startStats.ThemeInvalidations},
		{name: "datepicker-resolved-cache-generation-invalidations/op", value: endStats.GenerationInvalidations - startStats.GenerationInvalidations},
		{name: "datepicker-resolved-cache-scope-invalidations/op", value: endStats.ScopeInvalidations - startStats.ScopeInvalidations},
		{name: "datepicker-resolved-cache-variant-invalidations/op", value: endStats.VariantInvalidations - startStats.VariantInvalidations},
		{name: "datepicker-resolved-cache-unsafe-bypasses/op", value: endStats.UnsafeBypasses - startStats.UnsafeBypasses},
	} {
		b.ReportMetric(float64(metric.value)/float64(b.N), metric.name)
	}
}
