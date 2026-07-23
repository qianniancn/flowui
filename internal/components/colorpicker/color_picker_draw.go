package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

var hueSliderImage = newHueSliderImage()

func drawColorPickerPanel(gtx layout.Context, activeTheme *theme.Theme, size image.Point, radius int) {
	render.DrawSurface(
		gtx,
		image.Rectangle{Max: size},
		radius,
		activeTheme.Palette.OverlayColor(),
		render.ThemeShadow(activeTheme.Shadows.Overlay, activeTheme.Palette.OverlayShadowColor(), 1),
	)
}

func drawColorSwatch(gtx layout.Context, size image.Point, value color.NRGBA, radius int) {
	drawColorSwatchFill(gtx, size, value, radius)
	drawRoundedStroke(gtx, image.Rectangle{Max: size}, radius, 1, color.NRGBA{A: 26})
}

func drawColorSwatchFill(gtx layout.Context, size image.Point, value color.NRGBA, radius int) {
	rect := image.Rectangle{Max: size}
	if value.A < 255 {
		drawCheckerboard(gtx, rect, radius)
	}
	paint.FillShape(gtx.Ops, value, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawColorSwatchPickerItem(ctx *frame.Context, gtx layout.Context, size image.Point, value color.NRGBA, swatchSize ColorSwatchSize, shape ColorSwatchShape, selection, check, hover, focus float32) {
	tokens := frame.ActiveTheme(ctx).Components.ColorSwatchPicker
	rect := image.Rectangle{Max: size}
	minimum := min(size.X, size.Y)
	borderWidth := colorSwatchPickerBorderWidth(ctx, gtx, swatchSize)
	itemRadius := colorSwatchPickerItemRadius(ctx, gtx, minimum, swatchSize, shape)
	if focus > 0 {
		focusColor := frame.ActiveTheme(ctx).Palette.Focus
		focusColor.A = byte(float32(focusColor.A)*focus + .5)
		focusWidth := max(gtx.Dp(tokens.FocusRingWidth), 1)
		focusGap := max(gtx.Dp(tokens.FocusRingGap), 0)
		expand := focusGap + focusWidth
		focusRect := rect.Inset(-expand)
		focusRadius := itemRadius + expand
		if shape == ColorSwatchCircle {
			focusRadius = min(focusRect.Dx(), focusRect.Dy()) / 2
		}
		drawRoundedBorder(gtx, focusRect, focusRadius, focusWidth, focusColor, ctx.BackgroundColor())
	}
	if selection > 0 {
		shadowOpacity := tokens.ShadowOpacity * selection
		if shadowOpacity > 0 {
			shadowShape := render.EllipseShadow()
			if shape == ColorSwatchSquare {
				radius := colorSwatchPickerItemRadiusDp(ctx, swatchSize)
				shadowShape = render.RoundedShadowCorners(radius, radius, radius, radius)
			}
			activeTheme := frame.ActiveTheme(ctx)
			render.DrawShadow(gtx, rect, shadowShape, render.ThemeShadow(activeTheme.Shadows.Control, activeTheme.Palette.SurfaceShadow, shadowOpacity))
		}
		border := value
		border.A = byte(float32(255)*selection + .5)
		drawRoundedBorder(gtx, rect, itemRadius, borderWidth, border, ctx.BackgroundColor())
	}
	scale := float32(1) + .1*hover
	if selection > 0 {
		scale = 1 + (tokens.SelectedScale-1)*selection
	}
	swatchPixels := colorSwatchPickerVisualSize(size, borderWidth, scale)
	swatchRect := image.Rectangle{Min: image.Pt((size.X-swatchPixels.X)/2, (size.Y-swatchPixels.Y)/2)}
	swatchRect.Max = swatchRect.Min.Add(swatchPixels)
	offset := op.Offset(swatchRect.Min).Push(gtx.Ops)
	drawColorSwatchFill(gtx, swatchRect.Size(), value, colorSwatchPickerSwatchRadius(ctx, gtx, min(swatchRect.Dx(), swatchRect.Dy()), swatchSize, shape, selection))
	offset.Pop()
	if check > 0 {
		drawColorSwatchCheck(gtx, size, value, check, float32(gtx.Dp(tokens.CheckStroke)))
	}
}

func drawRoundedBorder(gtx layout.Context, rect image.Rectangle, radius, width int, border, background color.NRGBA) {
	if rect.Empty() || width <= 0 || border.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, border, clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Op(gtx.Ops))
	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(inner, min(max(radius-width, 0), min(inner.Dx(), inner.Dy())/2)).Op(gtx.Ops))
}

func colorSwatchPickerVisualSize(item image.Point, borderWidth int, scale float32) image.Point {
	content := image.Rectangle{Max: item}.Inset(max(borderWidth, 0)).Size()
	if content.X <= 0 || content.Y <= 0 {
		return image.Point{}
	}
	return image.Pt(
		max(int(float32(content.X)*scale+.5), 1),
		max(int(float32(content.Y)*scale+.5), 1),
	)
}

func drawColorSwatchCheck(gtx layout.Context, size image.Point, value color.NRGBA, progress, strokeWidth float32) {
	checkSize := float32(min(size.X, size.Y)) / 3
	x := (float32(size.X) - checkSize) / 2
	y := (float32(size.Y) - checkSize) / 2
	points := [3]f32.Point{
		f32.Pt(x+checkSize*.05, y+checkSize*.56),
		f32.Pt(x+checkSize*.40, y+checkSize*.86),
		f32.Pt(x+checkSize*.95, y+checkSize*.14),
	}
	path := render.CheckPath(gtx.Ops, points, progress)
	foreground := color.NRGBA{R: 255, G: 255, B: 255, A: byte(255*progress + .5)}
	if colorLuminance(value) > .5 {
		foreground.R, foreground.G, foreground.B = 0, 0, 0
	}
	stroke := clip.Stroke{Path: path, Width: max(strokeWidth, 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
}

func colorLuminance(value color.NRGBA) float64 {
	return (.2126*float64(value.R) + .7152*float64(value.G) + .0722*float64(value.B)) / 255
}

func drawColorArea(gtx layout.Context, size image.Point, value hsvColor, radius, thumbSize, thumbBorder int, focus float32, focusColor color.NRGBA, showDots bool, dotSize, dotGap int) {
	rect := image.Rectangle{Max: size}
	hue := hsvToNRGBA(hsvColor{h: value.h, s: 1, v: 1, a: 1})
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, hue)
	render.PaintLinearGradient(
		gtx,
		f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)),
		color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)),
		color.NRGBA{R: 255, G: 255, B: 255},
	)
	render.PaintLinearGradient(
		gtx,
		f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)),
		color.NRGBA{},
		f32.Pt(float32(rect.Min.X), float32(rect.Max.Y)),
		color.NRGBA{A: 255},
	)
	clipped.Pop()
	if showDots {
		drawColorAreaDots(gtx, rect, radius, dotSize, dotGap)
	}
	drawRoundedStroke(gtx, rect, radius, 1, color.NRGBA{A: 26})

	center := image.Pt(
		int(float64(max(size.X-1, 0))*value.s+0.5),
		int(float64(max(size.Y-1, 0))*(1-value.v)+0.5),
	)
	thumbColor := hsvToNRGBA(hsvColor{h: value.h, s: value.s, v: value.v, a: 1})
	drawColorThumb(gtx, center, thumbSize, thumbBorder, thumbColor, focus, focusColor)
}

func drawColorAreaDots(gtx layout.Context, rect image.Rectangle, radius, size, gap int) {
	if rect.Empty() || size <= 0 || gap <= 0 {
		return
	}
	clipped := clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Push(gtx.Ops)
	defer clipped.Pop()
	value := color.NRGBA{R: 255, G: 255, B: 255, A: 51}
	for y := rect.Min.Y + gap/2; y < rect.Max.Y; y += gap {
		for x := rect.Min.X + gap/2; x < rect.Max.X; x += gap {
			drawCircle(gtx, image.Pt(x, y), size, value)
		}
	}
}

func drawHueSlider(gtx layout.Context, size image.Point, value hsvColor, thumbSize, thumbBorder int, focus float32, focusColor color.NRGBA) {
	rect := image.Rectangle{Max: size}
	drawHueSliderTrack(gtx, rect, min(size.X, size.Y)/2)
	drawRoundedStroke(gtx, rect, min(size.X, size.Y)/2, 1, color.NRGBA{A: 26})
	center := colorSliderCenter(size, value.h)
	thumbColor := hsvToNRGBA(hsvColor{h: value.h, s: 1, v: 1, a: 1})
	drawColorThumb(gtx, center, thumbSize, thumbBorder, thumbColor, focus, focusColor)
}

func newHueSliderImage() paint.ImageOp {
	const width = 360
	value := image.NewNRGBA(image.Rect(0, 0, width, 1))
	for x := range width {
		value.SetNRGBA(x, 0, hsvToNRGBA(hsvColor{h: float64(x) / (width - 1), s: 1, v: 1, a: 1}))
	}
	return paint.NewImageOp(value)
}

func drawHueSliderTrack(gtx layout.Context, rect image.Rectangle, radius int) {
	if rect.Empty() || hueSliderImage.Size().X <= 0 {
		return
	}
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	scale := op.Affine(f32.AffineId().Scale(
		f32.Point{},
		f32.Pt(float32(rect.Dx())/float32(hueSliderImage.Size().X), float32(rect.Dy())/float32(hueSliderImage.Size().Y)),
	)).Push(gtx.Ops)
	hueSliderImage.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	scale.Pop()
	clipped.Pop()
}

func drawAlphaSlider(gtx layout.Context, size image.Point, value color.NRGBA, thumbSize, thumbBorder int, focus float32, focusColor color.NRGBA) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	drawCheckerboard(gtx, rect, radius)
	transparent := value
	transparent.A = 0
	opaque := value
	opaque.A = 255
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	render.PaintLinearGradient(
		gtx,
		f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)),
		transparent,
		f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)),
		opaque,
	)
	clipped.Pop()
	drawRoundedStroke(gtx, rect, radius, 1, color.NRGBA{A: 26})
	center := colorSliderCenter(size, float64(value.A)/255)
	drawColorThumb(gtx, center, thumbSize, thumbBorder, value, focus, focusColor)
}

func colorSliderCenter(size image.Point, value float64) image.Point {
	edge := float64(min(size.X, size.Y)) / 2
	length := max(float64(size.X)-edge*2, 0)
	return image.Pt(int(edge+clampUnit(value)*length+0.5), size.Y/2)
}

func drawColorThumb(gtx layout.Context, center image.Point, size, border int, value color.NRGBA, focus float32, focusColor color.NRGBA) {
	if size <= 0 {
		return
	}
	if focus > 0 {
		col := focusColor
		col.A = byte(float32(col.A)*focus + .5)
		drawCircle(gtx, center, size+8, col)
		drawCircle(gtx, center, size+4, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	}
	drawCircle(gtx, center, size+2, color.NRGBA{A: 38})
	drawCircle(gtx, center, size, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	drawCircle(gtx, center, max(size-border*2, 1), value)
}

func drawCircle(gtx layout.Context, center image.Point, diameter int, value color.NRGBA) {
	if diameter <= 0 || value.A == 0 {
		return
	}
	half := diameter / 2
	rect := image.Rect(center.X-half, center.Y-half, center.X-half+diameter, center.Y-half+diameter)
	paint.FillShape(gtx.Ops, value, clip.Ellipse(rect).Op(gtx.Ops))
}

func drawTriggerFocus(gtx layout.Context, size image.Point, radius, width int, opacity float32, value color.NRGBA) {
	if opacity <= 0 || width <= 0 {
		return
	}
	value.A = byte(float32(value.A)*opacity + .5)
	rect := image.Rectangle{Max: size}.Inset(width)
	if rect.Empty() {
		return
	}
	drawRoundedStroke(gtx, rect, max(radius-width, 0), width, value)
}

func drawCheckerboard(gtx layout.Context, rect image.Rectangle, radius int) {
	if rect.Empty() {
		return
	}
	clipped := clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Push(gtx.Ops)
	defer clipped.Pop()
	cell := max(min(rect.Dx(), rect.Dy())/4, 2)
	colors := [2]color.NRGBA{
		{R: 0xef, G: 0xef, B: 0xef, A: 255},
		{R: 0xf7, G: 0xf7, B: 0xf7, A: 255},
	}
	for y, row := rect.Min.Y, 0; y < rect.Max.Y; y, row = y+cell, row+1 {
		for x, column := rect.Min.X, 0; x < rect.Max.X; x, column = x+cell, column+1 {
			tile := image.Rect(x, y, min(x+cell, rect.Max.X), min(y+cell, rect.Max.Y))
			paint.FillShape(gtx.Ops, colors[(row+column)%2], clip.Rect(tile).Op())
		}
	}
}

func drawRoundedStroke(gtx layout.Context, rect image.Rectangle, radius, width int, value color.NRGBA) {
	if rect.Empty() || width <= 0 || value.A == 0 {
		return
	}
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, value)
	stroke.Pop()
}
