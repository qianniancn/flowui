package collapsiblestyle_test

import (
	"image"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/theme/collapsiblestyle"
)

// TestIndicatorStaysInsideItsSize checks that the chevron is painted inside the
// box the renderer reports. The trigger flex places that box at the end of the
// row, so centering the icon again inside the leftover constraints moved it
// towards the middle of the trigger.
func TestIndicatorStaysInsideItsSize(t *testing.T) {
	active := base.Light()
	visual := collapsiblestyle.BuildCollapsible(&active, theme.CollapsibleRequest{})
	if visual.Indicator == nil {
		t.Fatal("no indicator renderer")
	}
	window, err := headless.NewWindow(400, 40)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	for _, progress := range []float32{0, .5, 1} {
		var ops op.Ops
		gtx := layout.Context{
			Ops:         &ops,
			Constraints: layout.Constraints{Max: image.Pt(400, 40)},
		}
		dims := visual.Indicator(gtx, theme.CollapsibleIndicatorRequest{Progress: progress})
		if dims.Size.X != dims.Size.Y || dims.Size.X <= 0 {
			t.Fatalf("indicator size = %v, want a square box", dims.Size)
		}
		if err := window.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		pixels := image.NewRGBA(image.Rectangle{Max: image.Pt(400, 40)})
		if err := window.Screenshot(pixels); err != nil {
			t.Fatal(err)
		}
		painted := image.Rectangle{}
		for y := range 40 {
			for x := range 400 {
				if pixels.RGBAAt(x, y).A != 0 {
					painted = painted.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
		box := image.Rectangle{Max: dims.Size}
		if painted.Empty() {
			t.Fatalf("progress %.1f painted nothing", progress)
		}
		if !painted.In(box) {
			t.Fatalf("progress %.1f painted at %v, outside its %v box", progress, painted, box)
		}
	}
}
