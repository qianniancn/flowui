package heatmap

import (
	"testing"
	"time"
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
