package datepicker

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/components/icon"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawDatePickerPopover(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int) {
	render.DrawSurface(gtx, rect, radius, theme.Palette.OverlayColor(), render.ThemeShadow(theme.Shadows.Overlay, theme.Palette.OverlayShadowColor(), 1))
}

func drawDatePickerNavButton(gtx layout.Context, size image.Point, delta int, style datePickerNavButtonStyle) {
	if style.bg.A != 0 {
		paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(image.Rectangle{Max: size}, min(size.X, size.Y)/2).Op(gtx.Ops))
	}
	drawDatePickerChevron(gtx, style.theme, size, delta, style.fg)
}

func drawDatePickerYearPickerIndicator(gtx layout.Context, size image.Point, open bool, col color.NRGBA) {
	data := lucide.ChevronRight
	if open {
		data = lucide.ChevronDown
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(data, iconGtx, col)
}

func drawDatePickerChevron(gtx layout.Context, theme *theme.Theme, size image.Point, delta int, col color.NRGBA) {
	data := lucide.ChevronRight
	if delta < 0 {
		data = lucide.ChevronLeft
	}
	diameter := min(icon.LucideSizeForStroke(gtx, theme.Components.DatePicker.NavChevronStroke), min(size.X, size.Y))
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(data, iconGtx, col)
	offset.Pop()
}

func drawDatePickerCell(gtx layout.Context, size image.Point, style datePickerCellStyle) {
	if style.bg.A == 0 {
		return
	}
	radius := min(size.X, size.Y) / 2
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func drawDatePickerRangeTrack(gtx layout.Context, col color.NRGBA, size image.Point, startRadius, endRadius int) {
	if size.X <= 0 || size.Y <= 4 {
		return
	}
	rect := image.Rect(0, 2, size.X, size.Y-2)
	startRadius = min(max(startRadius, 0), rect.Dy()/2)
	endRadius = min(max(endRadius, 0), rect.Dy()/2)
	paint.FillShape(gtx.Ops, col, clip.RRect{
		Rect: rect,
		NW:   startRadius,
		SW:   startRadius,
		NE:   endRadius,
		SE:   endRadius,
	}.Op(gtx.Ops))
}

func drawDatePickerStrike(gtx layout.Context, theme *theme.Theme, size image.Point, col color.NRGBA) {
	col.A = byte(uint16(col.A) * 3 / 4)
	var path clip.Path
	path.Begin(gtx.Ops)
	y := float32(size.Y) / 2
	halfSize := render.DpFloat(gtx, theme.Components.DatePicker.CellStrikeHalfSize)
	path.MoveTo(f32.Pt(float32(size.X)/2-halfSize, y))
	path.LineTo(f32.Pt(float32(size.X)/2+halfSize, y))
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: max(render.DpFloat(gtx, theme.Components.DatePicker.CellStrikeWidth), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawDatePickerCalendarIcon(gtx layout.Context, size image.Point, col color.NRGBA) {
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(lucide.Calendar, iconGtx, col)
}

func drawDatePickerTriggerFocus(gtx layout.Context, activeTheme *theme.Theme, size image.Point, visible bool) {
	if !visible || size.X <= 0 || size.Y <= 0 {
		return
	}
	diameter := min(gtx.Dp(activeTheme.Components.DatePicker.IconSize+8), min(size.X, size.Y))
	rect := image.Rect((size.X-diameter)/2, (size.Y-diameter)/2, (size.X+diameter)/2, (size.Y+diameter)/2)
	drawDatePickerFocusRing(gtx, activeTheme, rect, gtx.Dp(activeTheme.Components.DatePicker.SegmentRadius), true)
}

func drawDatePickerControlFocus(gtx layout.Context, activeTheme *theme.Theme, size image.Point, radius int, visible bool) {
	if !visible || size.X <= 0 || size.Y <= 0 {
		return
	}
	drawDatePickerFocusRing(gtx, activeTheme, image.Rectangle{Max: size}, radius, true)
}

func drawDatePickerFocusRing(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, visible bool) {
	if !visible {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Input.FocusRingWidth), 1)
	rect = rect.Inset((width + 1) / 2)
	if rect.Empty() {
		return
	}
	radius = max(radius-(width+1)/2, 0)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, radius).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, activeTheme.Palette.Focus)
	stroke.Pop()
}

func datePickerNavStyle(theme *theme.Theme, hovered, pressed, disabled bool) datePickerNavButtonStyle {
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

func datePickerCellStyleFor(theme *theme.Theme, hovered, pressed, selected, today, outside, disabled bool) datePickerCellStyle {
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
	theme *theme.Theme
	bg    color.NRGBA
	fg    color.NRGBA
}

type datePickerCellStyle struct {
	bg color.NRGBA
	fg color.NRGBA
}
