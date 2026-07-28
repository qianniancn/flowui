package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	next      int
	toasts    []ui.ToastItem
	placement ui.ToastPlacement
	lastEvent string
}

type ShowToast struct {
	title       string
	description string
	variant     ui.ToastVariant
	action      string
	loading     bool
	persistent  bool
}

type CloseToast struct{ key string }
type ToastAction struct{ key string }
type SetPlacement struct{ placement ui.ToastPlacement }
type ClearToasts struct{}

func Update(model *Model, msg any) {
	switch msg := msg.(type) {
	case ShowToast:
		model.next++
		key := fmt.Sprintf("toast-%d", model.next)
		item := ui.Toast(key, msg.title).
			Description(msg.description).
			Variant(msg.variant).
			Loading(msg.loading)
		if msg.action != "" {
			item = item.Action(msg.action)
		}
		if msg.persistent {
			item = item.Timeout(0)
		}
		model.toasts = append([]ui.ToastItem{item}, model.toasts...)
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
}

func View(_ *ui.Context, model Model, send ui.Send[any]) ui.Widget {
	return ui.Stack(
		ui.Stacked(ui.Center(
			ui.Box(
				ui.Column(
					ui.Text("FlowUI Toasts").Size(24),
					ui.Divider(),
					section("Variants",
						ui.Wrap(
							showButton("default", "Default", ShowToast{title: "Team invitation", description: "Bob invited you to join the FlowUI team"}, send),
							showButton("accent", "Accent", ShowToast{title: "2 credits left", description: "Upgrade for more credits", variant: ui.ToastAccent, action: "Upgrade"}, send),
							showButton("success", "Success", ShowToast{title: "Plan upgraded", description: "You can continue using FlowUI", variant: ui.ToastSuccess}, send),
							showButton("warning", "Warning", ShowToast{title: "No credits left", description: "Upgrade to continue", variant: ui.ToastWarning, action: "Upgrade"}, send),
							showButton("danger", "Danger", ShowToast{title: "Storage is full", description: "Remove files to release space", variant: ui.ToastDanger, action: "Remove"}, send),
						).Gap(10).LineGap(10),
					),
					section("States",
						ui.Row(
							showButton("loading", "Loading", ShowToast{title: "Uploading file...", description: "Please wait while the file is uploaded", loading: true, persistent: true}, send),
							showButton("persistent", "Persistent", ShowToast{title: "Important notification", description: "This stays open until dismissed", persistent: true}, send),
							ui.Button("clear", ui.Text("Clear")).Variant(ui.ButtonTertiary).OnClick(func() { send(ClearToasts{}) }),
						).Gap(10).AlignMiddle(),
					),
					section("Placement",
						ui.Wrap(
							placementButton("bottom-start", "Bottom start", ui.ToastBottomStart, send),
							placementButton("bottom", "Bottom", ui.ToastBottom, send),
							placementButton("bottom-end", "Bottom end", ui.ToastBottomEnd, send),
							placementButton("top-start", "Top start", ui.ToastTopStart, send),
							placementButton("top", "Top", ui.ToastTop, send),
							placementButton("top-end", "Top end", ui.ToastTopEnd, send),
						).Gap(10).LineGap(10),
					),
					ui.Text(model.lastEvent).Size(12),
				).Gap(20),
			).Style(ui.FillWidth().MaxWidth(760).Padding(24)),
		)),
		ui.Overlay(ui.ToastProvider("notifications", model.toasts).
			Placement(model.placement).
			OnAction(func(key string) { send(ToastAction{key: key}) }).
			OnClose(func(key string) { send(CloseToast{key: key}) })).Expanded(),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(12)
}

func showButton(key, label string, toast ShowToast, send ui.Send[any]) ui.Widget {
	return ui.Button(key, ui.Text(label)).Variant(ui.ButtonSecondary).Size(ui.ButtonSmall).OnClick(func() {
		send(toast)
	})
}

func placementButton(key, label string, placement ui.ToastPlacement, send ui.Send[any]) ui.Widget {
	return ui.Button(key, ui.Text(label)).Variant(ui.ButtonTertiary).Size(ui.ButtonSmall).OnClick(func() {
		send(SetPlacement{placement: placement})
		send(ShowToast{title: label, description: "Toast placement preview"})
	})
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
	ui.Run(Model{placement: ui.ToastBottom}, Update, View,
		ui.Title("FlowUI Toasts"),
		ui.Size(940, 680),
		ui.OnError(func(err error) { fmt.Println(err) }),
	)
}
