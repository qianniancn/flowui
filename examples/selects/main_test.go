package main

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	ui "github.com/qianniancn/FlowUI"
)

func TestExampleViewLayouts(t *testing.T) {
	theme := ui.DefaultTheme()
	ctx := &ui.Context{Theme: &theme}
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(920, 760)},
		Source:      router.Source(),
		Ops:         &ops,
	}

	dims := View(ctx, Model{}, func(Msg) {}).Layout(ctx, gtx)

	if dims.Size.X <= 0 || dims.Size.Y <= 0 {
		t.Fatalf("example dimensions = %v, want non-empty", dims.Size)
	}
}
