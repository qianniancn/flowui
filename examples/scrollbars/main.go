package main

import (
	"fmt"
	"image/color"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(ctx *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	muted := ctx.Theme().Palette.MutedForeground
	return surface.New(
		layout.Box(
			layout.Column(
				text.New("Scrollbars").Size(24),
				layout.Row(
					layout.Expanded(scrollPanel("Recent activity", "activity-scrollbar", activityItems(muted))),
					layout.Expanded(scrollPanel("Project files", "files-scrollbar", fileItems(muted))),
				).Gap(16),
				horizontalPanel(muted),
			).Gap(18),
		).Style(style.New().Padding(24).FillWidth().FillHeight()),
	).Variant(surface.Secondary)
}

func scrollPanel(title, key string, content ui.Widget) ui.Widget {
	return (surface.New(
		layout.Box(
			layout.Column(
				text.New(title).Size(15),
				layout.Expanded(layout.Scrollbar(content).Key(key)),
			).Gap(10),
		).Style(style.New().Padding(14).FillWidth().Height(300)),
	).Variant(surface.Default)).Style(style.New().Radius(8))
}

func activityItems(muted color.NRGBA) ui.Widget {
	items := make([]ui.Widget, 18)
	for index := range items {
		items[index] = layout.Row(
			icon.New(lucide.CircleCheckBig).Size(16),
			layout.Expanded(layout.Column(
				text.New(fmt.Sprintf("Task %02d completed", index+1)).Size(14),
				text.New(fmt.Sprintf("Updated %d minutes ago", (index+1)*3)).Size(12).Color(muted),
			).Gap(3)),
		).AlignMiddle().Gap(10)
	}
	return layout.Column(items...).Gap(12)
}

func fileItems(muted color.NRGBA) ui.Widget {
	items := make([]ui.Widget, 22)
	for index := range items {
		items[index] = layout.Row(
			icon.New(lucide.FileText).Size(16),
			layout.Expanded(text.New(fmt.Sprintf("document-%02d.md", index+1)).Size(14)),
			text.New(fmt.Sprintf("%d KB", 8+index*3)).Size(12).Color(muted),
		).AlignMiddle().Gap(10)
	}
	return layout.Column(items...).Gap(10)
}

func horizontalPanel(muted color.NRGBA) ui.Widget {
	items := make([]ui.Widget, 12)
	for index := range items {
		items[index] = (surface.New(
			layout.Box(
				layout.Column(
					text.New(fmt.Sprintf("Sprint %02d", index+1)).Size(14),
					text.New(fmt.Sprintf("%d deliverables", 4+index%5)).Size(12).Color(muted),
				).Gap(5),
			).Style(style.New().Padding(12).Width(132).Height(68)),
		).Variant(surface.Tertiary)).Style(style.New().Radius(6))
	}
	return (surface.New(
		layout.Box(
			layout.Column(
				text.New("Delivery timeline").Size(15),
				layout.Expanded(
					layout.Scrollbar(layout.Row(items...).Gap(10)).
						Horizontal().
						Overlay(),
				),
			).Gap(10),
		).Style(style.New().Padding(14).FillWidth().Height(122)),
	).Variant(surface.Default)).Style(style.New().Radius(8))
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Scrollbars"),
		window.Size(920, 680),
		window.Theme(base.Light()),
	)
}
