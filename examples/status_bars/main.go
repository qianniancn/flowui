package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	toastpkg "github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tooltip"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Next   int
	Toasts []toastpkg.Item
}

type Msg any
type ShowNotification struct{}
type CloseNotification string

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case ShowNotification:
		model.Next++
		toast := toastpkg.NewItem(fmt.Sprintf("status-notification-%d", model.Next), "Status bar action").
			Description("The notification button was clicked.").
			Variant(toastpkg.Success)
		model.Toasts = append([]toastpkg.Item{toast}, model.Toasts...)
	case CloseNotification:
		model.Toasts = removeToast(model.Toasts, string(msg))
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := statusbar.New(statusLeft(), statusRight(send))

	return layout.Stack(
		layout.Stacked(
			surface.New(
				layout.Column(
					appBar(),
					layout.Divider(),
					layout.Expanded(workspace()),
					status,
				),
			).Variant(surface.Default),
		),
		layout.Overlay(
			toastpkg.New("status-notifications", model.Toasts).
				Width(360).
				Placement(toastpkg.BottomEnd).
				Offset(44).
				OnClose(func(key string) { send(CloseNotification(key)) }),
		).Expanded(),
	)
}

func appBar() ui.Widget {
	return layout.Box(
		layout.Row(
			icon.New(lucide.Braces).Size(18),
			text.New("FlowUI Editor").Size(14),
			layout.Expanded(layout.Spacer(0, 0)),
		).AlignMiddle().Gap(10),
	).Style(style.New().Padding(8).FillWidth())
}

func workspace() ui.Widget {
	return layout.SplitPane(fileExplorer(), editor()).
		DefaultRatio(.24).
		MinFirst(180).
		MinSecond(420).
		Label("Resize explorer and editor")
}

func fileExplorer() ui.Widget {
	return surface.New(
		layout.Box(
			layout.Column(
				text.New("EXPLORER").Size(12),
				layout.Divider(),
				layout.Row(icon.New(lucide.Braces).Size(14), text.New("main.go").Size(13)).AlignMiddle().Gap(7),
				layout.Row(icon.New(lucide.Braces).Size(14), text.New("go.mod").Size(13)).AlignMiddle().Gap(7),
			).Gap(10),
		).Style(style.New().Padding(12).FillWidth().FillHeight()),
	).Variant(surface.Secondary)
}

func editor() ui.Widget {
	return surface.New(
		layout.Box(
			layout.Column(
				text.New("main.go").Size(13),
				layout.Divider(),
				text.New("package main").Size(14),
				layout.Spacer(0, 6),
				text.New("func main() {").Size(14),
				text.New("    runApplication()").Size(14),
				text.New("}").Size(14),
			).Gap(7),
		).Style(style.New().Padding(14).FillWidth().FillHeight()),
	).Variant(surface.Default)
}

func statusLeft() ui.Widget {
	return layout.Row(
		tooltip.New("branch-status", layout.Row(icon.New(lucide.GitBranch).Size(13), text.New("main").Size(12)).AlignMiddle().Gap(5), "Current Git branch").Placement(tooltip.Top).Delay(0),
		layout.Row(icon.New(lucide.CircleCheck).Size(13), text.New("No problems").Size(12)).AlignMiddle().Gap(5),
	).AlignMiddle().Gap(12)
}

func statusRight(send ui.Send[Msg]) ui.Widget {
	return layout.Row(
		text.New("Ln 24, Col 8").Size(12),
		text.New("Spaces: 4").Size(12),
		text.New("UTF-8").Size(12),
		layout.Row(icon.New(lucide.Wifi).Size(13), text.New("Connected").Size(12)).AlignMiddle().Gap(5),
		tooltip.New("notifications-status", button.New("").Content(icon.New(lucide.Bell).Size(13)).
			Variant(button.Ghost).
			Size(button.Small).
			IconOnly().
			OnClick(func() { send(ShowNotification{}) }), "Show notification").Placement(tooltip.Top).Delay(0),
	).AlignMiddle().Gap(12)
}

func removeToast(items []toastpkg.Item, key string) []toastpkg.Item {
	result := make([]toastpkg.Item, 0, len(items))
	for _, item := range items {
		if item.Key() != key {
			result = append(result, item)
		}
	}
	return result
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI StatusBar"),
		window.Size(960, 640),
		window.Theme(base.Light()),
	)
}
