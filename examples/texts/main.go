package main

import (
	"gioui.org/font"
	"github.com/qianniancn/FlowUI/ui"
)

func View(_ *ui.Context, _ struct{}, _ ui.Send[struct{}]) ui.Widget {
	longTitle := "Quarterly performance summary for the desktop application workspace"
	paragraph := "FlowUI keeps text layout predictable across narrow and wide desktop panes while preserving readable line breaks and truncation."
	code := "package main\n\nfunc main() {\n    println(\"FlowUI\")\n}"

	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Text").Size(24).Weight(font.SemiBold),
				section("Typography", ui.Column(
					ui.Text("Interface text with a regular sans-serif face."),
					ui.Text("A serif italic accent for editorial content.").Typeface("serif").FontStyle(font.Italic),
					ui.Text("MONOSPACED STATUS OUTPUT").Typeface("monospace").Size(13),
				).Gap(8)),
				section("Truncation", ui.Column(
					ui.Box(ui.Text(longTitle).MaxLines(1)).Width(360),
					ui.Box(ui.Text(paragraph).MaxLines(2).Wrap(ui.TextWrapWords).LineHeight(21)).Width(360),
				).Gap(10)),
				section("Alignment", ui.Column(
					ui.Box(ui.Text("Start").Align(ui.TextAlignStart)).Width(360),
					ui.Box(ui.Text("Center").Align(ui.TextAlignCenter)).Width(360),
					ui.Box(ui.Text("End").Align(ui.TextAlignEnd)).Width(360),
				).Gap(6)),
				section("Selectable", ui.Surface(
					ui.Box(
						ui.SelectableText("code", code).
							Typeface("monospace").
							Size(14).
							LineHeight(22).
							MaxLines(5),
					).FillWidth().Padding(16),
				).Variant(ui.SurfaceSecondary).Style(ui.Radius(8))),
			).Gap(18),
		).FillWidth().MaxWidth(720).Padding(24),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(15).Weight(font.SemiBold),
		content,
	).Gap(9)
}

func main() {
	ui.Run(struct{}{}, func(*struct{}, struct{}) {}, View, ui.Title("FlowUI Text"), ui.Size(820, 680))
}
