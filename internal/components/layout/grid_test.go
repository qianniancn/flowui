package layoutui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestGridLaysOutFixedColumns(t *testing.T) {
	a := &cellWidget{height: 10}
	b := &cellWidget{height: 20}
	c := &cellWidget{height: 30}
	d := &cellWidget{height: 40}
	var ops op.Ops

	dims := Grid(3, a, b, c, d).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(320, 80) {
		t.Fatalf("grid size = %v, want (320,80)", dims.Size)
	}
	if a.constraints.Min.X != 100 || a.constraints.Max.X != 100 {
		t.Fatalf("cell constraints = %v, want exact width 100", a.constraints)
	}
}

func TestAutoGridComputesColumnsFromMinimumWidth(t *testing.T) {
	a := &cellWidget{height: 10}
	b := &cellWidget{height: 10}
	c := &cellWidget{height: 10}
	var ops op.Ops

	AutoGrid(90, a, b, c).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(290, 100)},
		Ops:         &ops,
	})

	if a.constraints.Min.X != 90 || a.constraints.Max.X != 90 {
		t.Fatalf("cell constraints = %v, want exact width 90", a.constraints)
	}
}

type cellWidget struct {
	constraints layout.Constraints
	height      int
}

func (w *cellWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	return layout.Dimensions{Size: image.Pt(gtx.Constraints.Min.X, w.height)}
}
