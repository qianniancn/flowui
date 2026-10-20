package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alert"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/collapsible"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/popover"
	"github.com/qianniancn/flowui/widget/progress"
	spinnerpkg "github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tooltip"
)

func statusPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Status & progress",
		demoSection{Title: "Alert", Content: layout.Column(
			alert.New("Information", "A neutral message for the current task."),
			alert.New("Update complete", "All component examples are available.").Status(alert.Success),
			alert.New("Review required", "Check the generated output before publishing.").Status(alert.Warning),
			alert.New("Build failed", "Resolve the reported errors and try again.").Status(alert.Danger),
		).Gap(10)},
		demoSection{Title: "ProgressBar", Content: demoPanel(layout.Column(
			progress.Bar("catalog-progress-default", 42).Label("Indexing").ShowValue(),
			progress.Bar("catalog-progress-success", 76).Label("Upload").ShowValue().Color(progress.BarSuccess),
			progress.Bar("catalog-progress-indeterminate", 0).Label("Syncing").Indeterminate(),
		).Gap(14))},
		demoSection{Title: "ProgressCircle", Content: demoPanel(demoRow(
			labeledProgressCircle("Small", progress.Circle("catalog-circle-small", 35).Size(progress.CircleSmall)),
			labeledProgressCircle("Medium", progress.Circle("catalog-circle-medium", 60)),
			labeledProgressCircle("Large", progress.Circle("catalog-circle-large", 82).Size(progress.CircleLarge).Color(progress.CircleSuccess)),
			labeledProgressCircle("Loading", progress.Circle("catalog-circle-loading", 0).Indeterminate()),
		))},
		demoSection{Title: "Meter", Content: demoPanel(layout.Column(
			progress.Meter("catalog-meter-storage", 68).Label("Storage").ShowValue(),
			progress.Meter("catalog-meter-memory", 84).Label("Memory").ShowValue().Color(progress.MeterWarning),
			progress.Meter("catalog-meter-budget", 750).Label("Budget").Range(0, 1000).
				ValueFormatter(func(value float64) string { return fmt.Sprintf("$%.0f", value) }).Color(progress.MeterSuccess),
		).Gap(14))},
		demoSection{Title: "Spinner", Content: demoPanel(demoRow(
			labeledSpinner("Small", spinnerpkg.New().Size(spinnerpkg.Small)),
			labeledSpinner("Medium", spinnerpkg.New()),
			labeledSpinner("Large", spinnerpkg.New().Size(spinnerpkg.Large).Color(spinnerpkg.Accent)),
			labeledSpinner("Danger", spinnerpkg.New().Color(spinnerpkg.Danger)),
		))},
	)
}

func disclosurePage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	items := []collapsible.Item{
		{Key: "profile", Label: "Profile", Leading: icon.New(lucide.User).Size(16), Content: description.New("Update your public profile and avatar.")},
		{Key: "security", Label: "Security", Leading: icon.New(lucide.ShieldCheck).Size(16), Content: description.New("Manage passwords and active sessions.")},
		{Key: "billing", Label: "Billing", Leading: icon.New(lucide.CreditCard).Size(16), Content: description.New("Review invoices and payment methods.")},
	}
	return demoPage("Disclosure",
		demoSection{Title: "Collapsible", Content: demoPanel(
			collapsible.New("What is FlowUI?", description.New("A controlled component library for Gio applications.")).Key("catalog-collapsible").Expanded(model.CollapsibleOpen).Leading(icon.New(lucide.Info).Size(16)).OnExpandedChange(func(open bool) {
				send(func(model *Model) { model.CollapsibleOpen = open })
			}),
		)},
		demoSection{Title: "CollapsibleGroup", Content: demoPanel(
			collapsible.Group("catalog-collapsible-group", model.ExpandedGroups, items).
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
			model.Toasts = append([]toast.Item{
				toast.NewItem(key, "Component saved").Description("The catalog state was updated.").Variant(toast.Success),
			}, model.Toasts...)
		})
	}
	return demoPage("Overlays",
		demoSection{Title: "Tooltip", Content: demoPanel(demoRow(
			tooltip.New("catalog-tooltip-top", button.New("Hover top").Variant(button.Secondary), "Tooltip content").Placement(tooltip.Top).Arrow(true).Delay(0),
			tooltip.New("catalog-tooltip-right", button.New("Hover right").Variant(button.Secondary), "Placed on the right").Placement(tooltip.Right).Arrow(true).Delay(0),
		))},
		demoSection{Title: "Popover", Content: demoPanel(
			popover.New(button.New("Open popover").Variant(button.Secondary).OnClick(func() {
				send(func(model *Model) { model.PopoverOpen = !model.PopoverOpen })
			}), layout.Box(layout.Column(
				text.New("Quick settings").Size(15),
				switchcontrol.New("Notifications").Key("catalog-popover-switch").Checked(model.SwitchOn).OnChange(func(value bool) {
					send(func(model *Model) { model.SwitchOn = value })
				}),
			).Gap(12)).Style(style.New().Width(260).Padding(14))).Key("catalog-popover").Open(model.PopoverOpen).Heading("Quick settings").Arrow(true).OnOpenChange(func(open bool) {
				send(func(model *Model) { model.PopoverOpen = open })
			}),
		)},
		demoSection{Title: "Modal, AlertDialog & Toast", Content: demoPanel(demoRow(
			button.New("Open modal").OnClick(func() {
				send(func(model *Model) { model.ModalOpen = true })
			}),
			button.New("Open alert dialog").Variant(button.DangerSoft).OnClick(func() {
				send(func(model *Model) { model.AlertDialogOpen = true })
			}),
			button.New("Show toast").Variant(button.Secondary).OnClick(showToast),
		))},
	)
}

func labeledProgressCircle(label string, circle ui.Widget) ui.Widget {
	return layout.Column(circle, text.New(label).Size(12)).Gap(7).AlignMiddle()
}

func labeledSpinner(label string, spinner ui.Widget) ui.Widget {
	return layout.Column(spinner, text.New(label).Size(12)).Gap(7).AlignMiddle()
}
