package button

import (
	"image"
	"testing"
	"time"

	"gioui.org/widget"
)

func TestUpdateRipplesLifecycle(t *testing.T) {
	now := time.Now()
	state := &rippleState{}
	duration := 600 * time.Millisecond

	if updateRipples(state, nil, duration, now) {
		t.Fatal("empty history reports animation")
	}

	presses := []widget.Press{{Position: image.Pt(10, 5), Start: now}}
	if !updateRipples(state, presses, duration, now) {
		t.Fatal("new press should animate")
	}
	if len(state.active) != 1 || state.active[0].X != 10 || state.active[0].Y != 5 {
		t.Fatalf("ripple = %#v", state.active)
	}

	// Scanning the same history again must not record duplicate presses.
	updateRipples(state, presses, duration, now)
	if len(state.active) != 1 {
		t.Fatalf("duplicate press recorded: %#v", state.active)
	}

	// Expired entries are removed.
	if updateRipples(state, presses, duration, now.Add(duration+time.Second)) {
		t.Fatal("ripple should have expired")
	}
	if len(state.active) != 0 {
		t.Fatalf("expired ripples kept: %#v", state.active)
	}
}

func TestUpdateRipplesIgnoresCancelled(t *testing.T) {
	now := time.Now()
	state := &rippleState{}
	presses := []widget.Press{{Position: image.Pt(1, 1), Start: now, Cancelled: true}}
	if updateRipples(state, presses, 600*time.Millisecond, now) {
		t.Fatal("cancelled press should not start a ripple")
	}
	if len(state.active) != 0 {
		t.Fatalf("cancelled press recorded: %#v", state.active)
	}
}
