package chart

import (
	"image"
	"image/color"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
)

func TestDataWindowGestureZoomsAroundPointerAndPans(t *testing.T) {
	plot := image.Rect(0, 0, 100, 100)
	window := NewDataWindow(0.2, 0.8)
	gesture := new(DataWindowGesture)

	zoomed, changed := gesture.Update(pointer.Event{
		Kind: pointer.Scroll, Position: f32.Pt(50, 50), Scroll: f32.Pt(0, -1),
	}, plot, window, false)
	if !changed || !closeWindow(zoomed, DataWindow{Start: 0.26, End: 0.74}) {
		t.Fatalf("zoomed data window = %#v, changed %v", zoomed, changed)
	}

	gesture.Update(pointer.Event{
		Kind: pointer.Press, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(50, 50),
	}, plot, window, false)
	panned, changed := gesture.Update(pointer.Event{
		Kind: pointer.Drag, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(70, 50),
	}, plot, window, false)
	if !changed || !closeWindow(panned, DataWindow{Start: 0.08, End: 0.68}) {
		t.Fatalf("panned data window = %#v, changed %v", panned, changed)
	}
}

func TestAnnotationsUseValueSemanticsAndValidate(t *testing.T) {
	base := NewMarkLine(AxisY, 10)
	configured := base.Text("Target").Color(color.NRGBA{R: 1, A: 0xff}).Width(2)
	if base.Label != "" || base.width != 0 || configured.Label != "Target" || configured.width != 2 {
		t.Fatalf("mark line values = base %#v configured %#v", base, configured)
	}
	area := NewMarkArea(AxisX, 1, 2).Text("Window")
	point := NewMarkPoint(1, 2).Text("Peak").Size(10)
	ValidateAnnotations([]MarkLine{configured}, []MarkArea{area}, []MarkPoint{point})

	for _, run := range []func(){
		func() { NewMarkLine(Axis(9), 1) },
		func() { NewMarkArea(AxisY, 2, 1) },
		func() { NewMarkPoint(math.NaN(), 1) },
		func() { base.Width(0) },
		func() { point.Size(0) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid chart annotation did not panic")
				}
			}()
			run()
		}()
	}
}

func TestDataWindowValidationAndBounds(t *testing.T) {
	for _, values := range [][2]float32{{-0.1, 1}, {0, 1.1}, {0.5, 0.5}, {0.8, 0.2}, {float32(math.NaN()), 1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewDataWindow(%v, %v) did not panic", values[0], values[1])
				}
			}()
			NewDataWindow(values[0], values[1])
		}()
	}
	if got := (DataWindow{Start: 0.8, End: 1}).pan(0.2); got != (DataWindow{Start: 0.8, End: 1}) {
		t.Fatalf("bounded pan = %#v", got)
	}
	if got := FullDataWindow(); !got.IsFull() {
		t.Fatalf("full data window = %#v", got)
	}
}

func closeWindow(got, want DataWindow) bool {
	return math.Abs(float64(got.Start-want.Start)) < 1e-6 && math.Abs(float64(got.End-want.End)) < 1e-6
}
