package main

import (
	"testing"
	"time"
)

func TestRollingYearCovers365CalendarDays(t *testing.T) {
	end := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)
	start := end.AddDate(0, 0, -364)
	values := calendarValues(start, end, 73)
	if len(values) != 365 {
		t.Fatalf("rolling-year value count = %d, want 365", len(values))
	}
	if !values[0].Date.Equal(start) || !values[len(values)-1].Date.Equal(end) {
		t.Fatalf("rolling-year bounds = %s to %s, want %s to %s", values[0].Date, values[len(values)-1].Date, start, end)
	}
}
