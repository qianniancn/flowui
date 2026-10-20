package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	toastpkg "github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	next      int
	toasts    []toastpkg.Item
	placement toastpkg.Placement
	lastEvent string
}

type ShowToast struct {
	title       string
	description string
	variant     toastpkg.Variant
	action      string
	loading     bool
	persistent  bool
	compact     bool
}

type CloseToast struct{ key string }
type ToastAction struct{ key string }
type SetPlacement struct{ placement toastpkg.Placement }
type ClearToasts struct{}

func Update(model *Model, msg any) ui.Cmd[any] {
	switch msg := msg.(type) {
	case ShowToast:
		model.next++
		key := fmt.Sprintf("toast-%d", model.next)
		item := toastpkg.NewItem(key, msg.title).
			Description(msg.description).
			Variant(msg.variant).
			Loading(msg.loading)
		if msg.action != "" {
			item = item.Action(msg.action)
		}
		if msg.persistent {
			item = item.Timeout(0)
		}
		if msg.compact {
			item = item.Compact(true)
		}
		model.toasts = append([]toastpkg.Item{item}, model.toasts...)
	case CloseToast:
		model.toasts = removeToast(model.toasts, msg.key)
		model.lastEvent = "Closed " + msg.key
	case ToastAction:
		model.toasts = removeToast(model.toasts, msg.key)
		model.lastEvent = "Action from " + msg.key
	case SetPlacement:
		model.placement = msg.placement
	case ClearToasts:
		model.toasts = nil
		model.lastEvent = "Cleared all toasts"
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[any]) ui.Widget {
	return layout.Stack(
		layout.Stacked(layout.Center(
			layout.Box(
				layout.Column(
					text.New("FlowUI Toasts").Size(24),
					layout.Divider(),
					section("Variants",
						layout.Wrap(
							showButton("default", "Default", ShowToast{title: "Team invitation", description: "Bob invited you to join the FlowUI team"}, send),
							showButton("accent", "Accent", ShowToast{title: "2 credits left", description: "Upgrade for more credits", variant: toastpkg.Accent, action: "Upgrade"}, send),
							showButton("success", "Success", ShowToast{title: "Plan upgraded", description: "You can continue using FlowUI", variant: toastpkg.Success}, send),
							showButton("warning", "Warning", ShowToast{title: "No credits left", description: "Upgrade to continue", variant: toastpkg.Warning, action: "Upgrade"}, send),
							showButton("danger", "Danger", ShowToast{title: "Storage is full", description: "Remove files to release space", variant: toastpkg.Danger, action: "Remove"}, send),
						).Gap(10).LineGap(10),
					),
					section("States",
						layout.Row(
							showButton("loading", "Loading", ShowToast{title: "Uploading file...", description: "Please wait while the file is uploaded", loading: true, persistent: true}, send),
							showButton("persistent", "Persistent", ShowToast{title: "Important notification", description: "This stays open until dismissed", persistent: true}, send),
							button.New("Clear").Variant(button.Tertiary).OnClick(func() { send(ClearToasts{}) }),
						).Gap(10).AlignMiddle(),
					),
					section("Messages",
						layout.Wrap(
							showButton("message-success", "Success message", ShowToast{title: "Saved", variant: toastpkg.Success, compact: true}, send),
							showButton("message-warning", "Warning message", ShowToast{title: "Check your connection", variant: toastpkg.Warning, compact: true}, send),
							showButton("message-danger", "Error message", ShowToast{title: "Upload failed", variant: toastpkg.Danger, compact: true}, send),
						).Gap(10).LineGap(10),
					),
					section("Placement",
						layout.Wrap(
							placementButton("bottom-start", "Bottom start", toastpkg.BottomStart, send),
							placementButton("bottom", "Bottom", toastpkg.Bottom, send),
							placementButton("bottom-end", "Bottom end", toastpkg.BottomEnd, send),
							placementButton("top-start", "Top start", toastpkg.TopStart, send),
							placementButton("top", "Top", toastpkg.Top, send),
							placementButton("top-end", "Top end", toastpkg.TopEnd, send),
						).Gap(10).LineGap(10),
					),
					text.New(model.lastEvent).Size(12),
				).Gap(20),
			).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
		)),
		layout.Overlay(toastpkg.New("notifications", model.toasts).
			Placement(model.placement).
			OnAction(func(key string) { send(ToastAction{key: key}) }).
			OnClose(func(key string) { send(CloseToast{key: key}) })).Expanded(),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(12)
}

func showButton(key, label string, toast ShowToast, send ui.Send[any]) ui.Widget {
	return button.New(label).Key(key).Variant(button.Secondary).Size(button.Small).OnClick(func() {
		send(toast)
	})
}

func placementButton(key, label string, placement toastpkg.Placement, send ui.Send[any]) ui.Widget {
	return button.New(label).Key(key).Variant(button.Tertiary).Size(button.Small).OnClick(func() {
		send(SetPlacement{placement: placement})
		send(ShowToast{title: label, description: "Toast placement preview"})
	})
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
	ui.Run(ui.NewProgram(Model{placement: toastpkg.Top},
		Update, View), window.Title("FlowUI Toasts"),
		window.Size(940, 680),
		window.OnError(func(err error) { fmt.Println(err) }),
		window.Theme(base.Light()),
	)
}
