// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/ui"
)

type shadowPreset struct {
	Title string
	Color color.NRGBA
}

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("阴影示例"), app.Size(620, 430))
		if err := run(w); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	var ops op.Ops
	var showPopup bool
	var wheelButton widget.Clickable
	var wheel ColorWheel
	popupControls := PopupControls{
		Brightness: widget.Float{Value: 1},
		Alpha:      widget.Float{Value: .6},
	}
	preset := shadowPreset{Title: "阴影预览", Color: wheel.ShadowColor(popupControls.BrightnessValue())}
	var closeButton widget.Clickable
	popupShape := ui.RoundedShadowCorners(unit.Dp(20), unit.Dp(10), unit.Dp(24), unit.Dp(14))
	th := material.NewTheme()

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			if wheelButton.Clicked(gtx) {
				preset = shadowPreset{
					Title: "阴影预览",
				}
				showPopup = true
			}
			if closeButton.Clicked(gtx) {
				showPopup = false
			}
			layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return page(gtx, th, &wheelButton, &wheel, &popupControls)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					if !showPopup {
						return layout.Dimensions{Size: gtx.Constraints.Max}
					}
					preset.Color = wheel.ShadowColor(popupControls.BrightnessValue())
					preset.Color.A = popupControls.AlphaByte()
					return popupOverlay(gtx, th, &closeButton, popupShape, preset)
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}

func page(gtx layout.Context, th *material.Theme, wheelButton *widget.Clickable, wheel *ColorWheel, popupControls *PopupControls) layout.Dimensions {
	size := gtx.Constraints.Max
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, clip.Rect{Max: size}.Op())
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
			Gap:       gtx.Dp(unit.Dp(16)),
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H6(th, "弹出层阴影")
				title.Alignment = text.Middle
				title.Color = color.NRGBA{R: 0x16, G: 0x24, B: 0x2f, A: 0xff}
				return title.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return controls(gtx, th, wheelButton, wheel, popupControls)
			}),
		)
	})
	return layout.Dimensions{Size: size}
}

func controls(gtx layout.Context, th *material.Theme, wheelButton *widget.Clickable, wheel *ColorWheel, popupControls *PopupControls) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(unit.Dp(10)),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wheel.Layout(gtx, unit.Dp(190))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return popupControls.Layout(gtx, th)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, wheelButton, "查看阴影")
			btn.Background = wheel.ShadowColor(popupControls.BrightnessValue())
			btn.Color = readableTextColor(btn.Background)
			return btn.Layout(gtx)
		}),
	)
}

func popupOverlay(gtx layout.Context, th *material.Theme, closeButton *widget.Clickable, popupShape ui.ShadowShape, preset shadowPreset) layout.Dimensions {
	size := gtx.Constraints.Max

	popupSize := responsivePopupSize(gtx, size)
	popupPos := responsivePopupPosition(gtx, size, popupSize)
	stack := op.Offset(popupPos).Push(gtx.Ops)
	gtx.Constraints = layout.Exact(popupSize)
	shadowPopup(gtx, th, closeButton, popupShape, preset, popupSize)
	stack.Pop()

	return layout.Dimensions{Size: size}
}

func responsivePopupSize(gtx layout.Context, viewport image.Point) image.Point {
	margin := gtx.Dp(unit.Dp(32))
	available := viewport.Sub(image.Pt(margin, margin))
	if available.X < 0 {
		available.X = 0
	}
	if available.Y < 0 {
		available.Y = 0
	}

	minSize := image.Pt(gtx.Dp(unit.Dp(240)), gtx.Dp(unit.Dp(150)))
	maxSize := image.Pt(gtx.Dp(unit.Dp(480)), gtx.Dp(unit.Dp(280)))
	preferred := image.Pt(int(float32(viewport.X)*0.46), int(float32(viewport.Y)*0.34))

	width := min(max(preferred.X, minSize.X), maxSize.X)
	height := min(max(preferred.Y, minSize.Y), maxSize.Y)
	width = min(width, available.X)
	height = min(height, available.Y)

	return image.Pt(width, height)
}

func responsivePopupPosition(gtx layout.Context, viewport, popupSize image.Point) image.Point {
	margin := gtx.Dp(unit.Dp(24))
	x := viewport.X - popupSize.X - margin
	y := (viewport.Y - popupSize.Y) / 2
	if viewport.X < gtx.Dp(unit.Dp(520)) {
		x = (viewport.X - popupSize.X) / 2
	}
	return image.Pt(max(0, x), max(0, y))
}

type PopupControls struct {
	Brightness widget.Float
	Alpha      widget.Float
}

func (c *PopupControls) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(unit.Dp(6)),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Caption(th, "明暗："+strconv.Itoa(c.BrightnessPercent()))
			label.Color = color.NRGBA{R: 0x4d, G: 0x5b, B: 0x68, A: 0xff}
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(170))
			gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(24))
			return material.Slider(th, &c.Brightness).Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Caption(th, "透明度："+itoa(c.AlphaByte()))
			label.Color = color.NRGBA{R: 0x4d, G: 0x5b, B: 0x68, A: 0xff}
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(170))
			gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(24))
			return material.Slider(th, &c.Alpha).Layout(gtx)
		}),
	)
}

func (c *PopupControls) BrightnessValue() float32 {
	return min(max(c.Brightness.Value, 0), 1)
}

func (c *PopupControls) BrightnessPercent() int {
	return int(math.Round(float64(c.BrightnessValue() * 100)))
}

func (c *PopupControls) AlphaByte() uint8 {
	alpha := int(math.Round(float64(c.Alpha.Value * 255)))
	alpha = min(max(alpha, 0), 255)
	return uint8(alpha)
}

func itoa(v uint8) string {
	return strconv.Itoa(int(v))
}

func readableTextColor(bg color.NRGBA) color.NRGBA {
	luma := 0.299*float32(bg.R) + 0.587*float32(bg.G) + 0.114*float32(bg.B)
	if luma < 140 {
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	}
	return color.NRGBA{R: 0x16, G: 0x24, B: 0x2f, A: 0xff}
}

func shadowPopup(gtx layout.Context, th *material.Theme, closeButton *widget.Clickable, popupShape ui.ShadowShape, preset shadowPreset, popupSize image.Point) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			bounds := image.Rectangle{Max: gtx.Constraints.Min}
			ui.DrawShadow(gtx, bounds, popupShape, ui.PopupShadow(preset.Color))
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, popupShape.ClipOp(gtx, bounds))
			return layout.Dimensions{Size: gtx.Constraints.Min}
		},
		func(gtx layout.Context) layout.Dimensions {
			contentInset := unit.Dp(24)
			if popupSize.Y < gtx.Dp(unit.Dp(210)) {
				contentInset = unit.Dp(16)
			}
			layout.Inset{
				Top:    contentInset,
				Right:  unit.Dp(52),
				Bottom: contentInset,
				Left:   contentInset,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
					Gap:  gtx.Dp(unit.Dp(14)),
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H6(th, preset.Title)
						title.Color = color.NRGBA{R: 0x16, G: 0x24, B: 0x2f, A: 0xff}
						return title.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						body := material.Body2(th, "拖动色轮和滑块，阴影会跟着变化。")
						body.Color = color.NRGBA{R: 0x4d, G: 0x5b, B: 0x68, A: 0xff}
						return body.Layout(gtx)
					}),
				)
			})
			closeSize := image.Pt(gtx.Dp(unit.Dp(32)), gtx.Dp(unit.Dp(32)))
			closePos := image.Pt(popupSize.X-closeSize.X-gtx.Dp(unit.Dp(12)), gtx.Dp(unit.Dp(12)))
			stack := op.Offset(closePos).Push(gtx.Ops)
			gtx.Constraints = layout.Exact(closeSize)
			closeIconButton(gtx, closeButton)
			stack.Pop()
			return layout.Dimensions{Size: popupSize}
		},
	)
}

func closeIconButton(gtx layout.Context, closeButton *widget.Clickable) layout.Dimensions {
	return closeButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := image.Pt(gtx.Dp(unit.Dp(32)), gtx.Dp(unit.Dp(32)))
		gtx.Constraints = layout.Exact(size)

		bg := color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}
		iconColor := color.NRGBA{R: 0x5d, G: 0x68, B: 0x74, A: 0xff}
		if closeButton.Hovered() {
			bg = color.NRGBA{R: 0xee, G: 0xf1, B: 0xf4, A: 0xff}
			iconColor = color.NRGBA{R: 0x2d, G: 0x36, B: 0x42, A: 0xff}
		}

		ellipse := clip.Ellipse(image.Rectangle{Max: size})
		paint.FillShape(gtx.Ops, bg, ellipse.Op(gtx.Ops))
		border := clip.Stroke{
			Path:  ellipse.Path(gtx.Ops),
			Width: float32(gtx.Dp(unit.Dp(1))),
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 0xdb, G: 0xe1, B: 0xe8, A: 0xff})
		border.Pop()

		drawCloseGlyph(gtx, size, iconColor)
		return layout.Dimensions{Size: size}
	})
}

func drawCloseGlyph(gtx layout.Context, size image.Point, col color.NRGBA) {
	inset := float32(gtx.Dp(unit.Dp(10)))
	maxX := float32(size.X) - inset
	maxY := float32(size.Y) - inset
	minX := inset
	minY := inset
	drawCloseLine(gtx, f32.Pt(minX, minY), f32.Pt(maxX, maxY), col)
	drawCloseLine(gtx, f32.Pt(maxX, minY), f32.Pt(minX, maxY), col)
}

func drawCloseLine(gtx layout.Context, from, to f32.Point, col color.NRGBA) {
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(from)
	p.LineTo(to)
	stroke := clip.Stroke{
		Path:  p.End(),
		Width: float32(gtx.Dp(unit.Dp(2))),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

type ColorWheel struct {
	drag  gesture.Drag
	hue   float32
	sat   float32
	cache colorWheelCache
}

type colorWheelCache struct {
	size image.Point
	op   paint.ImageOp
}

func (w *ColorWheel) ShadowColor(brightness float32) color.NRGBA {
	if w.sat == 0 {
		w.hue = 220
		w.sat = .72
	}
	col := hsvToNRGBA(w.hue, w.sat, brightness)
	col.A = 0x98
	return col
}

func (w *ColorWheel) Layout(gtx layout.Context, diameter unit.Dp) layout.Dimensions {
	size := image.Pt(gtx.Dp(diameter), gtx.Dp(diameter))
	gtx.Constraints = layout.Exact(size)
	for {
		e, ok := w.drag.Update(gtx.Metric, gtx.Source, gesture.Both)
		if !ok {
			break
		}
		if e.Kind == pointer.Press || e.Kind == pointer.Drag {
			w.setFromPosition(e.Position, size)
		}
	}

	if w.cache.op.Size() == (image.Point{}) || w.cache.size != size {
		w.cache.size = size
		w.cache.op = paint.NewImageOp(buildColorWheelImage(size))
	}

	w.cache.op.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	defer clip.Ellipse(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
	w.drag.Add(gtx.Ops)
	w.drawSelector(gtx, size)
	return layout.Dimensions{Size: size}
}

func (w *ColorWheel) setFromPosition(pos f32.Point, size image.Point) {
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	dx := pos.X - center.X
	dy := pos.Y - center.Y
	radius := float32(min(size.X, size.Y)) / 2
	dist := float32(math.Hypot(float64(dx), float64(dy)))
	if radius <= 0 {
		return
	}
	if dist > radius {
		scale := radius / dist
		dx *= scale
		dy *= scale
		dist = radius
	}
	angle := float32(math.Atan2(float64(dy), float64(dx)) * 180 / math.Pi)
	if angle < 0 {
		angle += 360
	}
	w.hue = angle
	w.sat = dist / radius
}

func (w *ColorWheel) drawSelector(gtx layout.Context, size image.Point) {
	if w.sat == 0 {
		w.hue = 220
		w.sat = .72
	}
	radius := float32(min(size.X, size.Y)) / 2
	angle := float64(w.hue) * math.Pi / 180
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	x := center.X + float32(math.Cos(angle))*radius*w.sat
	y := center.Y + float32(math.Sin(angle))*radius*w.sat
	selectorSize := gtx.Dp(unit.Dp(20))
	rect := image.Rect(0, 0, selectorSize, selectorSize).Add(image.Pt(
		int(math.Round(float64(x)))-selectorSize/2,
		int(math.Round(float64(y)))-selectorSize/2,
	))
	defer clip.Stroke{
		Path:  clip.Ellipse(rect).Path(gtx.Ops),
		Width: float32(gtx.Dp(unit.Dp(4))),
	}.Op().Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
	defer clip.Stroke{
		Path:  clip.Ellipse(rect.Inset(gtx.Dp(unit.Dp(3)))).Path(gtx.Ops),
		Width: float32(gtx.Dp(unit.Dp(2))),
	}.Op().Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, color.NRGBA{R: 0x3a, G: 0x6f, B: 0xd8, A: 0xff})
}

func buildColorWheelImage(size image.Point) *image.NRGBA {
	img := image.NewNRGBA(image.Rectangle{Max: size})
	cx := float64(size.X-1) / 2
	cy := float64(size.Y-1) / 2
	radius := math.Min(float64(size.X), float64(size.Y)) / 2
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Hypot(dx, dy)
			coverage := radius + .5 - dist
			if coverage <= 0 {
				continue
			}
			if coverage > 1 {
				coverage = 1
			}
			hue := math.Atan2(dy, dx) * 180 / math.Pi
			if hue < 0 {
				hue += 360
			}
			sat := dist / radius
			col := hsvToNRGBA(float32(hue), float32(sat), 1)
			col.A = uint8(math.Round(coverage * 255))
			img.SetNRGBA(x, y, col)
		}
	}
	return img
}

func hsvToNRGBA(h, s, v float32) color.NRGBA {
	h = float32(math.Mod(float64(h), 360))
	if h < 0 {
		h += 360
	}
	s = min(max(s, 0), 1)
	v = min(max(v, 0), 1)
	c := v * s
	x := c * (1 - float32(math.Abs(math.Mod(float64(h/60), 2)-1)))
	m := v - c
	var r, g, b float32
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.NRGBA{
		R: clampByte((r + m) * 255),
		G: clampByte((g + m) * 255),
		B: clampByte((b + m) * 255),
		A: 0xff,
	}
}

func clampByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(math.Round(float64(v)))
}
