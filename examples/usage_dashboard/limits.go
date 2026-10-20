package main

import (
	"fmt"
	"time"
)

// limitWindow is one plan window as its agent last reported it. Used is a
// percentage from 0 to 100.
type limitWindow struct {
	minutes int64
	used    float64
	resets  time.Time
}

// planLimit is one agent's latest limit report.
type planLimit struct {
	provider string
	seen     time.Time
	windows  []limitWindow
}

func demoLimits(now time.Time) []planLimit {
	return []planLimit{
		{
			provider: "claude",
			seen:     now.Add(-time.Minute),
			windows: []limitWindow{
				{minutes: 5 * 60, used: 72, resets: now.Add(time.Hour + 27*time.Minute)},
				{minutes: 7 * 24 * 60, used: 48, resets: now.Add(3*24*time.Hour + 10*time.Hour)},
			},
		},
		{
			provider: "codex",
			seen:     now.Add(-40 * time.Minute),
			windows: []limitWindow{
				{minutes: 5 * 60, used: 34, resets: now.Add(2*time.Hour + 8*time.Minute)},
				{minutes: 7 * 24 * 60, used: 61, resets: now.Add(2*24*time.Hour + 3*time.Hour)},
			},
		},
		{
			provider: "deepseek",
			seen:     now.Add(-3 * time.Minute),
			windows: []limitWindow{
				{minutes: 5 * 60, used: 26, resets: now.Add(3*time.Hour + 4*time.Minute)},
				{minutes: 7 * 24 * 60, used: 58, resets: now.Add(3*24*time.Hour + 18*time.Hour)},
			},
		},
	}
}

func limitTitle(limit planLimit) string {
	return providerName(limit.provider)
}

// windowName names a plan window by its length.
func windowName(minutes int64) string {
	switch minutes {
	case 5 * 60:
		return "5-hour"
	case 7 * 24 * 60:
		return "Weekly"
	default:
		if minutes%60 == 0 {
			return fmt.Sprintf("%d-hour", minutes/60)
		}
		return fmt.Sprintf("%d-minute", minutes)
	}
}

// windowNote is the line under a limit bar: when it resets, and when it would
// hit 100% at the pace seen so far.
func windowNote(window limitWindow, seen, now time.Time) string {
	if window.resets.IsZero() {
		return "Its reset time is unknown."
	}
	if !now.Before(window.resets) {
		return "Reset at " + clock(window.resets, now) + "; no report since."
	}
	note := "Resets in " + span(window.resets.Sub(now)) + ", at " + clock(window.resets, now)
	if window.used >= 100 {
		return note + " · Used up"
	}
	if hit, ok := windowHit(window, seen); ok {
		note += " · At this pace you hit 100% around " + clock(hit, now)
	}
	return note
}

// windowHit projects the current pace to 100%, but only when that lands
// before the reset and the window is old enough to say anything useful.
func windowHit(window limitWindow, seen time.Time) (time.Time, bool) {
	if window.used <= 0 || window.used >= 100 || window.minutes <= 0 || window.resets.IsZero() {
		return time.Time{}, false
	}
	length := time.Duration(window.minutes) * time.Minute
	elapsed := seen.Sub(window.resets.Add(-length))
	if elapsed < length/20 {
		return time.Time{}, false
	}
	hit := seen.Add(time.Duration(float64(elapsed) * (100 - window.used) / window.used))
	if !hit.Before(window.resets) {
		return time.Time{}, false
	}
	return hit, true
}

// span is a compact duration: "2h 14m", "3d 10h", "8m".
func span(duration time.Duration) string {
	if duration < time.Minute {
		return "less than a minute"
	}
	duration = duration.Round(time.Minute)
	days := int(duration / (24 * time.Hour))
	hours := int(duration/time.Hour) % 24
	minutes := int(duration/time.Minute) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func clock(value, now time.Time) string {
	return value.In(now.Location()).Format("15:04")
}

func ago(duration time.Duration) string {
	if duration < time.Minute {
		return "just now"
	}
	return span(duration) + " ago"
}
