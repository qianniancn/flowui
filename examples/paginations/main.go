package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct {
	Page int
}

type Msg struct {
	Page int
}

func Update(model *Model, msg Msg) {
	if msg.Page > 0 {
		model.Page = msg.Page
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Pagination").Size(24),
				section("Controlled with summary",
					ui.Pagination("controlled", model.Page, 12).
						Summary(ui.Text("61 to 72 of 144 results").Size(14)).
						OnChange(func(page int) { send(Msg{Page: page}) }),
				),
				section("Small",
					ui.Pagination("small", model.Page, 12).
						Size(ui.PaginationSmall).
						OnChange(func(page int) { send(Msg{Page: page}) }),
				),
				section("Large",
					ui.Pagination("large", model.Page, 12).
						Size(ui.PaginationLarge).
						OnChange(func(page int) { send(Msg{Page: page}) }),
				),
				section("Disabled", ui.Pagination("disabled", 3, 8).Disabled(true)),
			).Gap(24),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func main() {
	ui.Run(Model{Page: 6}, Update, View,
		ui.Title("FlowUI Pagination"),
		ui.Size(900, 620),
	)
}
