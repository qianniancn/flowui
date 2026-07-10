package frame

import (
	"image"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func TestDeferOverlayRecordsInputOps(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	}
	tag := new(int)

	DeferOverlay(New(nil, nil, locale.LanguageAuto), gtx, func(gtx layout.Context) layout.Dimensions {
		event.Op(gtx.Ops, tag)
		return layout.Dimensions{Size: image.Pt(10, 10)}
	})

	var router input.Router
	router.Frame(&ops)
}
