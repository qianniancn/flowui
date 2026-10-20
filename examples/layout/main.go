package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}

type Msg any

func Update(_ *Model, _ Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	cards := []ui.Widget{
		statCard("Open", "18"),
		statCard("Done", "42"),
		statCard("Late", "3"),
		statCard("Queued", "9"),
	}

	tags := make([]ui.Widget, 0, 8)
	for i := 1; i <= 8; i++ {
		tags = append(tags,
			layout.Box(text.New(fmt.Sprintf("Tag %d", i))).
				Style(style.New().PaddingTop(6).PaddingRight(10).PaddingBottom(6).PaddingLeft(10).Margin(2)),
		)
	}

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Layout").Size(24),
				layout.Divider(),
				layout.Row(
					text.New("Overview").Size(18),
					layout.Box(layout.Divider()).Style(style.New().Height(24)),
					text.New("Responsive cards").Size(18),
				).Gap(12).AlignMiddle(),
				layout.AutoGrid(160, cards...).Gap(12),
				layout.Box(
					layout.AspectRatio(16.0/9.0,
						layout.Stack(
							layout.Stacked(
								layout.Box(text.New("Aspect preview")).
									Style(style.New().FillWidth().
										FillHeight()).
									Align(layout.AlignCenter),
							),
							layout.Overlay(
								layout.Box(text.New("Overlay")).
									Style(style.New().PaddingTop(6).
										PaddingRight(10).
										PaddingBottom(6).
										PaddingLeft(10)),
							).Align(layout.AlignTopEnd),
						),
					),
				),
				layout.Wrap(tags...).Gap(8).AlignMiddle(),
			).Gap(16),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func statCard(label, value string) ui.Widget {
	return layout.Box(
		layout.Column(
			text.New(label).Size(14),
			text.New(value).Size(28),
		).Gap(6),
	).Style(style.New().FillWidth().Padding(16))
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Layout"),
		window.Size(900, 600),
		window.Theme(base.Light()),
	)
}
