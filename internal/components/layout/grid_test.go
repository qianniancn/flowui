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

func TestGridPropagatesCellPosition(t *testing.T) {
	probe := &overlayProbeWidget{
		key:    "grid",
		size:   image.Pt(10, 10),
		anchor: image.Rect(0, 0, 10, 10),
	}
	var got image.Rectangle
	probe.got = &got
	ctx := newContext(nil)
	frame.BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(100, 100)}, Ops: new(op.Ops)}
	Grid(2,
		Spacer(10, 10),
		Spacer(10, 20),
		probe,
	).ColumnGap(10).RowGap(5).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)

	want := image.Rect(0, 25, 10, 35)
	if got != want {
		t.Fatalf("grid anchor = %v, want %v", got, want)
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
