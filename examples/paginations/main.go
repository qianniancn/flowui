package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/pagination"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Page int
}

type Msg struct {
	Page int
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Page > 0 {
		model.Page = msg.Page
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Pagination").Size(24),
				section("Controlled with summary",
					pagination.New("controlled", model.Page, 12).
						Summary(text.New("61 to 72 of 144 results").Size(14)).
						OnChange(func(page int) { send(Msg{Page: page}) }),
				),
				section("Small",
					pagination.New("small", model.Page, 12).
						Size(pagination.SizeSmall).
						OnChange(func(page int) { send(Msg{Page: page}) }),
				),
				section("Large",
					pagination.New("large", model.Page, 12).
						Size(pagination.SizeLarge).
						OnChange(func(page int) { send(Msg{Page: page}) }),
				),
				section("Disabled", pagination.New("disabled", 3, 8).Disabled(true)),
			).Gap(24),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{Page: 6},
		Update, View), window.Title("FlowUI Pagination"),
		window.Size(900, 620),
		window.Theme(base.Light()),
	)
}
