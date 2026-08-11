package menubar

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawMenubarTrigger(gtx layout.Context, tokens theme.MenubarTheme, size image.Point, style menubarTriggerStyle, focusOpacity float32) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	radius := min(max(gtx.Dp(tokens.TriggerRadius), 0), min(size.X, size.Y)/2)
	rect := image.Rectangle{Max: size}
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focusOpacity <= 0 || style.focus.A == 0 {
		return
	}
	width := max(gtx.Dp(tokens.TriggerFocusRingWidth), 1)
	inset := max(gtx.Dp(tokens.TriggerFocusRingOffset), 0) + (width+1)/2
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	focusRadius := max(radius-inset, 0)
	col := style.focus
	col.A = byte(float32(col.A)*focusOpacity + 0.5)
	render.DrawRoundedInsetStroke(gtx, rect, focusRadius, width, inset, col)
}
