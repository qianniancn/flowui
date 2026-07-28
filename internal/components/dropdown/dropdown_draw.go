package dropdown

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawDropdownTriggerFocus(gtx layout.Context, activeTheme *theme.Theme, size image.Point, opacity float32) {
	if opacity <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	tokens := activeTheme.Components.Dropdown
	width := max(gtx.Dp(tokens.TriggerFocusRingWidth), 1)
	offset := max(gtx.Dp(tokens.TriggerFocusRingOffset), 0)
	rect, radius := dropdownFocusRingGeometry(image.Rectangle{Max: size}, gtx.Dp(tokens.TriggerFocusRadius), width, offset)
	if rect.Empty() {
		return
	}
	col := theme.ColorOr(tokens.FocusColor, activeTheme.Palette.Focus)
	col.A = byte(float32(col.A)*opacity + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func dropdownFocusRingGeometry(rect image.Rectangle, radius, width, offset int) (image.Rectangle, int) {
	inset := max(offset, 0) + (max(width, 1)+1)/2
	focusRect := rect.Inset(inset)
	return focusRect, max(radius-inset, 0)
}
