package flowui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func drawDatePickerPopover(gtx layout.Context, theme *Theme, rect image.Rectangle, radius int) {
	drawPopupSurface(gtx, theme, rect, radius)
}

func drawDatePickerNavButton(gtx layout.Context, size image.Point, delta int, style datePickerNavButtonStyle) {
	if style.bg.A != 0 {
		paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(image.Rectangle{Max: size}, min(size.X, size.Y)/2).Op(gtx.Ops))
	}
	drawDatePickerChevron(gtx, style.theme, size, delta, style.fg)
}

func drawDatePickerYearPickerIndicator(gtx layout.Context, theme *Theme, size image.Point, open bool, col color.NRGBA) {
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	width := max(dpFloat(gtx, theme.Components.DatePicker.NavChevronStroke), 1)
	var path clip.Path
	path.Begin(gtx.Ops)
	if open {
		path.MoveTo(f32.Pt(center.X-4, center.Y-2))
		path.LineTo(f32.Pt(center.X, center.Y+2))
		path.LineTo(f32.Pt(center.X+4, center.Y-2))
	} else {
		path.MoveTo(f32.Pt(center.X-2, center.Y-4))
		path.LineTo(f32.Pt(center.X+2, center.Y))
		path.LineTo(f32.Pt(center.X-2, center.Y+4))
	}
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: width,
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawDatePickerChevron(gtx layout.Context, theme *Theme, size image.Point, delta int, col color.NRGBA) {
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	dir := float32(delta)
	width := max(dpFloat(gtx, theme.Components.DatePicker.NavChevronStroke), 1)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X-dir*2, center.Y-5))
	path.LineTo(f32.Pt(center.X+dir*3, center.Y))
	path.LineTo(f32.Pt(center.X-dir*2, center.Y+5))
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: width,
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawDatePickerCell(gtx layout.Context, size image.Point, style datePickerCellStyle) {
	if style.bg.A == 0 {
		return
	}
	radius := min(size.X, size.Y) / 2
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func drawDatePickerStrike(gtx layout.Context, theme *Theme, size image.Point, col color.NRGBA) {
	col.A = byte(uint16(col.A) * 3 / 4)
	var path clip.Path
	path.Begin(gtx.Ops)
	y := float32(size.Y) / 2
	halfSize := dpFloat(gtx, theme.Components.DatePicker.CellStrikeHalfSize)
	path.MoveTo(f32.Pt(float32(size.X)/2-halfSize, y))
	path.LineTo(f32.Pt(float32(size.X)/2+halfSize, y))
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: max(dpFloat(gtx, theme.Components.DatePicker.CellStrikeWidth), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawDatePickerCalendarIcon(gtx layout.Context, theme *Theme, size image.Point, col color.NRGBA) {
	rect := image.Rect(1, 2, size.X-1, size.Y-1)
	radius := max(gtx.Dp(theme.Components.DatePicker.IconRadius), 1)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, radius).Path(gtx.Ops),
		Width: max(dpFloat(gtx, theme.Components.DatePicker.IconStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()

	top := image.Rect(rect.Min.X+1, rect.Min.Y+4, rect.Max.X-1, rect.Min.Y+5)
	paint.FillShape(gtx.Ops, col, clip.Rect(top).Op())
	for _, x := range []int{rect.Min.X + 4, rect.Max.X - 4} {
		line := image.Rect(x, rect.Min.Y-1, x+1, rect.Min.Y+3)
		paint.FillShape(gtx.Ops, col, clip.Rect(line).Op())
	}
}

func datePickerNavStyle(theme *Theme, hovered, pressed, disabled bool) datePickerNavButtonStyle {
	style := datePickerNavButtonStyle{
		theme: theme,
		fg:    theme.Palette.AccentSoftForeground,
	}
	if hovered || pressed {
		style.bg = theme.Palette.SurfaceRaised
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.fg = theme.DisabledColor(style.fg)
	}
	return style
}

func datePickerCellStyleFor(theme *Theme, hovered, pressed, selected, today, outside, disabled bool) datePickerCellStyle {
	style := datePickerCellStyle{
		fg: theme.Palette.Foreground,
	}
	if hovered || pressed {
		style.bg = theme.Palette.SurfaceRaised
	}
	if today {
		style.bg = theme.Palette.AccentSoft
		style.fg = theme.Palette.AccentSoftForeground
	}
	if selected {
		style.bg = theme.Palette.Accent
		style.fg = theme.Palette.AccentForeground
	}
	if outside {
		style.fg = theme.DisabledColor(theme.Palette.MutedForeground)
		if selected {
			style.bg = theme.Palette.SurfaceRaised
		}
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.fg = theme.DisabledColor(style.fg)
	}
	return style
}

type datePickerNavButtonStyle struct {
	theme *Theme
	bg    color.NRGBA
	fg    color.NRGBA
}

type datePickerCellStyle struct {
	bg color.NRGBA
	fg color.NRGBA
}
