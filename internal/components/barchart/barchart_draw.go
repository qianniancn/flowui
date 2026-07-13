package barchart

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawChartGrid(gtx layout.Context, geometry chartGeometry, style chartStyle, show bool, tokens theme.BarChartTheme) {
	if !show || geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.GridWidth), 1))
	for _, tick := range geometry.yTicks {
		if geometry.horizontal {
			drawChartLine(gtx, f32.Pt(tick.pixel, float32(geometry.plot.Min.Y)), f32.Pt(tick.pixel, float32(geometry.plot.Max.Y)), width, style.grid)
		} else {
			drawChartLine(gtx, f32.Pt(float32(geometry.plot.Min.X), tick.pixel), f32.Pt(float32(geometry.plot.Max.X), tick.pixel), width, style.grid)
		}
	}
}

func drawChartAxes(gtx layout.Context, geometry chartGeometry, style chartStyle, tokens theme.BarChartTheme) {
	if geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.AxisWidth), 1))
	drawChartLine(gtx,
		f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Min.Y)),
		f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)),
		width, style.axis,
	)
	drawChartLine(gtx,
		f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)),
		f32.Pt(float32(geometry.plot.Max.X), float32(geometry.plot.Max.Y)),
		width, style.axis,
	)
}

func drawCategoryHighlight(gtx layout.Context, geometry chartGeometry, selection chartSelection, style chartStyle) {
	if geometry.plot.Empty() || geometry.bandWidth <= 0 {
		return
	}
	var rect image.Rectangle
	if geometry.horizontal {
		top := int(math.Floor(float64(selection.pixelY - geometry.bandWidth/2)))
		bottom := int(math.Ceil(float64(selection.pixelY + geometry.bandWidth/2)))
		rect = image.Rect(geometry.plot.Min.X, max(top, geometry.plot.Min.Y), geometry.plot.Max.X, min(bottom, geometry.plot.Max.Y))
	} else {
		left := int(math.Floor(float64(selection.pixelX - geometry.bandWidth/2)))
		right := int(math.Ceil(float64(selection.pixelX + geometry.bandWidth/2)))
		rect = image.Rect(max(left, geometry.plot.Min.X), geometry.plot.Min.Y, min(right, geometry.plot.Max.X), geometry.plot.Max.Y)
	}
	if !rect.Empty() {
		paint.FillShape(gtx.Ops, style.categoryHover, clip.Rect(rect).Op())
	}
}

func drawChartSeries(ctx *frame.Context, gtx layout.Context, data chartData, geometry chartGeometry, style chartStyle, tokens theme.BarChartTheme) {
	if geometry.plot.Empty() || data.categories == 0 {
		return
	}
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	drawBarBackgrounds(gtx, data, geometry, style, tokens)
	for _, series := range data.series {
		column, ok := geometry.columnLayouts[series.columnID]
		if !ok {
			continue
		}
		start := min(max(geometry.categoryStart, 0), len(series.values))
		end := min(max(geometry.categoryEnd, start), len(series.values))
		for index := start; index < end; index++ {
			bar := series.values[index]
			if !bar.valid {
				continue
			}
			rect := barRectangle(geometry, column, index, bar, max(series.minHeight, gtx.Dp(tokens.MinBarHeight)))
			if rect.Empty() {
				continue
			}
			radius := series.radius
			if !series.hasRadius {
				radius = gtx.Dp(tokens.BarRadius)
			}
			radius = min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)
			paint.FillShape(gtx.Ops, bar.color, clip.UniformRRect(rect, radius).Op(gtx.Ops))
			if series.showLabels {
				drawBarLabel(ctx, gtx, geometry, rect, series, bar, style, tokens)
			}
		}
	}
	area.Pop()
}

func drawBarLabel(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, rect image.Rectangle, series resolvedSeries, bar resolvedBar, style chartStyle, tokens theme.BarChartTheme) {
	text := formatAxisNumber(bar.value, 1)
	if series.formatLabel != nil {
		text = series.formatLabel(bar.value)
	}
	if text == "" {
		return
	}
	position := series.labelPosition
	labelColor := style.axisLabel
	label := recordChartText(ctx, gtx, text, tokens.AxisTextSize, font.Medium, labelColor, max(geometry.bandWidthToInt(), 1))
	if position == LabelAuto {
		position = LabelOutside
		if (geometry.horizontal && rect.Dx() >= label.dims.Size.X+8) || (!geometry.horizontal && rect.Dy() >= label.dims.Size.Y+8) {
			position = LabelInside
		}
	}
	if position == LabelInside {
		labelColor = barLabelContrast(bar.color)
		label = recordChartText(ctx, gtx, text, tokens.AxisTextSize, font.Medium, labelColor, max(rect.Dx()-4, 1))
	}
	x := rect.Min.X + (rect.Dx()-label.dims.Size.X)/2
	y := rect.Min.Y + (rect.Dy()-label.dims.Size.Y)/2
	if position == LabelOutside && geometry.horizontal {
		if bar.value >= 0 {
			x = rect.Max.X + 3
		} else {
			x = rect.Min.X - label.dims.Size.X - 3
		}
	} else if position == LabelOutside {
		if bar.value >= 0 {
			y = rect.Min.Y - label.dims.Size.Y - 3
		} else {
			y = rect.Max.Y + 3
		}
	}
	x = min(max(x, geometry.plot.Min.X), max(geometry.plot.Max.X-label.dims.Size.X, geometry.plot.Min.X))
	y = min(max(y, geometry.plot.Min.Y), max(geometry.plot.Max.Y-label.dims.Size.Y, geometry.plot.Min.Y))
	placeChartText(gtx, label, image.Pt(x, y))
}

func (g chartGeometry) bandWidthToInt() int {
	return int(math.Ceil(float64(max(g.bandWidth, 1))))
}

func barLabelContrast(background color.NRGBA) color.NRGBA {
	luminance := (299*int(background.R) + 587*int(background.G) + 114*int(background.B)) / 1000
	if luminance > 150 {
		return color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}
	}
	return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
}

func drawBarBackgrounds(gtx layout.Context, data chartData, geometry chartGeometry, style chartStyle, tokens theme.BarChartTheme) {
	radius := max(gtx.Dp(tokens.BackgroundRadius), 0)
	for _, column := range data.columns {
		if !column.showBackground {
			continue
		}
		columnLayout, ok := geometry.columnLayouts[column.id]
		if !ok {
			continue
		}
		for index := geometry.categoryStart; index < geometry.categoryEnd; index++ {
			first, second := barCategoryBounds(geometry, columnLayout, index)
			rect := image.Rect(first, geometry.plot.Min.Y, second, geometry.plot.Max.Y)
			if geometry.horizontal {
				rect = image.Rect(geometry.plot.Min.X, first, geometry.plot.Max.X, second)
			}
			if rect.Empty() {
				continue
			}
			barRadius := min(radius, min(rect.Dx(), rect.Dy())/2)
			paint.FillShape(gtx.Ops, style.barBackground, clip.UniformRRect(rect, barRadius).Op(gtx.Ops))
		}
	}
}

func barRectangle(geometry chartGeometry, column columnLayout, category int, bar resolvedBar, minHeight int) image.Rectangle {
	first, second := barCategoryBounds(geometry, column, category)
	start := geometry.mapY(bar.start)
	end := geometry.mapY(bar.end)
	if geometry.horizontal {
		if bar.value != 0 && minHeight > 0 && math.Abs(float64(end-start)) < float64(minHeight) {
			if end < start {
				end = start - float32(minHeight)
			} else {
				end = start + float32(minHeight)
			}
		}
		left := int(math.Round(float64(min(start, end))))
		right := int(math.Round(float64(max(start, end))))
		if bar.value != 0 && right <= left {
			right = left + 1
		}
		return image.Rect(left, first, right, second)
	}
	if bar.value != 0 && minHeight > 0 && math.Abs(float64(end-start)) < float64(minHeight) {
		if end < start {
			end = start - float32(minHeight)
		} else {
			end = start + float32(minHeight)
		}
	}
	top := int(math.Round(float64(min(start, end))))
	bottom := int(math.Round(float64(max(start, end))))
	if bar.value != 0 && bottom <= top {
		bottom = top + 1
	}
	return image.Rect(first, top, second, bottom)
}

func barCategoryBounds(geometry chartGeometry, column columnLayout, category int) (int, int) {
	center := geometry.categoryCenter(category)
	left := int(math.Round(float64(center + column.offset)))
	right := int(math.Round(float64(center + column.offset + column.width)))
	if right <= left {
		right = left + 1
	}
	return left, right
}

func drawChartFocus(gtx layout.Context, size image.Point, col color.NRGBA, opacity float32, tokens theme.BarChartTheme) {
	if opacity <= 0 || size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	width := max(gtx.Dp(tokens.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(tokens.FocusRadius), 0), min(rect.Dx(), rect.Dy())/2)
	col.A = byte(float32(col.A)*opacity + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawChartLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	if width <= 0 || col.A == 0 {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}
