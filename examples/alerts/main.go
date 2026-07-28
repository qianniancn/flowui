package main

import "github.com/qianniancn/flowui/ui"

type Model struct {
	Last        string
	ShowProfile bool
}

type Msg struct {
	Action string
}

func Update(m *Model, msg Msg) {
	m.Last = msg.Action
	if msg.Action == "Dismissed profile update" {
		m.ShowProfile = false
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No action selected"
	if m.Last != "" {
		status = m.Last
	}

	alerts := []ui.Widget{
		ui.Alert(
			"New features available",
			"Check out the latest updates including dark mode support and improved accessibility features.",
		),
		ui.Alert(
			"Update available",
			"A new application version is available. Refresh to get the latest fixes.",
		).
			Status(ui.AlertAccent).
			Action(alertButton("refresh", "Refresh", ui.ButtonPrimary, send)),
		ui.Alert(
			"Payment successful",
			"Your payment has been processed and a confirmation email was sent.",
		).
			Status(ui.AlertSuccess).
			Action(alertButton("receipt", "View receipt", ui.ButtonSecondary, send)),
		ui.Alert(
			"Storage almost full",
			"You are using 90% of your storage quota. Remove unused files to avoid interruption.",
		).
			Status(ui.AlertWarning).
			Action(alertButton("storage", "Manage", ui.ButtonSecondary, send)),
		ui.Alert("Unable to connect to server", "").
			Status(ui.AlertDanger).
			Content(
				ui.Column(
					ui.Text("Check your internet connection, then retry the request."),
					ui.Text("The service status page may contain more details.").Size(13),
				).Gap(4),
			).
			Action(alertButton("retry", "Retry", ui.ButtonDanger, send)),
		ui.Alert(
			"Processing your request",
			"Please wait while we sync your data. This may take a few moments.",
		).
			Status(ui.AlertAccent).
			Indicator(ui.Spinner().Color(ui.SpinnerCurrent).Size(ui.SpinnerSmall)),
	}
	if m.ShowProfile {
		alerts = append(alerts,
			ui.Alert("Profile updated successfully", "").
				Status(ui.AlertSuccess).
				Action(
					ui.CloseButton("dismiss-profile").
						Label("Dismiss profile update").
						OnClick(func() { send(Msg{Action: "Dismissed profile update"}) }),
				),
		)
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("alerts",
				ui.Box(
					ui.Column(
						ui.Text("FlowUI Alert").Size(24),
						ui.Text(status).Size(14),
						ui.Divider(),
						ui.Column(alerts...).Gap(16),
					).Gap(20),
				).Style(ui.Padding(3)),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(680)).Style(ui.Padding(24)),
	)
}

func alertButton(key, label string, variant ui.ButtonVariant, send ui.Send[Msg]) ui.ButtonWidget {
	return ui.Button(key, ui.Text(label)).
		Variant(variant).
		Size(ui.ButtonSmall).
		OnClick(func() { send(Msg{Action: label}) })
}

func main() {
	ui.Run(
		Model{ShowProfile: true},
		Update,
		View,
		ui.Title("FlowUI Alerts"),
		ui.Size(860, 820),
	)
}
