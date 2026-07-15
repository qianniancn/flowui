package listbox

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawListBoxItem(gtx layout.Context, theme *theme.Theme, size image.Point, style listBoxItemStyle) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Components.ListBox.ItemRadius), 1), min(size.X, size.Y)/2)
	optionrow.DrawBackground(gtx, size, theme.Components.ListBox.ItemRadius, style.bg)
	drawListBoxFocus(gtx, theme, rect, radius, style)
}

func drawListBoxFocus(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int, style listBoxItemStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(theme.Components.ListBox.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawListBoxIndicator(gtx layout.Context, theme *theme.Theme, size image.Point, style listBoxItemStyle) {
	optionrow.DrawCheck(gtx, size, style.selected, theme.Components.ListBox.ItemIndicatorInset, theme.Components.ListBox.ItemIndicatorStroke, style.indicator)
}
