package main

import (
	"fmt"

	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
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

func Update(m *Model, msg Msg) {
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
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No alert dialog open"
	if m.Last != "" {
		status = m.Last
	}
	return ui.Stack(
		ui.Stacked(
			ui.Center(
				ui.Box(
					ui.Scroll("alert-dialogs",
						ui.Column(
							ui.Text("FlowUI AlertDialog").Size(24),
							ui.Text(status).Size(14),
							ui.Divider(),
							section("Default",
								buttonGrid(openButton("delete-project", "Delete project", ui.ButtonDanger, send)),
							),
							section("Statuses",
								buttonGrid(
									openButton("default", "Default", ui.ButtonSecondary, send),
									openButton("accent", "Accent", ui.ButtonSecondary, send),
									openButton("success", "Success", ui.ButtonSecondary, send),
									openButton("warning", "Warning", ui.ButtonSecondary, send),
									openButton("danger", "Danger", ui.ButtonDangerSoft, send),
								),
							),
							section("Placements",
								buttonGrid(
									openButton("top", "Top", ui.ButtonSecondary, send),
									openButton("center", "Center", ui.ButtonSecondary, send),
									openButton("bottom", "Bottom", ui.ButtonSecondary, send),
								),
							),
							section("Sizes and backdrop",
								buttonGrid(
									openButton("xs", "XSmall", ui.ButtonSecondary, send),
									openButton("sm", "Small", ui.ButtonSecondary, send),
									openButton("lg", "Large", ui.ButtonSecondary, send),
									openButton("cover", "Cover", ui.ButtonSecondary, send),
									openButton("blur", "Blur", ui.ButtonSecondary, send),
									openButton("transparent", "Transparent", ui.ButtonSecondary, send),
								),
							),
							section("Behavior and customization",
								buttonGrid(
									openButton("dismissable", "Backdrop closes", ui.ButtonSecondary, send),
									openButton("escape", "Escape closes", ui.ButtonSecondary, send),
									openButton("custom-icon", "Custom icon", ui.ButtonSecondary, send),
								),
							),
						).Gap(18),
					).Vertical(),
				).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
			),
		),
		ui.Overlay(alertDialogLayer(m, send)).Expanded(),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func buttonGrid(children ...ui.Widget) ui.Widget {
	items := make([]ui.Widget, 0, len(children))
	for _, child := range children {
		items = append(items, ui.Box(child))
	}
	return ui.Wrap(items...).Gap(8).AlignMiddle()
}

func openButton(key, label string, variant ui.ButtonVariant, send ui.Send[Msg]) ui.Widget {
	return ui.Button("open-"+key, ui.Text(label)).
		Variant(variant).
		OnClick(func() { send(Open{Key: key}) })
}

func alertDialogLayer(m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		dialog("delete-project", m, send,
			"Delete project permanently?",
			"This will permanently delete My Awesome Project and all of its data. This action cannot be undone.",
			ui.AlertDialogDanger,
		),
		dialog("default", m, send, "Leave this page?", "Your current view will be closed.", ui.AlertDialogDefault),
		dialog("accent", m, send, "Sign out of your account?", "You'll need to sign in again to access your account.", ui.AlertDialogAccent),
		dialog("success", m, send, "Complete this task?", "This will mark the task as complete and notify team members.", ui.AlertDialogSuccess),
		dialog("warning", m, send, "Discard unsaved changes?", "Your unsaved changes will be permanently lost.", ui.AlertDialogWarning),
		dialog("danger", m, send, "Delete your account?", "This action is irreversible and removes all account data.", ui.AlertDialogDanger),
		dialog("top", m, send, "Top placement", placementDescription("top"), ui.AlertDialogAccent).Placement(ui.AlertDialogTop),
		dialog("center", m, send, "Center placement", placementDescription("center"), ui.AlertDialogAccent).Placement(ui.AlertDialogCenter),
		dialog("bottom", m, send, "Bottom placement", placementDescription("bottom"), ui.AlertDialogAccent).Placement(ui.AlertDialogBottom),
		dialog("xs", m, send, "XSmall dialog", sizeDescription("xs"), ui.AlertDialogAccent).Size(ui.AlertDialogXSmall),
		dialog("sm", m, send, "Small dialog", sizeDescription("sm"), ui.AlertDialogAccent).Size(ui.AlertDialogSmall),
		dialog("lg", m, send, "Large dialog", sizeDescription("lg"), ui.AlertDialogAccent).Size(ui.AlertDialogLarge),
		dialog("cover", m, send, "Cover dialog", sizeDescription("cover"), ui.AlertDialogAccent).Size(ui.AlertDialogCover),
		dialog("blur", m, send, "Blur backdrop", "The backdrop uses the blur-style treatment.", ui.AlertDialogAccent).Backdrop(ui.AlertDialogBackdropBlur),
		dialog("transparent", m, send, "Transparent backdrop", "The page remains visible while input is still blocked.", ui.AlertDialogAccent).Backdrop(ui.AlertDialogBackdropTransparent),
		dialog("dismissable", m, send, "Backdrop dismissal enabled", "Click outside this dialog to close it.", ui.AlertDialogAccent).Dismissable(true),
		dialog("escape", m, send, "Keyboard dismissal enabled", "Press Escape to close this dialog.", ui.AlertDialogAccent).KeyboardDismissDisabled(false),
		dialog("custom-icon", m, send, "Reset your password?", "We'll send a reset link to your email address.", ui.AlertDialogWarning).
			Icon(ui.Icon(lucide.KeyRound).Size(20)),
	)
}

func dialog(key string, m Model, send ui.Send[Msg], title, description string, status ui.AlertDialogStatus) ui.AlertDialogWidget {
	return ui.AlertDialog(key, m.Active == key, title, description).
		Status(status).
		Footer(dialogFooter(key, status, send)).
		OnOpenChange(func(open bool) {
			if !open {
				send(Close{Action: "closed"})
			}
		})
}

func dialogFooter(key string, status ui.AlertDialogStatus, send ui.Send[Msg]) ui.Widget {
	confirmVariant := ui.ButtonPrimary
	confirmLabel := "Confirm"
	if status == ui.AlertDialogDanger {
		confirmVariant = ui.ButtonDanger
		confirmLabel = "Delete"
	}
	return ui.Row(
		ui.Button(key+"-cancel", ui.Text("Cancel")).
			Variant(ui.ButtonTertiary).
			OnClick(func() { send(Close{Action: "cancelled"}) }),
		ui.Button(key+"-confirm", ui.Text(confirmLabel)).
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
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI AlertDialog"),
		ui.Size(900, 720),
	)
}
