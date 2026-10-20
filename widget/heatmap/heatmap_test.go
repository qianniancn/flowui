package heatmap

import (
	"errors"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestHeatmapBuildsCalendarWeekGrid(t *testing.T) {
	start := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	widget := New("calendar", start, start.AddDate(0, 0, 6), []CalendarValue{{Date: start, Value: 4}})
	if len(widget.values) != 7 || len(widget.values[0]) != 2 {
		t.Fatalf("calendar grid = %d rows x %d columns", len(widget.values), len(widget.values[0]))
	}
	if widget.values[int(start.Weekday())][0] != 4 {
		t.Fatalf("calendar value = %v", widget.values[int(start.Weekday())][0])
	}
}

func TestHeatmapCalendarLabelsFollowWindowLanguage(t *testing.T) {
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	widget := New("calendar-labels", start, start.AddDate(0, 0, 6), nil)
	english, englishWeekdays := widget.calendarLabels(core.New(nil, testtheme.Behavior(), locale.LanguageEnglish))
	chinese, chineseWeekdays := widget.calendarLabels(core.New(nil, testtheme.Behavior(), locale.LanguageChinese))
	if english[0] != "Jan" || englishWeekdays[int(time.Monday)] != "Mon" {
		t.Fatalf("English labels = %q, %q", english[0], englishWeekdays[int(time.Monday)])
	}
	if chinese[0] != "1月" || chineseWeekdays[int(time.Monday)] != "一" {
		t.Fatalf("Chinese labels = %q, %q", chinese[0], chineseWeekdays[int(time.Monday)])
	}
}

func TestHeatmapMapsValuesAcrossTimeZones(t *testing.T) {
	// Range bounds in UTC; values carry a different location but the same civil
	// days. The cells must still be filled (regression: time.Time map keys
	// compare their location, which dropped every value across zones).
	tokyo := time.FixedZone("JST", 9*60*60)
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC) // Wednesday
	end := start.AddDate(0, 0, 6)
	values := []CalendarValue{
		{Date: time.Date(2025, time.January, 1, 0, 0, 0, 0, tokyo), Value: 7},
		{Date: time.Date(2025, time.January, 3, 0, 0, 0, 0, tokyo), Value: 9},
	}
	w := New("cross-tz", start, end, values)

	jan1 := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	if got := w.values[int(jan1.Weekday())][0]; got != 7 {
		t.Fatalf("Jan 1 cell = %v, want 7", got)
	}
	jan3 := time.Date(2025, time.January, 3, 0, 0, 0, 0, time.UTC)
	if got := w.values[int(jan3.Weekday())][0]; got != 9 {
		t.Fatalf("Jan 3 cell = %v, want 9", got)
	}
}

func TestHeatmapComparesRangeAsCivilDates(t *testing.T) {
	west := time.FixedZone("UTC-12", -12*60*60)
	east := time.FixedZone("UTC+14", 14*60*60)
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, west)
	end := time.Date(2025, time.January, 1, 0, 0, 0, 0, east)

	w := New("civil-range", start, end, nil)
	if len(w.values) != 7 || len(w.values[0]) != 1 {
		t.Fatalf("calendar grid = %d rows x %d columns", len(w.values), len(w.values[0]))
	}
}

func TestHeatmapValueRangeOverridesData(t *testing.T) {
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	w := New("test", start, start, []CalendarValue{{Date: start, Value: 10}}).ValueRange(0, 100)
	minimum, maximum := w.valueExtent(1, 1)
	if minimum != 0 || maximum != 100 {
		t.Fatalf("range = %v, %v", minimum, maximum)
	}
}

func TestHeatmapVisualCacheTracksThemeLifecycle(t *testing.T) {
	active := testtheme.Behavior()
	value := New("cache", time.Unix(1, 0), time.Unix(2, 0), nil)

	value.resolveVisual(active, nil)
	value.resolveVisual(active, nil)
	replacement := *active
	value.resolveVisual(&replacement, nil)

	stats := value.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 0 {
		t.Fatalf("heatmap visual cache stats = %#v", stats)
	}
}

type incompleteHeatmapProvider struct{ theme.ComponentVisualProvider }

func (incompleteHeatmapProvider) Heatmap(*theme.Theme, theme.HeatmapRequest) theme.HeatmapVisual {
	return theme.HeatmapVisual{}
}

func TestHeatmapIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteHeatmapProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	value := New("invalid-cache", time.Unix(1, 0), time.Unix(2, 0), nil)

	deferred := func() (recovered any) {
		defer func() { recovered = recover() }()
		value.resolveVisual(&active, nil)
		return nil
	}()
	if recoveredErr, ok := deferred.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
		t.Fatalf("panic = %v, want %v", deferred, theme.ErrIncompleteVisualProvider)
	}
	if stats := value.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
		t.Fatalf("incomplete heatmap visual polluted cache: %#v", stats)
	}

	active.VisualProvider = testtheme.BehaviorProvider()
	value.resolveVisual(&active, nil)
	if stats := value.visualCache.Stats(); stats.Misses != 1 || stats.Hits != 0 {
		t.Fatalf("valid heatmap visual did not replace failed load: %#v", stats)
	}
}

func TestHeatmapInstanceOverrideBypassesVisualCache(t *testing.T) {
	active := testtheme.Behavior()
	value := New("override", time.Unix(1, 0), time.Unix(2, 0), nil)
	override := &theme.HeatmapOverride{}

	value.resolveVisual(active, override)
	value.resolveVisual(active, override)

	stats := value.visualCache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Fatalf("override unexpectedly populated heatmap cache: %#v", stats)
	}
}

func TestHeatmapColorOverridesAreAppliedByProvider(t *testing.T) {
	active := testtheme.Behavior()
	minColor := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	maxColor := color.NRGBA{R: 0xd0, G: 0xe0, B: 0xf0, A: 0xff}
	value := New("colors", time.Unix(1, 0), time.Unix(2, 0), nil).
		MinColor(minColor).
		MaxColor(maxColor)

	visual := value.resolveVisual(active, nil)
	if visual.Colors.Min != minColor || visual.Colors.Max != maxColor {
		t.Fatalf("heatmap colors = %#v, %#v; want %#v, %#v", visual.Colors.Min, visual.Colors.Max, minColor, maxColor)
	}
	if stats := value.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
		t.Fatalf("instance colors unexpectedly populated visual cache: %#v", stats)
	}
}

func TestHeatmapStyledDeclarationComposesWithColorOverride(t *testing.T) {
	active := testtheme.Behavior()
	minColor := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	value := New("styled-colors", time.Unix(1, 0), time.Unix(2, 0), nil).MinColor(minColor)
	declaration := style.Style{}.PaddingX(9)

	visual := value.resolveVisual(active, &theme.HeatmapOverride{Declaration: declaration})
	if visual.Colors.Min != minColor {
		t.Fatalf("composed heatmap min color = %#v, want %#v", visual.Colors.Min, minColor)
	}
	if visual.Override.Hash64() != declaration.Hash64() {
		t.Fatalf("composed heatmap override = %#v, want %#v", visual.Override, declaration)
	}
}

func TestHeatmapSelectionUsesProviderColorScale(t *testing.T) {
	value := New("scale", time.Unix(1, 0), time.Unix(2, 0), nil)
	value.values = [][]float64{{5}}
	value.cellLabels = [][]string{{"cell"}}
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	calls := 0
	scale := func(minValue, maxValue, value float64) color.NRGBA {
		calls++
		if minValue != 0 || maxValue != 10 || value != 5 {
			t.Fatalf("scale inputs = %v, %v, %v; want 0, 10, 5", minValue, maxValue, value)
		}
		return want
	}
	selection := value.selection(f32.Pt(5, 5), true, image.Rect(0, 0, 10, 10), 10, 10, 1, 1, 0, 10, scale)
	if !selection.valid || selection.color != want {
		t.Fatalf("selection = %#v, want provider color %#v", selection, want)
	}
	if calls != 1 {
		t.Fatalf("provider scale calls = %d, want 1", calls)
	}
}
