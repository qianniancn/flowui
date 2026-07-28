package state

import (
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestFocusAnimationOpacityTracksVisible(t *testing.T) {
	var animation FocusAnimation
	gtx := layout.Context{Ops: new(op.Ops), Now: time.Unix(1, 0)}
	if got := animation.Opacity(gtx, true); got != 1 {
		t.Fatalf("visible opacity = %v, want 1", got)
	}

	gtx.Now = gtx.Now.Add(focusAnimationDuration + time.Millisecond)
	if got := animation.Opacity(gtx, false); got != 0 {
		t.Fatalf("hidden opacity = %v, want 0", got)
	}
}

func TestFocusAnimationRespectsMotionTheme(t *testing.T) {
	var animation FocusAnimation
	gtx := layout.Context{Ops: new(op.Ops), Now: time.Unix(1, 0)}
	motion := theme.MotionTheme{Enabled: false, DurationScale: 1}
	animation.Opacity(gtx, false, motion)
	gtx.Now = gtx.Now.Add(time.Millisecond)
	if got := animation.Opacity(gtx, true, motion); got != 1 {
		t.Fatalf("disabled motion focus opacity = %v, want 1", got)
	}
}

func TestFocusAnimationTargetOpacityTracksLatestValue(t *testing.T) {
	var animation FocusAnimation
	gtx := layout.Context{Ops: new(op.Ops), Now: time.Unix(1, 0)}
	animation.Opacity(gtx, true)
	if got := animation.TargetOpacity(); got != 1 {
		t.Fatalf("target opacity = %v, want 1", got)
	}
	animation.Opacity(gtx, false)
	if got := animation.TargetOpacity(); got != 0 {
		t.Fatalf("target opacity = %v, want 0", got)
	}
}
