package flowui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestBoxAppliesWidthAndPadding(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := Box(child).Width(100).Padding(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size.X != 100 {
		t.Fatalf("width = %d, want 100", dims.Size.X)
	}
	if child.constraints.Min.X != 80 || child.constraints.Max.X != 80 {
		t.Fatalf("child width constraints = %v, want exact 80", child.constraints)
	}
}

func TestBoxAppliesMinMaxConstraints(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).MinWidth(80).MaxWidth(120).MinHeight(40).MaxHeight(90).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if child.constraints.Min.X != 80 || child.constraints.Max.X != 120 {
		t.Fatalf("width constraints = %v, want min 80 max 120", child.constraints)
	}
	if child.constraints.Min.Y != 40 || child.constraints.Max.Y != 90 {
		t.Fatalf("height constraints = %v, want min 40 max 90", child.constraints)
	}
}

func TestBoxFill(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).FillWidth().FillHeight().Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if child.constraints.Min != image.Pt(300, 200) || child.constraints.Max != image.Pt(300, 200) {
		t.Fatalf("constraints = %v, want exact (300,200)", child.constraints)
	}
}

func TestBoxAppliesSidePadding(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).
		PaddingTop(1).
		PaddingRight(2).
		PaddingBottom(3).
		PaddingLeft(4).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         &ops,
		})

	if child.constraints.Max != image.Pt(94, 96) {
		t.Fatalf("child max constraints = %v, want (94,96)", child.constraints.Max)
	}
}

func TestBoxAppliesMarginOutsideSize(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := Box(child).Width(100).Height(50).Margin(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(120, 70) {
		t.Fatalf("box size = %v, want (120,70)", dims.Size)
	}
	if child.constraints.Min != image.Pt(100, 50) || child.constraints.Max != image.Pt(100, 50) {
		t.Fatalf("child constraints = %v, want exact (100,50)", child.constraints)
	}
}

func TestBoxAppliesSideMargin(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).MarginLeft(4).MarginRight(2).MarginTop(1).MarginBottom(3).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	if child.constraints.Max != image.Pt(94, 96) {
		t.Fatalf("child max constraints = %v, want (94,96)", child.constraints.Max)
	}
}

func TestBoxAlignClearsChildMinimum(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).FillWidth().FillHeight().Align(AlignCenter).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if child.constraints.Min != (image.Point{}) {
		t.Fatalf("child min constraints = %v, want zero", child.constraints.Min)
	}
	if child.constraints.Max != image.Pt(300, 200) {
		t.Fatalf("child max constraints = %v, want (300,200)", child.constraints.Max)
	}
}

func TestBoxClipConstrainsReportedSize(t *testing.T) {
	var ops op.Ops

	dims := Box(overflowWidget{}).Width(100).Height(50).Clip().Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(100, 50) {
		t.Fatalf("clipped size = %v, want (100,50)", dims.Size)
	}
}

func TestBoxOverflowVisible(t *testing.T) {
	b := Box(Spacer(10, 10)).Overflow(OverflowVisible)

	if b.overflow != OverflowVisible {
		t.Fatal("overflow was not visible")
	}
}

func TestBoxOverflowHidden(t *testing.T) {
	b := Box(Spacer(10, 10)).Overflow(OverflowHidden)

	if b.overflow != OverflowHidden {
		t.Fatal("overflow was not hidden")
	}
}

func TestRowColumnAlignment(t *testing.T) {
	if Row().AlignMiddle().align != layout.Middle {
		t.Fatal("row alignment was not middle")
	}
	if Column().AlignEnd().align != layout.End {
		t.Fatal("column alignment was not end")
	}
}

func TestExpandedUsesRemainingRowSpace(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Row(
		Spacer(50, 10),
		Expanded(child),
	).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 40)},
		Ops:         &ops,
	})

	if child.constraints.Min.X != 140 || child.constraints.Max.X != 140 {
		t.Fatalf("expanded width constraints = %v, want exact 140", child.constraints)
	}
}

func TestFlexibleSplitsRemainingSpaceByWeight(t *testing.T) {
	left := &constraintWidget{}
	right := &constraintWidget{}
	var ops op.Ops

	Row(
		Flexible(1, left),
		Flexible(2, right),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 40)},
		Ops:         &ops,
	})

	if left.constraints.Min.X != 100 || left.constraints.Max.X != 100 {
		t.Fatalf("left constraints = %v, want exact 100", left.constraints)
	}
	if right.constraints.Min.X != 200 || right.constraints.Max.X != 200 {
		t.Fatalf("right constraints = %v, want exact 200", right.constraints)
	}
}

func TestWrapMovesChildrenToNextLine(t *testing.T) {
	var ops op.Ops

	dims := Wrap(
		Spacer(60, 10),
		Spacer(60, 20),
		Spacer(60, 30),
	).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(130, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(130, 60) {
		t.Fatalf("wrap size = %v, want (130,60)", dims.Size)
	}
}

func TestWrapAlignmentOffset(t *testing.T) {
	if Wrap().AlignEnd().align != layout.End {
		t.Fatal("wrap alignment was not end")
	}
	if got := alignmentOffset(layout.Middle, 11); got != 5 {
		t.Fatalf("middle offset = %d, want 5", got)
	}
	if got := alignmentOffset(layout.End, 11); got != 11 {
		t.Fatalf("end offset = %d, want 11", got)
	}
	if got := alignmentOffset(layout.Start, 11); got != 0 {
		t.Fatalf("start offset = %d, want 0", got)
	}
}

func TestAdaptiveReceivesAvailableSize(t *testing.T) {
	var got ViewSize
	var ops op.Ops

	Adaptive(func(size ViewSize) Widget {
		got = size
		return Spacer(10, 10)
	}).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(400, 200)},
		Metric:      unit.Metric{PxPerDp: 2, PxPerSp: 2},
		Ops:         &ops,
	})

	if got.Width != 200 || got.Height != 100 {
		t.Fatalf("size = %v, want (200dp,100dp)", got)
	}
}

func TestSpacer(t *testing.T) {
	dims := Spacer(12, 8).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
	})
	if dims.Size != image.Pt(12, 8) {
		t.Fatalf("spacer size = %v, want (12,8)", dims.Size)
	}
}

type constraintWidget struct {
	constraints layout.Constraints
}

func (w *constraintWidget) Layout(_ *Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

type overflowWidget struct{}

func (overflowWidget) Layout(_ *Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(500, 500)}
}
