package main

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

func contextMenuTrigger() ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		trigger := ui.Surface(
			ui.Box(ui.Row(
				ui.Icon(lucide.MouseRight).Size(18),
				ui.Text("Right-click here").Size(14),
			).AlignMiddle().Gap(9)).Style(ui.Width(300).Padding(18)),
		).Variant(ui.SurfaceSecondary).Style(ui.Radius(10).BorderWidth(0))

		macro := op.Record(gtx.Ops)
		dims := trigger.Layout(ctx, gtx)
		call := macro.Stop()
		call.Add(gtx.Ops)

		drawContextMenuDashedBorder(gtx, dims.Size, ctx.Theme().Palette.Accent)
		return dims
	})
}

func drawContextMenuDashedBorder(gtx layout.Context, size image.Point, col color.NRGBA) {
	width := max(gtx.Dp(1.5), 1)
	inset := float32(width) / 2
	left := inset
	top := inset
	right := float32(size.X) - inset
	bottom := float32(size.Y) - inset
	if right <= left || bottom <= top {
		return
	}

	dash := float32(max(gtx.Dp(8), 1))
	gap := float32(max(gtx.Dp(5), 1))
	radius := min(float32(gtx.Dp(10)), min((right-left)/2, (bottom-top)/2))
	points := contextMenuRoundedRectPoints(left, top, right, bottom, radius)
	var path clip.Path
	path.Begin(gtx.Ops)
	appendContextMenuDashedPath(&path, points, dash, gap)

	stroke := clip.Stroke{Path: path.End(), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func contextMenuRoundedRectPoints(left, top, right, bottom, radius float32) []f32.Point {
	const cornerSteps = 8
	points := []f32.Point{
		f32.Pt(left+radius, top),
		f32.Pt(right-radius, top),
	}
	appendContextMenuArc(&points, right-radius, top+radius, radius, -math.Pi/2, 0, cornerSteps)
	points = append(points, f32.Pt(right, bottom-radius))
	appendContextMenuArc(&points, right-radius, bottom-radius, radius, 0, math.Pi/2, cornerSteps)
	points = append(points, f32.Pt(left+radius, bottom))
	appendContextMenuArc(&points, left+radius, bottom-radius, radius, math.Pi/2, math.Pi, cornerSteps)
	points = append(points, f32.Pt(left, top+radius))
	appendContextMenuArc(&points, left+radius, top+radius, radius, math.Pi, math.Pi*1.5, cornerSteps)
	return append(points, f32.Pt(left+radius, top))
}

func appendContextMenuArc(points *[]f32.Point, centerX, centerY, radius, start, end float32, steps int) {
	for step := 1; step <= steps; step++ {
		angle := start + (end-start)*float32(step)/float32(steps)
		*points = append(*points, f32.Pt(
			centerX+float32(math.Cos(float64(angle)))*radius,
			centerY+float32(math.Sin(float64(angle)))*radius,
		))
	}
}

func appendContextMenuDashedPath(path *clip.Path, points []f32.Point, dash, gap float32) {
	if len(points) < 2 {
		return
	}
	period := dash + gap
	distance := float32(0)
	for index := 1; index < len(points); index++ {
		from, to := points[index-1], points[index]
		dx := to.X - from.X
		dy := to.Y - from.Y
		length := float32(math.Hypot(float64(dx), float64(dy)))
		if length <= 0 {
			continue
		}
		for offset := float32(0); offset < length; {
			phase := float32(math.Mod(float64(distance+offset), float64(period)))
			step := period - phase
			draw := phase < dash
			if phase < dash {
				step = dash - phase
			}
			step = min(length-offset, step)
			if draw {
				end := offset + step
				path.MoveTo(f32.Pt(from.X+dx*offset/length, from.Y+dy*offset/length))
				path.LineTo(f32.Pt(from.X+dx*end/length, from.Y+dy*end/length))
			}
			offset += step
		}
		distance += length
	}
}
