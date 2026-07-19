package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
	lucide "github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Next   int
	Toasts []ui.ToastItem
}

type Msg any
type ShowNotification struct{}
type CloseNotification string

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case ShowNotification:
		model.Next++
		toast := ui.Toast(fmt.Sprintf("status-notification-%d", model.Next), "Status bar action").
			Description("The notification button was clicked.").
			Variant(ui.ToastSuccess)
		model.Toasts = append([]ui.ToastItem{toast}, model.Toasts...)
	case CloseNotification:
		model.Toasts = removeToast(model.Toasts, string(msg))
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := ui.StatusBar(statusLeft(), statusRight(send))

	return ui.Stack(
		ui.Stacked(
			ui.Surface(
				ui.Column(
					appBar(),
					ui.Divider(),
					ui.Expanded(workspace()),
					status,
				),
			).Variant(ui.SurfaceDefault),
		),
		ui.Overlay(
			ui.ToastProvider("status-notifications", model.Toasts).
				Width(360).
				Placement(ui.ToastBottomEnd).
				Offset(44).
				OnClose(func(key string) { send(CloseNotification(key)) }),
		).Expanded(),
	)
}

func appBar() ui.Widget {
	return ui.Box(
		ui.Row(
			ui.Icon(lucide.Braces).Size(18),
			ui.Text("FlowUI Editor").Size(14),
			ui.Expanded(ui.Spacer(0, 0)),
		).AlignMiddle().Gap(10),
	).Padding(8).FillWidth()
}

func workspace() ui.Widget {
	return ui.SplitPane("status-workspace", fileExplorer(), editor()).
		DefaultRatio(.24).
		MinFirst(180).
		MinSecond(420).
		Label("Resize explorer and editor")
}

func fileExplorer() ui.Widget {
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Text("EXPLORER").Size(12),
				ui.Divider(),
				ui.Row(ui.Icon(lucide.Braces).Size(14), ui.Text("main.go").Size(13)).AlignMiddle().Gap(7),
				ui.Row(ui.Icon(lucide.Braces).Size(14), ui.Text("go.mod").Size(13)).AlignMiddle().Gap(7),
			).Gap(10),
		).Padding(12).FillWidth().FillHeight(),
	).Variant(ui.SurfaceSecondary)
}

func editor() ui.Widget {
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Text("main.go").Size(13),
				ui.Divider(),
				ui.Text("package main").Size(14),
				ui.Spacer(0, 6),
				ui.Text("func main() {").Size(14),
				ui.Text("    runApplication()").Size(14),
				ui.Text("}").Size(14),
			).Gap(7),
		).Padding(14).FillWidth().FillHeight(),
	).Variant(ui.SurfaceDefault)
}

func statusLeft() ui.Widget {
	return ui.Row(
		ui.Tooltip(
			"branch-status",
			ui.Row(ui.Icon(lucide.GitBranch).Size(13), ui.Text("main").Size(12)).AlignMiddle().Gap(5),
			ui.Text("Current Git branch"),
		).Placement(ui.TooltipTop).Delay(0),
		ui.Row(ui.Icon(lucide.CircleCheck).Size(13), ui.Text("No problems").Size(12)).AlignMiddle().Gap(5),
	).AlignMiddle().Gap(12)
}

func statusRight(send ui.Send[Msg]) ui.Widget {
	return ui.Row(
		ui.Text("Ln 24, Col 8").Size(12),
		ui.Text("Spaces: 4").Size(12),
		ui.Text("UTF-8").Size(12),
		ui.Row(ui.Icon(lucide.Wifi).Size(13), ui.Text("Connected").Size(12)).AlignMiddle().Gap(5),
		ui.Tooltip(
			"notifications-status",
			ui.Button("show-status-notification", ui.Icon(lucide.Bell).Size(13)).
				Variant(ui.ButtonGhost).
				Size(ui.ButtonSmall).
				IconOnly().
				OnClick(func() { send(ShowNotification{}) }),
			ui.Text("Show notification"),
		).Placement(ui.TooltipTop).Delay(0),
	).AlignMiddle().Gap(12)
}

func removeToast(items []ui.ToastItem, key string) []ui.ToastItem {
	result := make([]ui.ToastItem, 0, len(items))
	for _, item := range items {
		if item.Key() != key {
			result = append(result, item)
		}
	}
	return result
}

func main() {
	ui.Run(
		Model{},
		Update,
		View,
		ui.Title("FlowUI StatusBar"),
		ui.Size(960, 640),
	)
}
