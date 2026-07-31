package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

func statusPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Status & progress",
		demoSection{Title: "Alert", Content: ui.Column(
			ui.Alert("Information", "A neutral message for the current task."),
			ui.Alert("Update complete", "All component examples are available.").Status(ui.AlertSuccess),
			ui.Alert("Review required", "Check the generated output before publishing.").Status(ui.AlertWarning),
			ui.Alert("Build failed", "Resolve the reported errors and try again.").Status(ui.AlertDanger),
		).Gap(10)},
		demoSection{Title: "ProgressBar", Content: demoPanel(ui.Column(
			ui.ProgressBar("catalog-progress-default", 42).Label("Indexing").ShowValue(),
			ui.ProgressBar("catalog-progress-success", 76).Label("Upload").ShowValue().Color(ui.ProgressBarSuccess),
			ui.ProgressBar("catalog-progress-indeterminate", 0).Label("Syncing").Indeterminate(),
		).Gap(14))},
		demoSection{Title: "ProgressCircle", Content: demoPanel(demoRow(
			labeledProgressCircle("Small", ui.ProgressCircle("catalog-circle-small", 35).Size(ui.ProgressCircleSmall)),
			labeledProgressCircle("Medium", ui.ProgressCircle("catalog-circle-medium", 60)),
			labeledProgressCircle("Large", ui.ProgressCircle("catalog-circle-large", 82).Size(ui.ProgressCircleLarge).Color(ui.ProgressCircleSuccess)),
			labeledProgressCircle("Loading", ui.ProgressCircle("catalog-circle-loading", 0).Indeterminate()),
		))},
		demoSection{Title: "Meter", Content: demoPanel(ui.Column(
			ui.Meter("catalog-meter-storage", 68).Label("Storage").ShowValue(),
			ui.Meter("catalog-meter-memory", 84).Label("Memory").ShowValue().Color(ui.MeterWarning),
			ui.Meter("catalog-meter-budget", 750).Label("Budget").Range(0, 1000).
				ValueFormatter(func(value float64) string { return fmt.Sprintf("$%.0f", value) }).Color(ui.MeterSuccess),
		).Gap(14))},
		demoSection{Title: "Spinner", Content: demoPanel(demoRow(
			labeledSpinner("Small", ui.Spinner().Size(ui.SpinnerSmall)),
			labeledSpinner("Medium", ui.Spinner()),
			labeledSpinner("Large", ui.Spinner().Size(ui.SpinnerLarge).Color(ui.SpinnerAccent)),
			labeledSpinner("Danger", ui.Spinner().Color(ui.SpinnerDanger)),
		))},
	)
}

func disclosurePage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	items := []ui.CollapsibleItem{
		{Key: "profile", Label: "Profile", Leading: ui.Icon(lucide.User).Size(16), Content: ui.Description("Update your public profile and avatar.")},
		{Key: "security", Label: "Security", Leading: ui.Icon(lucide.ShieldCheck).Size(16), Content: ui.Description("Manage passwords and active sessions.")},
		{Key: "billing", Label: "Billing", Leading: ui.Icon(lucide.CreditCard).Size(16), Content: ui.Description("Review invoices and payment methods.")},
	}
	return demoPage("Disclosure",
		demoSection{Title: "Collapsible", Content: demoPanel(
			ui.Collapsible("catalog-collapsible", model.CollapsibleOpen, "What is FlowUI?",
				ui.Description("A controlled component library for Gio applications."),
			).Leading(ui.Icon(lucide.Info).Size(16)).OnExpandedChange(func(open bool) {
				send(func(model *Model) { model.CollapsibleOpen = open })
			}),
		)},
		demoSection{Title: "CollapsibleGroup", Content: demoPanel(
			ui.CollapsibleGroup("catalog-collapsible-group", model.ExpandedGroups, items).
				AllowMultipleExpanded(true).
				OnExpandedChange(func(keys []string) {
					send(func(model *Model) { model.ExpandedGroups = append([]string(nil), keys...) })
				}),
		)},
	)
}

func overlaysPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	showToast := func() {
		send(func(model *Model) {
			model.ToastSequence++
			key := fmt.Sprintf("catalog-toast-%d", model.ToastSequence)
			model.Toasts = append([]ui.ToastItem{
				ui.Toast(key, "Component saved").Description("The catalog state was updated.").Variant(ui.ToastSuccess),
			}, model.Toasts...)
		})
	}
	return demoPage("Overlays",
		demoSection{Title: "Tooltip", Content: demoPanel(demoRow(
			ui.Tooltip("catalog-tooltip-top",
				ui.Button("catalog-tooltip-top-trigger", ui.Text("Hover top")).Variant(ui.ButtonSecondary),
				ui.Text("Tooltip content"),
			).Placement(ui.TooltipTop).Arrow(true).Delay(0),
			ui.Tooltip("catalog-tooltip-right",
				ui.Button("catalog-tooltip-right-trigger", ui.Text("Hover right")).Variant(ui.ButtonSecondary),
				ui.Text("Placed on the right"),
			).Placement(ui.TooltipRight).Arrow(true).Delay(0),
		))},
		demoSection{Title: "Popover", Content: demoPanel(
			ui.Popover("catalog-popover", model.PopoverOpen,
				ui.Button("catalog-popover-trigger", ui.Text("Open popover")).Variant(ui.ButtonSecondary).OnClick(func() {
					send(func(model *Model) { model.PopoverOpen = !model.PopoverOpen })
				}),
				ui.Box(ui.Column(
					ui.Text("Quick settings").Size(15),
					ui.Switch("catalog-popover-switch", model.SwitchOn, "Notifications").OnChange(func(value bool) {
						send(func(model *Model) { model.SwitchOn = value })
					}),
				).Gap(12)).Style(ui.Width(260).Padding(14)),
			).Heading("Quick settings").Arrow(true).OnOpenChange(func(open bool) {
				send(func(model *Model) { model.PopoverOpen = open })
			}),
		)},
		demoSection{Title: "Modal, AlertDialog & Toast", Content: demoPanel(demoRow(
			ui.Button("catalog-open-modal", ui.Text("Open modal")).OnClick(func() {
				send(func(model *Model) { model.ModalOpen = true })
			}),
			ui.Button("catalog-open-alert-dialog", ui.Text("Open alert dialog")).Variant(ui.ButtonDangerSoft).OnClick(func() {
				send(func(model *Model) { model.AlertDialogOpen = true })
			}),
			ui.Button("catalog-show-toast", ui.Text("Show toast")).Variant(ui.ButtonSecondary).OnClick(showToast),
		))},
	)
}

func labeledProgressCircle(label string, circle ui.Widget) ui.Widget {
	return ui.Column(circle, ui.Text(label).Size(12)).Gap(7).AlignMiddle()
}

func labeledSpinner(label string, spinner ui.Widget) ui.Widget {
	return ui.Column(spinner, ui.Text(label).Size(12)).Gap(7).AlignMiddle()
}
