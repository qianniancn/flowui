package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alert"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/closebutton"
	"github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Last        string
	ShowProfile bool
}

type Msg struct {
	Action string
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	m.Last = msg.Action
	if msg.Action == "Dismissed profile update" {
		m.ShowProfile = false
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No action selected"
	if m.Last != "" {
		status = m.Last
	}

	alerts := []ui.Widget{
		alert.New(
			"New features available",
			"Check out the latest updates including dark mode support and improved accessibility features.",
		),
		alert.New(
			"Update available",
			"A new application version is available. Refresh to get the latest fixes.",
		).
			Status(alert.Accent).
			Action(alertButton("refresh", "Refresh", button.Primary, send)),
		alert.New(
			"Payment successful",
			"Your payment has been processed and a confirmation email was sent.",
		).
			Status(alert.Success).
			Action(alertButton("receipt", "View receipt", button.Secondary, send)),
		alert.New(
			"Storage almost full",
			"You are using 90% of your storage quota. Remove unused files to avoid interruption.",
		).
			Status(alert.Warning).
			Action(alertButton("storage", "Manage", button.Secondary, send)),
		alert.New("Unable to connect to server", "").
			Status(alert.Danger).
			Content(
				layout.Column(
					text.New("Check your internet connection, then retry the request."),
					text.New("The service status page may contain more details.").Size(13),
				).Gap(4),
			).
			Action(alertButton("retry", "Retry", button.Danger, send)),
		alert.New(
			"Processing your request",
			"Please wait while we sync your data. This may take a few moments.",
		).
			Status(alert.Accent).
			Indicator(spinner.New().Color(spinner.Current).Size(spinner.Small)),
	}
	if m.ShowProfile {
		alerts = append(alerts,
			alert.New("Profile updated successfully", "").
				Status(alert.Success).
				Action(
					closebutton.New("dismiss-profile").
						Label("Dismiss profile update").
						OnClick(func() { send(Msg{Action: "Dismissed profile update"}) }),
				),
		)
	}

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Box(
				layout.Column(
					text.New("FlowUI Alert").Size(24),
					text.New(status).Size(14),
					layout.Divider(),
					layout.Column(alerts...).Gap(16),
				).Gap(20),
			).Style(style.New().Padding(3)),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(680).Padding(24)),
	)
}

func alertButton(key, label string, variant button.Variant, send ui.Send[Msg]) button.Widget {
	return button.New(label).Key(key).
		Variant(variant).
		Size(button.Small).
		OnClick(func() { send(Msg{Action: label}) })
}

func main() {
	ui.Run(ui.NewProgram(Model{ShowProfile: true},
		Update, View), window.Title("FlowUI Alerts"),
		window.Size(860, 820),
		window.Theme(base.Light()),
	)
}
