package main

import (
	"image"
	"strconv"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/closebutton"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	clicks int
}

type Increment struct{}

func Update(model *Model, msg Increment) ui.Cmd[Increment] {
	model.clicks++
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Increment]) ui.Widget {
	return flowlayout.Center(
		flowlayout.Box(
			flowlayout.Column(
				text.New("FlowUI Close Buttons").Size(24),
				flowlayout.Divider(),
				section("Default",
					flowlayout.Row(
						labeled("Default", closebutton.New("default")),
						labeled("Disabled", closebutton.New("disabled").Disabled(true)),
						labeled("Custom icon", closebutton.New("custom").Icon(circleCloseIcon{})),
					).Gap(32).AlignMiddle(),
				),
				section("Interactive",
					flowlayout.Row(
						closebutton.New("interactive").
							Label("Increment close count").
							OnClick(func() { send(Increment{}) }),
						text.New("Clicked: "+strconv.Itoa(model.clicks)),
					).Gap(12).AlignMiddle(),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(640).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return flowlayout.Column(text.New(title).Size(18), child).Gap(12)
}

func labeled(label string, button ui.Widget) ui.Widget {
	return flowlayout.Column(flowlayout.Center(button), text.New(label).Size(12)).Gap(8).AlignMiddle()
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
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Close Buttons"),
		window.Size(720, 460),
		window.Theme(base.Light()),
	)
}
