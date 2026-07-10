package field

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func TestDisabledStateClearsHover(t *testing.T) {
	state := &State{Hovered: true}
	state.Update(frame.New(nil, nil, locale.LanguageAuto), fieldTestContext(), true, new(int))
	if state.Hovered {
		t.Fatal("disabled field kept hover state")
	}
}

func TestBackgroundTransition(t *testing.T) {
	state := new(State)
	start := time.Unix(1, 0)
	from := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	to := color.NRGBA{R: 110, G: 120, B: 130, A: 255}
	gtx := fieldTestContext()
	gtx.Now = start

	if got := state.Background(gtx, from); got != from {
		t.Fatalf("initial background = %v, want %v", got, from)
	}
	if got := state.Background(gtx, to); got != from {
		t.Fatalf("transition start = %v, want %v", got, from)
	}

	gtx.Now = start.Add(colorDuration / 2)
	mid := state.Background(gtx, to)
	if mid == from || mid == to {
		t.Fatalf("transition midpoint = %v, want between %v and %v", mid, from, to)
	}

	gtx.Now = start.Add(colorDuration)
	if got := state.Background(gtx, to); got != to {
		t.Fatalf("transition end = %v, want %v", got, to)
	}
}

func fieldTestContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}
