package checkbox

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// ControlOptions describes the shared Checkbox control used by Checkbox and Table.
type ControlOptions struct {
	Variant         CheckboxVariant
	Selection       float32
	Indeterminate   bool
	Hovered         bool
	Pressed         bool
	Focused         float32
	Disabled        bool
	Invalid         bool
	CustomIndicator bool
	Indicator       frame.Widget
}

// DrawControl renders a HeroUI-aligned checkbox control without owning interaction state.
func DrawControl(ctx *frame.Context, gtx layout.Context, options ControlOptions) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	style := checkboxStyleFor(activeTheme, options.Variant, options.Hovered, options.Pressed, options.Disabled, options.Invalid)
	style.selected = min(max(options.Selection, 0), 1)
	style.focus = min(max(options.Focused, 0), 1)
	return drawCheckbox(ctx, gtx, activeTheme, style, options.Indeterminate, options.CustomIndicator, options.Indicator)
}

func drawCheckbox(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, style checkboxStyle, indeterminate, customIndicator bool, indicator frame.Widget) layout.Dimensions {
	size := gtx.Dp(activeTheme.Components.Checkbox.Size)
	focusSpace := max(gtx.Dp(activeTheme.Components.Checkbox.FocusSpace), 1)
	maxSize := min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y) - focusSpace*2
	size = min(size, max(maxSize, 0))
	bounds := image.Pt(size+focusSpace*2, size+focusSpace*2)
	dims := gtx.Constraints.Constrain(bounds)
	if size <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-size)/2, (dims.Y-size)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(image.Pt(size, size)),
	}
	radius := min(max(gtx.Dp(activeTheme.Shape.CheckboxRadius), 1), size/2)

	drawCheckboxFocus(gtx, activeTheme, rect, radius, style)
	drawCheckboxFrame(gtx, activeTheme, rect, radius, style)
	drawCheckboxFill(gtx, rect, radius, style)
	if !customIndicator {
		if indeterminate {
			drawCheckboxIndeterminate(gtx, activeTheme, rect, style)
		} else {
			drawCheckboxCheck(gtx, activeTheme, rect, style)
		}
	} else if indicator != nil {
		layoutCheckboxIndicator(ctx, gtx, activeTheme, rect, style, indicator)
	}

	return layout.Dimensions{Size: dims}
}

func drawCheckboxFrame(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int, style checkboxStyle) {
	if style.shadow > 0 {
		render.DrawShadow(gtx, rect, render.RoundedShadowCorners(theme.Shape.CheckboxRadius, theme.Shape.CheckboxRadius, theme.Shape.CheckboxRadius, theme.Shape.CheckboxRadius), render.ThemeShadow(theme.Shadows.Checkbox, theme.Palette.Shadow, style.shadow))
	}
	border := style.border
	border.A = byte(float32(border.A)*(1-style.selected) + 0.5)
	if border.A == 0 {
		paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
		return
	}

	width := max(gtx.Dp(theme.Components.Checkbox.BorderWidth), 1)
	paint.FillShape(gtx.Ops, border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(inner, max(radius-width, 0)).Op(gtx.Ops))
}

func drawCheckboxFocus(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style checkboxStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Checkbox.FocusRingWidth), 1)
	focusRect := rect.Inset(-max(width/2, 1))
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, radius+width).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawCheckboxFill(gtx layout.Context, rect image.Rectangle, radius int, style checkboxStyle) {
	if style.selected == 0 {
		return
	}
	scale := 0.7 + 0.3*style.selected
	size := rect.Size()
	center := f32.Pt(
		float32(rect.Min.X)+float32(size.X)/2,
		float32(rect.Min.Y)+float32(size.Y)/2,
	)
	stack := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
	col := style.accent
	col.A = byte(float32(col.A)*style.selected + 0.5)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	stack.Pop()
}

func drawCheckboxCheck(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, style checkboxStyle) {
	if style.selected == 0 {
		return
	}
	width := float32(rect.Dx())
	height := float32(rect.Dy())
	x := float32(rect.Min.X)
	y := float32(rect.Min.Y)
	points := [3]f32.Point{
		f32.Pt(x+width*0.27, y+height*0.52),
		f32.Pt(x+width*0.43, y+height*0.68),
		f32.Pt(x+width*0.75, y+height*0.31),
	}
	path := render.CheckPath(gtx.Ops, points, style.selected)
	col := style.accentFg
	col.A = byte(float32(col.A)*style.selected + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(render.DpFloat(gtx, theme.Components.Checkbox.CheckStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawCheckboxIndeterminate(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, style checkboxStyle) {
	if style.selected == 0 {
		return
	}
	length := max(int(float32(rect.Dx())*0.58*style.selected+0.5), 1)
	thickness := max(gtx.Dp(activeTheme.Components.Checkbox.IndeterminateStroke), 1)
	center := rect.Min.Add(image.Pt(rect.Dx()/2, rect.Dy()/2))
	line := image.Rect(center.X-length/2, center.Y-thickness/2, center.X+(length+1)/2, center.Y+(thickness+1)/2)
	col := style.accentFg
	col.A = byte(float32(col.A)*style.selected + 0.5)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(line, max(thickness/2, 1)).Op(gtx.Ops))
}

func layoutCheckboxIndicator(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, style checkboxStyle, indicator frame.Widget) {
	diameter := min(gtx.Dp(activeTheme.Components.Checkbox.IndicatorSize), min(rect.Dx(), rect.Dy()))
	if diameter <= 0 {
		return
	}
	indicatorRect := image.Rectangle{Min: image.Pt((rect.Min.X+rect.Max.X-diameter)/2, (rect.Min.Y+rect.Max.Y-diameter)/2)}
	indicatorRect.Max = indicatorRect.Min.Add(image.Pt(diameter, diameter))
	foreground := style.fg
	background := style.bg
	if style.selected > 0 {
		foreground = style.accentFg
		background = style.accent
	}
	restore := frame.PushColors(ctx, foreground, background)
	defer restore()
	stack := op.Offset(indicatorRect.Min).Push(gtx.Ops)
	indicatorGtx := gtx
	indicatorGtx.Constraints = layout.Exact(indicatorRect.Size())
	clipStack := clip.Rect{Max: indicatorRect.Size()}.Push(gtx.Ops)
	layout.Center.Layout(indicatorGtx, func(gtx layout.Context) layout.Dimensions {
		return indicator.Layout(ctx, gtx)
	})
	clipStack.Pop()
	stack.Pop()
}
