package animation

import (
	"image/color"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestLocalTransitionsTrackTargets(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := layout.Context{Ops: new(op.Ops), Now: start}
	var scalar FloatTransition
	if got := scalar.Value(gtx, 0, time.Second, EaseLinear); got != 0 {
		t.Fatalf("initial float = %v", got)
	}
	gtx.Now = start.Add(time.Millisecond)
	scalar.Value(gtx, 1, time.Second, EaseLinear)
	gtx.Now = start.Add(501 * time.Millisecond)
	if got := scalar.Value(gtx, 1, time.Second, EaseLinear); got != .5 {
		t.Fatalf("mid float = %v, want .5", got)
	}
	if !scalar.Ready() || scalar.Target() != 1 || !scalar.Increasing() {
		t.Fatalf("float state = ready %v, target %v, increasing %v", scalar.Ready(), scalar.Target(), scalar.Increasing())
	}

	var colors ColorTransition
	from := color.NRGBA{R: 10, A: 255}
	to := color.NRGBA{R: 110, A: 255}
	gtx.Now = start
	colors.Value(gtx, from, time.Second, EaseLinear)
	gtx.Now = start.Add(time.Millisecond)
	colors.Value(gtx, to, time.Second, EaseLinear)
	gtx.Now = start.Add(501 * time.Millisecond)
	if got := colors.Value(gtx, to, time.Second, EaseLinear); got.R != 60 {
		t.Fatalf("mid color = %#v, want red 60", got)
	}
}
