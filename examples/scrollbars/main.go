package main

import (
	"fmt"
	"image/color"

	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) {}

func View(ctx *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	muted := ctx.Theme().Palette.MutedForeground
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Text("Scrollbars").Size(24),
				ui.Row(
					ui.Expanded(scrollPanel("Recent activity", "activity-scrollbar", activityItems(muted))),
					ui.Expanded(scrollPanel("Project files", "files-scrollbar", fileItems(muted))),
				).Gap(16),
				horizontalPanel(muted),
			).Gap(18),
		).Style(ui.Padding(24)).Style(ui.FillWidth()).Style(ui.FillHeight()),
	).Variant(ui.SurfaceSecondary)
}

func scrollPanel(title, key string, content ui.Widget) ui.Widget {
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Text(title).Size(15),
				ui.Expanded(ui.Scrollbar(key, content)),
			).Gap(10),
		).Style(ui.Padding(14)).Style(ui.FillWidth()).Style(ui.Height(300)),
	).Variant(ui.SurfaceDefault).Style(ui.Radius(8))
}

func activityItems(muted color.NRGBA) ui.Widget {
	items := make([]ui.Widget, 18)
	for index := range items {
		items[index] = ui.Row(
			ui.Icon(lucide.CircleCheckBig).Size(16),
			ui.Expanded(ui.Column(
				ui.Text(fmt.Sprintf("Task %02d completed", index+1)).Size(14),
				ui.Text(fmt.Sprintf("Updated %d minutes ago", (index+1)*3)).Size(12).Color(muted),
			).Gap(3)),
		).AlignMiddle().Gap(10)
	}
	return ui.Column(items...).Gap(12)
}

func fileItems(muted color.NRGBA) ui.Widget {
	items := make([]ui.Widget, 22)
	for index := range items {
		items[index] = ui.Row(
			ui.Icon(lucide.FileText).Size(16),
			ui.Expanded(ui.Text(fmt.Sprintf("document-%02d.md", index+1)).Size(14)),
			ui.Text(fmt.Sprintf("%d KB", 8+index*3)).Size(12).Color(muted),
		).AlignMiddle().Gap(10)
	}
	return ui.Column(items...).Gap(10)
}

func horizontalPanel(muted color.NRGBA) ui.Widget {
	items := make([]ui.Widget, 12)
	for index := range items {
		items[index] = ui.Surface(
			ui.Box(
				ui.Column(
					ui.Text(fmt.Sprintf("Sprint %02d", index+1)).Size(14),
					ui.Text(fmt.Sprintf("%d deliverables", 4+index%5)).Size(12).Color(muted),
				).Gap(5),
			).Style(ui.Padding(12)).Style(ui.Width(132)).Style(ui.Height(68)),
		).Variant(ui.SurfaceTertiary).Style(ui.Radius(6))
	}
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Text("Delivery timeline").Size(15),
				ui.Expanded(
					ui.Scrollbar("timeline-scrollbar", ui.Row(items...).Gap(10)).
						Horizontal().
						Overlay(),
				),
			).Gap(10),
		).Style(ui.Padding(14)).Style(ui.FillWidth()).Style(ui.Height(122)),
	).Variant(ui.SurfaceDefault).Style(ui.Radius(8))
}

func main() {
	ui.Run(
		Model{},
		Update,
		View,
		ui.Title("FlowUI Scrollbars"),
		ui.Size(920, 680),
	)
}
