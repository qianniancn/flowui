package colorpickerstyle

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
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

type wheelImageCache struct {
	size image.Point
	op   paint.ImageOp
}

func wheelRenderer(tokens theme.ColorWheelTheme, focusColor color.NRGBA) theme.ColorWheelRenderer {
	var cache wheelImageCache
	return func(gtx layout.Context, request theme.ColorWheelRenderRequest) {
		drawColorWheel(
			gtx,
			request.Size,
			request.Value,
			&cache,
			max(gtx.Dp(tokens.ThumbSize), 1),
			max(gtx.Dp(tokens.ThumbBorderWidth), 1),
			max(gtx.Dp(tokens.FocusRingWidth), 1),
			request.Focus,
			focusColor,
		)
	}
}

func areaRenderer(tokens theme.ColorAreaTheme, focus color.NRGBA) theme.ColorAreaRenderer {
	return func(gtx layout.Context, request theme.ColorAreaRenderRequest) {
		thumbSize := tokens.ThumbSize
		if request.Dragging {
			thumbSize = tokens.DraggingThumbSize
		}
		drawColorArea(
			gtx,
			request.Size,
			request.Value,
			max(gtx.Dp(tokens.Radius), 0),
			max(gtx.Dp(thumbSize), 1),
			max(gtx.Dp(tokens.ThumbBorderWidth), 1),
			max(gtx.Dp(tokens.FocusRingWidth), 1),
			request.Focus,
			focus,
			request.ShowDots,
			max(gtx.Dp(tokens.DotSize), 1),
			max(gtx.Dp(tokens.DotGap), 1),
		)
	}
}

func sliderRenderer(tokens theme.ColorSliderTheme, focus color.NRGBA) theme.ColorSliderRenderer {
	return func(gtx layout.Context, request theme.ColorSliderRenderRequest) {
		thumbSize := max(gtx.Dp(tokens.ThumbSize), 1)
		thumbBorder := max(gtx.Dp(tokens.ThumbBorderWidth), 1)
		focusWidth := max(gtx.Dp(tokens.FocusRingWidth), 1)
		if request.Channel == theme.ColorSliderAlpha {
			drawAlphaSlider(gtx, request.Size, request.Color, thumbSize, thumbBorder, focusWidth, request.Focus, focus)
			return
		}
		drawHueSlider(gtx, request.Size, request.Value, thumbSize, thumbBorder, focusWidth, request.Focus, focus)
	}
}

func swatchRenderer(tokens theme.ColorSwatchTheme) theme.ColorSwatchRenderer {
	return func(gtx layout.Context, request theme.ColorSwatchRenderRequest) {
		size := min(request.Size.X, request.Size.Y)
		radius := 0
		if request.Shape == theme.ColorSwatchCircle {
			radius = size / 2
		} else {
			radius = min(max(gtx.Dp(tokens.SquareRadius), 1), size/2)
		}
		drawColorSwatch(gtx, request.Size, request.Value, radius)
	}
}

func swatchPickerRenderer(active *theme.Theme, tokens theme.ColorSwatchPickerTheme) theme.ColorSwatchPickerItemRenderer {
	return func(gtx layout.Context, request theme.ColorSwatchPickerItemRequest) {
		drawColorSwatchPickerItem(gtx, active, tokens, request)
	}
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

func drawColorSwatchPickerItem(gtx layout.Context, active *theme.Theme, tokens theme.ColorSwatchPickerTheme, request theme.ColorSwatchPickerItemRequest) {
	rect := image.Rectangle{Max: request.Size}
	minimum := min(request.Size.X, request.Size.Y)
	borderWidth := swatchPickerBorderWidth(gtx, tokens, request.SwatchSize)
	itemRadius := swatchPickerItemRadius(gtx, tokens, minimum, request.SwatchSize, request.Shape)
	if request.Focus > 0 {
		focusColor := active.Semantic().Ring
		focusColor.A = byte(float32(focusColor.A)*request.Focus + .5)
		focusWidth := max(gtx.Dp(tokens.FocusRingWidth), 1)
		focusGap := max(gtx.Dp(tokens.FocusRingGap), 0)
		expand := focusGap + focusWidth
		focusRect := rect.Inset(-expand)
		focusRadius := itemRadius + expand
		if request.Shape == theme.ColorSwatchCircle {
			focusRadius = min(focusRect.Dx(), focusRect.Dy()) / 2
		}
		drawRoundedBorder(gtx, focusRect, focusRadius, focusWidth, focusColor, request.Background)
	}
	if request.Selection > 0 {
		shadowOpacity := tokens.ShadowOpacity * request.Selection
		if shadowOpacity > 0 {
			shadowShape := draw.EllipseShadow()
			if request.Shape == theme.ColorSwatchSquare {
				radius := swatchPickerItemRadiusDp(tokens, request.SwatchSize)
				shadowShape = draw.RoundedShadowCorners(radius, radius, radius, radius)
			}
			sem := active.Semantic()
			draw.Shadow(gtx, rect, shadowShape, draw.ThemeShadow(active.Shadows.Control, sem.SurfaceShadow, shadowOpacity))
		}
		border := request.Value
		border.A = byte(float32(255)*request.Selection + .5)
		drawRoundedBorder(gtx, rect, itemRadius, borderWidth, border, request.Background)
	}
	scale := float32(1) + .1*request.Hover
	if request.Selection > 0 {
		scale = 1 + (tokens.SelectedScale-1)*request.Selection
	}
	swatchPixels := swatchPickerVisualSize(request.Size, borderWidth, scale)
	swatchRect := image.Rectangle{Min: image.Pt((request.Size.X-swatchPixels.X)/2, (request.Size.Y-swatchPixels.Y)/2)}
	swatchRect.Max = swatchRect.Min.Add(swatchPixels)
	offset := op.Offset(swatchRect.Min).Push(gtx.Ops)
	drawColorSwatchFill(gtx, swatchRect.Size(), request.Value, swatchPickerSwatchRadius(gtx, tokens, min(swatchRect.Dx(), swatchRect.Dy()), request.SwatchSize, request.Shape, request.Selection))
	offset.Pop()
	if request.Check > 0 {
		drawColorSwatchCheck(gtx, request.Size, request.Value, request.Check, float32(gtx.Dp(tokens.CheckStroke)))
	}
}

func swatchPickerBorderWidth(gtx layout.Context, tokens theme.ColorSwatchPickerTheme, size uint8) int {
	switch size {
	case 0:
		return max(gtx.Dp(tokens.ExtraSmallBorderWidth), 1)
	case 3, 4:
		return max(gtx.Dp(tokens.LargeBorderWidth), 1)
	default:
		return max(gtx.Dp(tokens.BorderWidth), 1)
	}
}

func swatchPickerItemRadius(gtx layout.Context, tokens theme.ColorSwatchPickerTheme, side int, size uint8, shape theme.ColorSwatchShape) int {
	if shape == theme.ColorSwatchCircle {
		return side / 2
	}
	radius := tokens.SquareItemRadius
	if size == 0 {
		radius = tokens.SquareItemRadiusExtraSmall
	} else if size == 1 {
		radius = tokens.SquareItemRadiusSmall
	}
	return min(max(gtx.Dp(radius), 1), side/2)
}

func swatchPickerItemRadiusDp(tokens theme.ColorSwatchPickerTheme, size uint8) unit.Dp {
	value := tokens.SquareItemRadius
	if size == 0 {
		value = tokens.SquareItemRadiusExtraSmall
	} else if size == 1 {
		value = tokens.SquareItemRadiusSmall
	}
	return value
}

func swatchPickerSwatchRadius(gtx layout.Context, tokens theme.ColorSwatchPickerTheme, side int, size uint8, shape theme.ColorSwatchShape, selection float32) int {
	if shape == theme.ColorSwatchCircle {
		return side / 2
	}
	radius := tokens.SquareSwatchRadius
	switch size {
	case 0:
		radius = tokens.SquareSwatchRadiusExtraSmall
	case 1:
		radius = unit.Dp(float32(tokens.SquareSwatchRadiusSmall) + (float32(tokens.SquareSelectedSmallRadius)-float32(tokens.SquareSwatchRadiusSmall))*selection)
	}
	return min(max(gtx.Dp(radius), 1), side/2)
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
	path := draw.CheckPath(gtx.Ops, points, progress)
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

func drawColorArea(gtx layout.Context, size image.Point, value theme.ColorPickerHSV, radius, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA, showDots bool, dotSize, dotGap int) {
	rect := image.Rectangle{Max: size}
	hue := hsvToNRGBA(theme.ColorPickerHSV{Hue: value.Hue, Saturation: 1, Value: 1, Alpha: 1})
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, hue)
	draw.PaintLinearGradient(gtx, f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)), color.NRGBA{R: 255, G: 255, B: 255, A: 255}, f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)), color.NRGBA{R: 255, G: 255, B: 255})
	draw.PaintLinearGradient(gtx, f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)), color.NRGBA{}, f32.Pt(float32(rect.Min.X), float32(rect.Max.Y)), color.NRGBA{A: 255})
	clipped.Pop()
	if showDots {
		drawColorAreaDots(gtx, rect, radius, dotSize, dotGap)
	}
	drawRoundedStroke(gtx, rect, radius, 1, color.NRGBA{A: 26})
	center := image.Pt(int(float64(max(size.X-1, 0))*value.Saturation+0.5), int(float64(max(size.Y-1, 0))*(1-value.Value)+0.5))
	thumbColor := hsvToNRGBA(theme.ColorPickerHSV{Hue: value.Hue, Saturation: value.Saturation, Value: value.Value, Alpha: 1})
	drawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
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

func drawHueSlider(gtx layout.Context, size image.Point, value theme.ColorPickerHSV, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	rect := image.Rectangle{Max: size}
	drawHueSliderTrack(gtx, rect, min(size.X, size.Y)/2)
	drawRoundedStroke(gtx, rect, min(size.X, size.Y)/2, 1, color.NRGBA{A: 26})
	center := colorSliderCenter(size, value.Hue)
	thumbColor := hsvToNRGBA(theme.ColorPickerHSV{Hue: value.Hue, Saturation: 1, Value: 1, Alpha: 1})
	drawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
}

func drawAlphaSlider(gtx layout.Context, size image.Point, value color.NRGBA, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	drawCheckerboard(gtx, rect, radius)
	transparent := value
	transparent.A = 0
	opaque := value
	opaque.A = 255
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	draw.PaintLinearGradient(gtx, f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)), transparent, f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)), opaque)
	clipped.Pop()
	drawRoundedStroke(gtx, rect, radius, 1, color.NRGBA{A: 26})
	center := colorSliderCenter(size, float64(value.A)/255)
	drawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, value, focus, focusColor)
}

func colorSliderCenter(size image.Point, value float64) image.Point {
	edge := float64(min(size.X, size.Y)) / 2
	length := max(float64(size.X)-edge*2, 0)
	return image.Pt(int(edge+clampUnit(value)*length+0.5), size.Y/2)
}

func drawColorThumb(gtx layout.Context, center image.Point, size, border, focusWidth int, value color.NRGBA, focus float32, focusColor color.NRGBA) {
	if size <= 0 {
		return
	}
	if focus > 0 {
		col := focusColor
		col.A = byte(float32(col.A)*focus + .5)
		drawCircle(gtx, center, size+focusWidth*4, col)
		drawCircle(gtx, center, size+focusWidth*2, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
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

func drawRoundedStroke(gtx layout.Context, rect image.Rectangle, radius, width int, value color.NRGBA) {
	if rect.Empty() || width <= 0 || value.A == 0 {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, value)
	stroke.Pop()
}

func drawCheckerboard(gtx layout.Context, rect image.Rectangle, radius int) {
	if rect.Empty() {
		return
	}
	clipped := clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Push(gtx.Ops)
	defer clipped.Pop()
	cell := max(min(rect.Dx(), rect.Dy())/4, 2)
	colors := [2]color.NRGBA{{R: 0xef, G: 0xef, B: 0xef, A: 255}, {R: 0xf7, G: 0xf7, B: 0xf7, A: 255}}
	for y, row := rect.Min.Y, 0; y < rect.Max.Y; y, row = y+cell, row+1 {
		for x, column := rect.Min.X, 0; x < rect.Max.X; x, column = x+cell, column+1 {
			tile := image.Rect(x, y, min(x+cell, rect.Max.X), min(y+cell, rect.Max.Y))
			paint.FillShape(gtx.Ops, colors[(row+column)%2], clip.Rect(tile).Op())
		}
	}
}

func drawHueSliderTrack(gtx layout.Context, rect image.Rectangle, radius int) {
	if rect.Empty() {
		return
	}
	imageOp := hueSliderImage()
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	scale := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(float32(rect.Dx())/float32(imageOp.Size().X), float32(rect.Dy())/float32(imageOp.Size().Y)))).Push(gtx.Ops)
	imageOp.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	scale.Pop()
	clipped.Pop()
}

var hueSliderImageOp = newHueSliderImage()

func hueSliderImage() paint.ImageOp {
	return hueSliderImageOp
}

func newHueSliderImage() paint.ImageOp {
	const width = 360
	value := image.NewNRGBA(image.Rect(0, 0, width, 1))
	for x := range width {
		value.SetNRGBA(x, 0, hsvToNRGBA(theme.ColorPickerHSV{Hue: float64(x) / (width - 1), Saturation: 1, Value: 1, Alpha: 1}))
	}
	return paint.NewImageOp(value)
}

func drawColorWheel(gtx layout.Context, size image.Point, value theme.ColorPickerHSV, cache *wheelImageCache, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	if cache.size != size || cache.op.Size() != size {
		cache.size = size
		cache.op = paint.NewImageOp(buildColorWheelImage(size))
	}
	clipped := clip.Ellipse(image.Rectangle{Max: size}).Push(gtx.Ops)
	cache.op.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	clipped.Pop()
	drawRoundedStroke(gtx, image.Rectangle{Max: size}, min(size.X, size.Y)/2, 1, color.NRGBA{A: 26})
	center := colorWheelSelectorCenter(size, value, (thumbSize+focusWidth*4+1)/2)
	thumbColor := hsvToNRGBA(theme.ColorPickerHSV{Hue: value.Hue, Saturation: value.Saturation, Value: value.Value, Alpha: 1})
	drawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
}

func colorWheelSelectorCenter(size image.Point, value theme.ColorPickerHSV, inset int) image.Point {
	radius := max(float64(min(size.X, size.Y))/2-float64(max(inset, 0)), 0)
	angle := wrapUnit(value.Hue) * 2 * math.Pi
	centerX := float64(size.X) / 2
	centerY := float64(size.Y) / 2
	distance := radius * clampUnit(value.Saturation)
	return image.Pt(int(math.Round(centerX+math.Cos(angle)*distance)), int(math.Round(centerY+math.Sin(angle)*distance)))
}

func buildColorWheelImage(size image.Point) *image.NRGBA {
	value := image.NewNRGBA(image.Rectangle{Max: size})
	centerX := float64(size.X-1) / 2
	centerY := float64(size.Y-1) / 2
	radius := float64(min(size.X, size.Y)) / 2
	if radius <= 0 {
		return value
	}
	for y := range size.Y {
		for x := range size.X {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distance := math.Hypot(dx, dy)
			coverage := min(max(radius+.5-distance, 0), 1)
			if coverage == 0 {
				continue
			}
			pixel := hsvToNRGBA(theme.ColorPickerHSV{Hue: wrapUnit(math.Atan2(dy, dx) / (2 * math.Pi)), Saturation: clampUnit(distance / radius), Value: 1, Alpha: coverage})
			value.SetNRGBA(x, y, pixel)
		}
	}
	return value
}

func hsvToNRGBA(value theme.ColorPickerHSV) color.NRGBA {
	hue := value.Hue - math.Floor(value.Hue)
	saturation := clampUnit(value.Saturation)
	brightness := clampUnit(value.Value)
	chroma := brightness * saturation
	section := hue * 6
	x := chroma * (1 - math.Abs(math.Mod(section, 2)-1))
	var r, g, b float64
	switch int(section) % 6 {
	case 0:
		r, g = chroma, x
	case 1:
		r, g = x, chroma
	case 2:
		g, b = chroma, x
	case 3:
		g, b = x, chroma
	case 4:
		r, b = x, chroma
	default:
		r, b = chroma, x
	}
	match := brightness - chroma
	channel := func(value float64) uint8 { return uint8(clampUnit(value+match)*255 + .5) }
	return color.NRGBA{R: channel(r), G: channel(g), B: channel(b), A: uint8(clampUnit(value.Alpha)*255 + .5)}
}

func clampUnit(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return min(max(value, 0), 1)
}

func wrapUnit(value float64) float64 {
	value -= math.Floor(value)
	if value < 0 {
		value++
	}
	return value
}

func swatchPickerVisualSize(item image.Point, borderWidth int, scale float32) image.Point {
	content := image.Rectangle{Max: item}.Inset(max(borderWidth, 0)).Size()
	if content.X <= 0 || content.Y <= 0 {
		return image.Point{}
	}
	return image.Pt(max(int(float32(content.X)*scale+.5), 1), max(int(float32(content.Y)*scale+.5), 1))
}
