package main

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

func View(_ *ui.Context, _ struct{}, _ ui.Send[struct{}]) ui.Widget {
	longTitle := "Quarterly performance summary for the desktop application workspace"
	paragraph := "FlowUI keeps text layout predictable across narrow and wide desktop panes while preserving readable line breaks and truncation."
	code := "package main\n\nfunc main() {\n    println(\"FlowUI\")\n}"

	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Text").Size(24).Weight(font.SemiBold),
				section("Typography", layout.Column(
					text.New("Interface text with a regular sans-serif face."),
					text.New("A serif italic accent for editorial content.").Typeface("serif").FontStyle(font.Italic),
					text.New("MONOSPACED STATUS OUTPUT").Typeface("monospace").Size(13),
				).Gap(8)),
				section("Truncation", layout.Column(
					layout.Box(text.New(longTitle).MaxLines(1)).Style(style.New().Width(360)),
					layout.Box(text.New(paragraph).MaxLines(2).Wrap(text.WrapWords).LineHeight(21)).Style(style.New().Width(360)),
				).Gap(10)),
				section("Alignment", layout.Column(
					layout.Box(text.New("Start").Align(text.AlignStart)).Style(style.New().Width(360)),
					layout.Box(text.New("Center").Align(text.AlignCenter)).Style(style.New().Width(360)),
					layout.Box(text.New("End").Align(text.AlignEnd)).Style(style.New().Width(360)),
				).Gap(6)),
				section("Selectable", (surface.New(
					layout.Box(
						text.Selectable("code", code).
							Typeface("monospace").
							Size(14).
							LineHeight(22).
							MaxLines(5),
					).Style(style.New().FillWidth().Padding(16)),
				).Variant(surface.Secondary)).Style(style.New().Radius(8))),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(15).Weight(font.SemiBold),
		content,
	).Gap(9)
}

func main() {
	ui.Run(ui.NewProgram(struct{}{},
		func(*struct{}, struct{}) ui.Cmd[struct{}] { return nil }, View), window.Title("FlowUI Text"), window.Size(820, 680), window.Theme(base.Light()))
}
