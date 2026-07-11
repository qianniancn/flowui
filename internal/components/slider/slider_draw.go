package slider

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type sliderGeometry struct {
	size       image.Point
	axis       layout.Axis
	edge       int
	inner      int
	centers    [2]image.Point
	thumbRects [2]image.Rectangle
}

func newSliderGeometry(size image.Point, axis layout.Axis, edge int, lower, upper float32, rangeMode bool, thumbLength, thumbExtra int) sliderGeometry {
	main := axis.Convert(size).X
	edge = min(max(edge, 0), max(main/2, 0))
	inner := max(main-2*edge, 0)
	outerLength := max(thumbLength+thumbExtra, 1)
	outerCross := max(sizeCross(axis, size), 1)
	geometry := sliderGeometry{size: size, axis: axis, edge: edge, inner: inner}
	ratios := [2]float32{min(max(lower, 0), 1), min(max(upper, 0), 1)}
	if !rangeMode {
		ratios[1] = ratios[0]
	}
	for index, ratio := range ratios {
		mainCenter := edge + int(float32(inner)*ratio+0.5)
		if axis == layout.Vertical {
			mainCenter = main - mainCenter
		}
		center := axis.Convert(image.Pt(mainCenter, outerCross/2))
		geometry.centers[index] = center
		outerSize := axis.Convert(image.Pt(outerLength, outerCross))
		geometry.thumbRects[index] = image.Rectangle{
			Min: center.Sub(outerSize.Div(2)),
			Max: center.Sub(outerSize.Div(2)).Add(outerSize),
		}
	}
	return geometry
}

func sizeCross(axis layout.Axis, size image.Point) int {
	return axis.Convert(size).Y
}

func drawSliderTrack(gtx layout.Context, activeTheme *theme.Theme, style sliderStyle, geometry sliderGeometry, rangeMode bool) {
	rect := image.Rectangle{Max: geometry.size}
	radius := min(max(gtx.Dp(activeTheme.Components.Slider.TrackRadius), 0), min(rect.Dx(), rect.Dy())/2)
	track := clip.UniformRRect(rect, radius)
	paint.FillShape(gtx.Ops, style.track, track.Op(gtx.Ops))

	fill := sliderFillRect(geometry, rangeMode)
	if fill.Empty() {
		return
	}
	stack := track.Push(gtx.Ops)
	paint.FillShape(gtx.Ops, style.fill, clip.Rect(fill).Op())
	stack.Pop()
}

func sliderFillRect(geometry sliderGeometry, rangeMode bool) image.Rectangle {
	if geometry.axis == layout.Vertical {
		start := geometry.centers[0].Y
		end := geometry.size.Y
		if rangeMode {
			start = geometry.centers[1].Y
			end = geometry.centers[0].Y
		}
		return image.Rect(0, min(start, end), geometry.size.X, max(start, end))
	}
	start := 0
	end := geometry.centers[0].X
	if rangeMode {
		start = geometry.centers[0].X
		end = geometry.centers[1].X
	}
	return image.Rect(min(start, end), 0, max(start, end), geometry.size.Y)
}

func drawSliderThumb(gtx layout.Context, activeTheme *theme.Theme, style sliderStyle, rect image.Rectangle, axis layout.Axis, focus, scale float32) {
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(activeTheme.Components.Slider.TrackRadius), 0), min(rect.Dx(), rect.Dy())/2)
	if focus > 0 {
		drawSliderFocus(gtx, activeTheme, rect, radius, style.focus, focus)
	}
	paint.FillShape(gtx.Ops, style.thumb, clip.UniformRRect(rect, radius).Op(gtx.Ops))

	innerSize := axis.Convert(image.Pt(gtx.Dp(activeTheme.Components.Slider.ThumbLength), gtx.Dp(activeTheme.Components.Slider.ThumbCross)))
	innerSize.X = max(int(float32(innerSize.X)*scale+0.5), 1)
	innerSize.Y = max(int(float32(innerSize.Y)*scale+0.5), 1)
	center := rect.Min.Add(rect.Size().Div(2))
	inner := image.Rectangle{Min: center.Sub(innerSize.Div(2))}
	inner.Max = inner.Min.Add(innerSize)
	drawSliderThumbShadow(gtx, inner, gtx.Dp(activeTheme.Components.Slider.ThumbRadius), style.thumbInner.A)
	innerRadius := min(max(gtx.Dp(activeTheme.Components.Slider.ThumbRadius), 0), min(inner.Dx(), inner.Dy())/2)
	paint.FillShape(gtx.Ops, style.thumbInner, clip.UniformRRect(inner, innerRadius).Op(gtx.Ops))
}

func drawSliderFocus(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, col color.NRGBA, opacity float32) {
	width := max(gtx.Dp(activeTheme.Components.Slider.FocusRingWidth), 1)
	offset := max(gtx.Dp(activeTheme.Components.Slider.FocusRingOffset), 0)
	ringRect := rect.Inset(-offset - width/2)
	ring := col
	ring.A = byte(float32(ring.A)*opacity + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(ringRect, radius+offset).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, ring)
	stroke.Pop()
}

func drawSliderThumbShadow(gtx layout.Context, rect image.Rectangle, radius int, opacity byte) {
	for _, layer := range []struct {
		offset image.Point
		spread int
		alpha  byte
	}{
		{offset: image.Pt(0, 2), spread: 2, alpha: 0x0a},
		{offset: image.Pt(0, 1), spread: 1, alpha: 0x0f},
	} {
		shadow := rect.Add(layer.offset).Inset(-layer.spread)
		alpha := byte(uint16(layer.alpha) * uint16(opacity) / 0xff)
		paint.FillShape(gtx.Ops, color.NRGBA{A: alpha}, clip.UniformRRect(shadow, radius+layer.spread).Op(gtx.Ops))
	}
}
