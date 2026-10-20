package selectstyle

import (
	"image"
	"math"
	"sync"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/internal/textrender"
)

var selectIndicatorIcon struct {
	sync.Once
	value *widget.Icon
}

func indicatorRenderer(iconSize, stroke unit.Dp) theme.SelectIndicatorRenderer {
	return func(gtx layout.Context, size image.Point, progress float32, resolved style.ResolvedStyle) layout.Dimensions {
		col := textrender.Color(resolved)
		diameter := max(gtx.Dp(iconSize), gtx.Dp(stroke*12))
		diameter = min(diameter, min(size.X, size.Y))
		if diameter <= 0 {
			return layout.Dimensions{Size: size}
		}
		selectIndicatorIcon.Do(func() {
			value, err := widget.NewIcon(lucide.ChevronDown)
			if err == nil {
				selectIndicatorIcon.value = value
			}
		})
		if selectIndicatorIcon.value == nil {
			return layout.Dimensions{Size: size}
		}
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
		rotation := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
		offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
		dims := selectIndicatorIcon.value.Layout(iconGtx, col)
		offset.Pop()
		rotation.Pop()
		return dims
	}
}

func textRenderer(active *theme.Theme) theme.SelectTextRenderer {
	return func(gtx layout.Context, request theme.SelectTextRequest) layout.Dimensions {
		return textrender.Label(active, gtx, request.Text, request.Style, 0)
	}
}

var _ theme.SelectTextRenderer = func(layout.Context, theme.SelectTextRequest) layout.Dimensions { return layout.Dimensions{} }
