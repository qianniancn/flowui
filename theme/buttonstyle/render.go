package buttonstyle

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// spinnerRenderer paints the shared loading indicator geometry. It lives in
// the theme side of the dependency graph so widget/button only schedules the
// provider-declared renderer.
func spinnerRenderer(gtx layout.Context, request theme.ButtonSpinnerRequest) layout.Dimensions {
	d := max(gtx.Dp(request.Size), 1)
	strokeWidth := float32(max(gtx.Dp(request.StrokeWidth), 1))
	radius := float32(d)/2 - strokeWidth/2
	if radius <= 0 {
		return layout.Dimensions{Size: image.Pt(d, d)}
	}

	phase := float32(0)
	if request.Period > 0 && !gtx.Now.IsZero() {
		elapsed := gtx.Now.UnixNano() % int64(request.Period)
		if elapsed < 0 {
			elapsed += int64(request.Period)
		}
		phase = float32(elapsed) / float32(request.Period) * float32(math.Pi*2)
	}
	const segments = 12
	const sweep = float32(math.Pi * 1.6)
	center := f32.Pt(float32(d)/2, float32(d)/2)
	var path clip.Path
	path.Begin(gtx.Ops)
	for index := range segments + 1 {
		angle := phase + sweep*float32(index)/segments
		point := f32.Pt(
			center.X+float32(math.Cos(float64(angle)))*radius,
			center.Y+float32(math.Sin(float64(angle)))*radius,
		)
		if index == 0 {
			path.MoveTo(point)
			continue
		}
		path.LineTo(point)
	}
	stroke := clip.Stroke{Path: path.End(), Width: strokeWidth}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, request.Color)
	stroke.Pop()
	return layout.Dimensions{Size: image.Pt(d, d)}
}

func foregroundColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Text == nil {
		return color.NRGBA{}, false
	}
	return solidColor(resolved.Text.Color)
}

// rippleRenderer paints the provider-declared press feedback. A disabled
// duration produces a static/no-op frame and avoids invalidating a window
// when the active motion policy disables the effect.
func rippleRenderer(gtx layout.Context, request theme.ButtonRippleRequest) {
	hasRipples := len(request.Ripples) > 0
	hasSize := request.Size.X > 0 && request.Size.Y > 0
	hasDurations := request.SpreadDuration > 0 && request.FadeDuration > 0
	if !request.ColorValid || !hasRipples || !hasSize || !hasDurations {
		return
	}
	gtx.Execute(op.InvalidateCmd{})
	for _, ripple := range request.Ripples {
		spread := circularOut(clamp01(float32(gtx.Now.Sub(ripple.StartedAt)) / float32(request.SpreadDuration)))
		fade := circularOut(clamp01(float32(gtx.Now.Sub(ripple.StartedAt)) / float32(request.FadeDuration)))
		if spread <= 0 || fade >= 1 {
			continue
		}
		color := request.Color
		color.A = uint8(float32(color.A)*0.2*(1-fade) + .5)
		width := float32(gtx.Dp(6)) * spread
		inset := int(width/2 + .5)
		rect := image.Rectangle{Max: request.Size}.Inset(-inset)
		radius := gtx.Dp(request.Radius) + inset
		stroke := clip.Stroke{
			Path:  clip.UniformRRect(rect, radius).Path(gtx.Ops),
			Width: width,
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, color)
		stroke.Pop()
	}
}

// rippleColor resolves the target surface used by press feedback. The border
// wins over the background, while transparent and white surfaces do not
// produce a visible wave.
func rippleColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Paint == nil {
		return color.NRGBA{}, false
	}
	if resolved.Paint.Border != nil {
		if border, ok := solidColor(resolved.Paint.Border.Color); ok && validRippleColor(border) {
			return border, true
		}
	}
	if background, ok := solidColor(resolved.Paint.Background); ok && validRippleColor(background) {
		return background, true
	}
	return color.NRGBA{}, false
}

func solidColor(source any) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		return value.Color, true
	default:
		return color.NRGBA{}, false
	}
}

func validRippleColor(value color.NRGBA) bool {
	return value.A != 0 && (value.R != 0xff || value.G != 0xff || value.B != 0xff)
}

// groupSeparatorRenderer paints one provider-owned separator. The widget
// supplies only the item offset, axis, bounds, and resolved foreground.
func groupSeparatorRenderer(
	gtx layout.Context,
	request theme.ButtonGroupSeparatorRequest,
	width unit.Dp,
	length, opacity float32,
) {
	thickness := max(gtx.Dp(width), 1)
	ratio := clamp01(length)
	alpha := clamp01(opacity)
	foreground := request.Foreground
	foreground.A = separatorAlpha(foreground.A, alpha)
	if request.Axis == theme.ButtonGroupVertical {
		rect := groupSeparatorRect(request.Size, request.MainOffset, thickness, request.Axis, ratio)
		paint.FillShape(gtx.Ops, foreground, clip.Rect(rect).Op())
		return
	}
	rect := groupSeparatorRect(request.Size, request.MainOffset, thickness, request.Axis, ratio)
	paint.FillShape(gtx.Ops, foreground, clip.Rect(rect).Op())
}

func groupSeparatorRect(size image.Point, mainOffset, thickness int, axis theme.ButtonGroupAxis, length float32) image.Rectangle {
	if axis == theme.ButtonGroupVertical {
		lineLength := max(int(float32(size.X)*length), 1)
		return image.Rect(
			(size.X-lineLength)/2,
			mainOffset-thickness/2,
			(size.X+lineLength)/2,
			mainOffset-thickness/2+thickness,
		)
	}
	lineLength := max(int(float32(size.Y)*length), 1)
	return image.Rect(
		mainOffset-thickness/2,
		(size.Y-lineLength)/2,
		mainOffset-thickness/2+thickness,
		(size.Y+lineLength)/2,
	)
}

func separatorAlpha(base uint8, opacity float32) uint8 {
	return byte(float32(base)*clamp01(opacity) + 0.5)
}

func clamp01(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func circularOut(progress float32) float32 {
	progress = clamp01(progress) - 1
	return float32(math.Sqrt(float64(1 - progress*progress)))
}
