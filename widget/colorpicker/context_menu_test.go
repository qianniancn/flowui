package colorpicker

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/uitest"
)

// TestContextMenuInsidePickerStaysOpen pins that a right-click edit menu opened
// on a field inside the picker popup remains open, and next to the pointer,
// across frames.
//
// Two separate defects used to break this. Disabling the picker panel when a
// nested overlay owned policy input unregistered the menu on every other frame
// (the flicker). And the panel's flex rows did not record child offsets, so the
// pointer position was resolved against the panel origin (the menu appeared far
// from the input).
func TestContextMenuInsidePickerStaysOpen(t *testing.T) {
	for _, target := range []string{"Color", "R"} {
		t.Run(target, func(t *testing.T) {
			h, root, field := openPickerWithContextMenu(t, target)
			for frame := range 24 {
				h.Advance(16 * time.Millisecond)
				h.Frame(root)
				if !core.HasVisibleOverlay(h.Context()) {
					t.Fatalf("frame %d: context menu overlay disappeared", frame)
				}
				if !uitest.HasSemanticLabel(h.Semantics(), "Select All") {
					t.Fatalf("frame %d: context menu items disappeared", frame)
				}
				if !uitest.HasSemanticLabel(h.Semantics(), "Color area") {
					t.Fatalf("frame %d: picker panel disappeared", frame)
				}
			}
			cut, ok := uitest.SemanticBoundsForLabel(h.Semantics(), "Cut")
			if !ok {
				t.Fatalf("menu item not found")
			}
			center := image.Pt(field.Min.X+field.Dx()/2, field.Min.Y+field.Dy()/2)
			if abs(cut.Min.X-center.X) > 120 || abs(cut.Min.Y-center.Y) > 120 {
				t.Fatalf("menu item at %v is far from pointer %v", cut.Min, center)
			}
		})
	}
}

// TestPickerContextMenuDoesNotFlicker reproduces the pixel-level flicker: with
// the menu open, consecutive rendered frames must not alternate between menu
// and no menu. Headless rendering is skipped when no GPU backend is available.
func TestPickerContextMenuDoesNotFlicker(t *testing.T) {
	size := image.Pt(720, 900)
	renderer, err := headless.NewWindow(size.X, size.Y)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer renderer.Release()

	h, pickerRoot, _ := openPickerWithContextMenu(t, "Color")
	var ops *op.Ops
	root := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		ops = gtx.Ops
		active := core.ActiveTheme(ctx)
		paint.Fill(gtx.Ops, active.Semantic().Background)
		return pickerRoot.Layout(ctx, gtx)
	})

	previous := image.NewRGBA(image.Rectangle{Max: size})
	for frame := range 14 {
		h.Advance(16 * time.Millisecond)
		h.Frame(root)
		if err := renderer.Frame(ops); err != nil {
			t.Fatal(err)
		}
		current := image.NewRGBA(image.Rectangle{Max: size})
		if err := renderer.Screenshot(current); err != nil {
			t.Fatal(err)
		}
		diff := pixelDiff(previous, current)
		if frame >= 8 && diff > 3000 {
			t.Fatalf("frame %d: %d pixels changed; menu is flickering", frame, diff)
		}
		previous = current
	}
}

// openPickerWithContextMenu opens a picker, right-clicks the field with the
// given semantic label, and waits for the edit menu to become visible.
func openPickerWithContextMenu(t *testing.T, target string) (*uitest.Harness, core.Widget, image.Rectangle) {
	t.Helper()
	active := base.New(theme.DefaultSeed())
	h := uitest.NewWithConfig(uitest.Config{
		Size:   image.Pt(720, 900),
		Theme:  &active,
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
	})
	picker := New("picker", color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}).
		Label("Pick").
		ShowField().
		ShowHistory(false)
	root := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return flowlayout.Box(flowlayout.Column(picker)).Style(style.New().Padding(24)).Layout(ctx, gtx)
	})
	h.Frame(root)

	trigger, ok := uitest.SemanticBoundsForLabel(h.Semantics(), "Pick")
	if !ok {
		t.Fatalf("trigger not found: %v", uitest.SemanticLabels(h.Semantics()))
	}
	h.Click(center(trigger))
	h.Frame(root)
	h.Advance(400 * time.Millisecond)
	h.Frame(root)
	if !core.HasVisibleOverlay(h.Context()) {
		t.Fatalf("picker popup did not open")
	}

	field, ok := uitest.SemanticBoundsForLabel(h.Semantics(), target)
	if !ok {
		t.Fatalf("%s field not found: %v", target, uitest.SemanticLabels(h.Semantics()))
	}
	h.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonSecondary, Position: center(field)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: center(field)},
	)
	h.Frame(root)
	for frame := 0; frame < 6 && !uitest.HasSemanticLabel(h.Semantics(), "Select All"); frame++ {
		h.Frame(root)
		h.Advance(50 * time.Millisecond)
		h.Frame(root)
	}
	if !uitest.HasSemanticLabel(h.Semantics(), "Select All") {
		t.Fatalf("context menu did not open: %v", uitest.SemanticLabels(h.Semantics()))
	}
	return h, root, field
}

func center(r image.Rectangle) f32.Point {
	return f32.Pt(float32(r.Min.X+r.Dx()/2)+0.5, float32(r.Min.Y+r.Dy()/2)+0.5)
}

func pixelDiff(a, b *image.RGBA) int {
	diff := 0
	for i := 0; i < len(a.Pix); i += 4 {
		if a.Pix[i] != b.Pix[i] || a.Pix[i+1] != b.Pix[i+1] || a.Pix[i+2] != b.Pix[i+2] {
			diff++
		}
	}
	return diff
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
