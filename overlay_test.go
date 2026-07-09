package flowui

import (
	"image"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestDeferOverlayRecordsInputOps(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	}
	tag := new(int)

	newContext(nil).deferOverlay(gtx, func(gtx layout.Context) layout.Dimensions {
		event.Op(gtx.Ops, tag)
		return layout.Dimensions{Size: image.Pt(10, 10)}
	})

	var router input.Router
	router.Frame(&ops)
}
