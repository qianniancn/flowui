package layout

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/style"
)

func TestPageCentersAndPadsContent(t *testing.T) {
	probe := &constraintWidget{}
	dims := Page(probe).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(1000, 800)},
		Ops:         new(op.Ops),
	})
	if dims.Size.X != 760 {
		t.Fatalf("page width = %d, want the 760 max width", dims.Size.X)
	}
	// 760 max width minus 24 padding on both sides.
	if probe.constraints.Max.X != 712 {
		t.Fatalf("content max width = %d, want 712", probe.constraints.Max.X)
	}
}

func TestPageStyleOverridesDefaults(t *testing.T) {
	probe := &constraintWidget{}
	Page(probe).Style(style.New().MaxWidth(400).Padding(8)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(1000, 800)},
		Ops:         new(op.Ops),
	})
	if probe.constraints.Max.X != 384 {
		t.Fatalf("content max width = %d, want 384", probe.constraints.Max.X)
	}
}
