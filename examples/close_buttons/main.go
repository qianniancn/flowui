package main

import (
	"image"
	"strconv"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	clicks int
}

type Increment struct{}

func Update(model *Model, msg Increment) {
	model.clicks++
}

func View(_ *ui.Context, model Model, send ui.Send[Increment]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Close Buttons").Size(24),
				ui.Divider(),
				section("Default",
					ui.Row(
						labeled("Default", ui.CloseButton("default")),
						labeled("Disabled", ui.CloseButton("disabled").Disabled(true)),
						labeled("Custom icon", ui.CloseButton("custom").Icon(circleCloseIcon{})),
					).Gap(32).AlignMiddle(),
				),
				section("Interactive",
					ui.Row(
						ui.CloseButton("interactive").
							Label("Increment close count").
							OnClick(func() { send(Increment{}) }),
						ui.Text("Clicked: "+strconv.Itoa(model.clicks)),
					).Gap(12).AlignMiddle(),
				),
			).Gap(20),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(640)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(12)
}

func labeled(label string, button ui.Widget) ui.Widget {
	return ui.Column(ui.Center(button), ui.Text(label).Size(12)).Gap(8).AlignMiddle()
}

type circleCloseIcon struct{}

func (circleCloseIcon) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Min
	col := ctx.ForegroundColor()
	strokeWidth := float32(1)
	rect := image.Rectangle{Max: size}.Inset(1)
	stroke := clip.Stroke{Path: clip.Ellipse(rect).Path(gtx.Ops), Width: strokeWidth}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()

	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	half := float32(min(size.X, size.Y)) * 0.18
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X-half, center.Y-half))
	path.LineTo(f32.Pt(center.X+half, center.Y+half))
	path.MoveTo(f32.Pt(center.X+half, center.Y-half))
	path.LineTo(f32.Pt(center.X-half, center.Y+half))
	lines := clip.Stroke{Path: path.End(), Width: strokeWidth}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	lines.Pop()
	return layout.Dimensions{Size: size}
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Close Buttons"),
		ui.Size(720, 460),
	)
}
