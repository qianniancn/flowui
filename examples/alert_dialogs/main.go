package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alertdialog"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Active string
	Last   string
}

type Msg any

type Open struct {
	Key string
}

type Close struct {
	Action string
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Open:
		m.Active = msg.Key
		m.Last = "Opened " + msg.Key
	case Close:
		if m.Active != "" {
			m.Last = fmt.Sprintf("%s: %s", m.Active, msg.Action)
		}
		m.Active = ""
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No alert dialog open"
	if m.Last != "" {
		status = m.Last
	}
	return layout.Stack(
		layout.Stacked(
			layout.Center(
				layout.Box(
					layout.Scroll(layout.Column(
						text.New("FlowUI AlertDialog").Size(24),
						text.New(status).Size(14),
						layout.Divider(),
						section("Default",
							buttonGrid(openButton("delete-project", "Delete project", button.Danger, send)),
						),
						section("Statuses",
							buttonGrid(
								openButton("default", "Default", button.Secondary, send),
								openButton("accent", "Accent", button.Secondary, send),
								openButton("success", "Success", button.Secondary, send),
								openButton("warning", "Warning", button.Secondary, send),
								openButton("danger", "Danger", button.DangerSoft, send),
							),
						),
						section("Placements",
							buttonGrid(
								openButton("top", "Top", button.Secondary, send),
								openButton("center", "Center", button.Secondary, send),
								openButton("bottom", "Bottom", button.Secondary, send),
							),
						),
						section("Sizes and backdrop",
							buttonGrid(
								openButton("xs", "XSmall", button.Secondary, send),
								openButton("sm", "Small", button.Secondary, send),
								openButton("lg", "Large", button.Secondary, send),
								openButton("cover", "Cover", button.Secondary, send),
								openButton("blur", "Blur", button.Secondary, send),
								openButton("transparent", "Transparent", button.Secondary, send),
							),
						),
						section("Behavior and customization",
							buttonGrid(
								openButton("dismissable", "Backdrop closes", button.Secondary, send),
								openButton("escape", "Escape closes", button.Secondary, send),
								openButton("custom-icon", "Custom icon", button.Secondary, send),
							),
						),
					).Gap(18),
					).Vertical(),
				).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
			),
		),
		layout.Overlay(alertDialogLayer(m, send)).Expanded(),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func buttonGrid(children ...ui.Widget) ui.Widget {
	items := make([]ui.Widget, 0, len(children))
	for _, child := range children {
		items = append(items, layout.Box(child))
	}
	return layout.Wrap(items...).Gap(8).AlignMiddle()
}

func openButton(key, label string, variant button.Variant, send ui.Send[Msg]) ui.Widget {
	return button.New(label).Key("open-" + key).
		Variant(variant).
		OnClick(func() { send(Open{Key: key}) })
}

func alertDialogLayer(m Model, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		dialog("delete-project", m, send,
			"Delete project permanently?",
			"This will permanently delete My Awesome Project and all of its data. This action cannot be undone.",
			alertdialog.StatusDanger,
		),
		dialog("default", m, send, "Leave this page?", "Your current view will be closed.", alertdialog.StatusDefault),
		dialog("accent", m, send, "Sign out of your account?", "You'll need to sign in again to access your account.", alertdialog.StatusAccent),
		dialog("success", m, send, "Complete this task?", "This will mark the task as complete and notify team members.", alertdialog.StatusSuccess),
		dialog("warning", m, send, "Discard unsaved changes?", "Your unsaved changes will be permanently lost.", alertdialog.StatusWarning),
		dialog("danger", m, send, "Delete your account?", "This action is irreversible and removes all account data.", alertdialog.StatusDanger),
		dialog("top", m, send, "Top placement", placementDescription("top"), alertdialog.StatusAccent).Placement(alertdialog.PlacementTop),
		dialog("center", m, send, "Center placement", placementDescription("center"), alertdialog.StatusAccent).Placement(alertdialog.PlacementCenter),
		dialog("bottom", m, send, "Bottom placement", placementDescription("bottom"), alertdialog.StatusAccent).Placement(alertdialog.PlacementBottom),
		dialog("xs", m, send, "XSmall dialog", sizeDescription("xs"), alertdialog.StatusAccent).Size(alertdialog.SizeXSmall),
		dialog("sm", m, send, "Small dialog", sizeDescription("sm"), alertdialog.StatusAccent).Size(alertdialog.SizeSmall),
		dialog("lg", m, send, "Large dialog", sizeDescription("lg"), alertdialog.StatusAccent).Size(alertdialog.SizeLarge),
		dialog("cover", m, send, "Cover dialog", sizeDescription("cover"), alertdialog.StatusAccent).Size(alertdialog.SizeCover),
		dialog("blur", m, send, "Blur backdrop", "The backdrop uses the blur-style treatment.", alertdialog.StatusAccent).Backdrop(alertdialog.BackdropBlur),
		dialog("transparent", m, send, "Transparent backdrop", "The page remains visible while input is still blocked.", alertdialog.StatusAccent).Backdrop(alertdialog.BackdropTransparent),
		dialog("dismissable", m, send, "Backdrop dismissal enabled", "Click outside this dialog to close it.", alertdialog.StatusAccent).Dismissable(true),
		dialog("escape", m, send, "Keyboard dismissal enabled", "Press Escape to close this dialog.", alertdialog.StatusAccent).KeyboardDismissDisabled(false),
		dialog("custom-icon", m, send, "Reset your password?", "We'll send a reset link to your email address.", alertdialog.StatusWarning).
			Icon(icon.New(lucide.KeyRound).Size(20)),
	)
}

func dialog(key string, m Model, send ui.Send[Msg], title, description string, status alertdialog.Status) alertdialog.Widget {
	return alertdialog.New(title, description).Key(key).Open(m.Active == key).
		Status(status).
		Footer(dialogFooter(key, status, send)).
		OnOpenChange(func(open bool) {
			if !open {
				send(Close{Action: "closed"})
			}
		})
}

func dialogFooter(key string, status alertdialog.Status, send ui.Send[Msg]) ui.Widget {
	confirmVariant := button.Primary
	confirmLabel := "Confirm"
	if status == alertdialog.StatusDanger {
		confirmVariant = button.Danger
		confirmLabel = "Delete"
	}
	return layout.Row(
		button.New("Cancel").Key(key+"-cancel").
			Variant(button.Tertiary).
			OnClick(func() { send(Close{Action: "cancelled"}) }),
		button.New(confirmLabel).Key(key+"-confirm").
			Variant(confirmVariant).
			OnClick(func() { send(Close{Action: "confirmed"}) }),
	).Gap(8).AlignMiddle()
}

func placementDescription(placement string) string {
	return fmt.Sprintf("This alert dialog is positioned at the %s of the viewport.", placement)
}

func sizeDescription(size string) string {
	return fmt.Sprintf("The %s preset changes the maximum dialog width while preserving the standard padding.", size)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI AlertDialog"),
		window.Size(900, 720),
		window.Theme(base.Light()),
	)
}
