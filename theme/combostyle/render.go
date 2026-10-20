package combostyle

import (
	"image"
	"image/color"
	"math"
	"sync"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

var chevronIcon struct {
	sync.Once
	value *widget.Icon
}

func renderChevron(gtx layout.Context, size image.Point, progress float32, col color.NRGBA, diameter int) {
	col.A = byte(float32(col.A) * 0.9)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	stack := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
	diameter = min(diameter, min(size.X, size.Y))
	if diameter <= 0 {
		stack.Pop()
		return
	}
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	chevronIcon.Do(func() {
		value, err := widget.NewIcon(lucide.ChevronDown)
		if err == nil {
			chevronIcon.value = value
		}
	})
	if chevronIcon.value != nil {
		chevronIcon.value.Layout(iconGtx, col)
	}
	offset.Pop()
	stack.Pop()
}

func renderCheck(gtx layout.Context, size image.Point, progress float32, foreground color.NRGBA, inset, strokeWidth int) {
	progress = min(max(progress, 0), 1)
	if progress == 0 {
		return
	}
	rect := image.Rectangle{Max: size}.Inset(max(inset, 0))
	if rect.Empty() {
		return
	}
	x, y := float32(rect.Min.X), float32(rect.Min.Y)
	w, h := float32(rect.Dx()), float32(rect.Dy())
	path := draw.CheckPath(gtx.Ops, [3]f32.Point{
		f32.Pt(x+w*0.05, y+h*0.56), f32.Pt(x+w*0.40, y+h*0.86), f32.Pt(x+w*0.95, y+h*0.14),
	}, progress)
	foreground.A = byte(float32(foreground.A)*progress + 0.5)
	stroke := clip.Stroke{Path: path, Width: float32(max(strokeWidth, 1))}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
}

var _ theme.ComboChevronRenderer = func(layout.Context, image.Point, float32, style.ResolvedStyle) {}
var _ theme.ComboCheckRenderer = func(layout.Context, image.Point, float32, style.ResolvedStyle) {}
