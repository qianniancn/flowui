package chart

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestLegendItemClicksWithoutEnteringFocusOrder(t *testing.T) {
	router := new(input.Router)
	item := new(LegendItem)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(40, 24)),
		Source:      router.Source(),
		Ops:         &ops,
	}
	item.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(40, 24)}
	})
	router.Frame(&ops)

	router.MoveFocus(key.FocusForward)
	if router.Source().Focused(item) {
		t.Fatal("legend item entered keyboard focus order")
	}

	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(10, 10)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(10, 10)},
	)
	if !item.Clicked(gtx) {
		t.Fatal("legend item did not report a mouse click")
	}
}
