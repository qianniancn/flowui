package main

import (
	"fmt"

	ui "github.com/qianniancn/FlowUI"
)

type Model struct {
	Notifications bool
	Newsletter    bool
	Marketing     bool
	Social        bool
	PublicProfile bool
	Sound         bool
	Power         bool
	Small         bool
	Medium        bool
	Large         bool
	Last          string
}

type Field string

const (
	fieldNotifications Field = "notifications"
	fieldNewsletter    Field = "newsletter"
	fieldMarketing     Field = "marketing"
	fieldSocial        Field = "social"
	fieldPublicProfile Field = "public-profile"
	fieldSound         Field = "sound"
	fieldPower         Field = "power"
	fieldSmall         Field = "small"
	fieldMedium        Field = "medium"
	fieldLarge         Field = "large"
)

type Msg struct {
	Field   Field
	Checked bool
}

func Update(m *Model, msg Msg) {
	switch msg.Field {
	case fieldNotifications:
		m.Notifications = msg.Checked
	case fieldNewsletter:
		m.Newsletter = msg.Checked
	case fieldMarketing:
		m.Marketing = msg.Checked
	case fieldSocial:
		m.Social = msg.Checked
	case fieldPublicProfile:
		m.PublicProfile = msg.Checked
	case fieldSound:
		m.Sound = msg.Checked
	case fieldPower:
		m.Power = msg.Checked
	case fieldSmall:
		m.Small = msg.Checked
	case fieldMedium:
		m.Medium = msg.Checked
	case fieldLarge:
		m.Large = msg.Checked
	}
	m.Last = fmt.Sprintf("%s is %s", msg.Field, switchStateText(msg.Checked))
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No switch changed"
	if m.Last != "" {
		status = m.Last
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("switches",
				ui.Column(
					ui.Text("FlowUI Switch").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Account",
						ui.Column(
							switchField("notifications", fieldNotifications, m.Notifications, "Enable notifications", send).
								Description("Receive product, security, and billing updates"),
							switchField("newsletter", fieldNewsletter, m.Newsletter, "Subscribe to newsletter", send),
							switchField("public-profile", fieldPublicProfile, m.PublicProfile, "Public profile", send).
								Description("Allow others to see your public profile information"),
						).Gap(14),
					),
					section("SwitchGroup",
						ui.SwitchGroup(
							switchField("marketing", fieldMarketing, m.Marketing, "Marketing emails", send),
							switchField("social", fieldSocial, m.Social, "Social updates", send),
							switchField("sound", fieldSound, m.Sound, "Sound alerts", send),
						),
					),
					section("Horizontal group",
						ui.SwitchGroup(
							switchField("small", fieldSmall, m.Small, "Small", send).Size(ui.SwitchSmall),
							switchField("medium", fieldMedium, m.Medium, "Medium", send),
							switchField("large", fieldLarge, m.Large, "Large", send).Size(ui.SwitchLarge),
						).Horizontal(),
					),
					section("States",
						ui.Column(
							switchField("power", fieldPower, m.Power, "Power", send).
								Thumb(powerThumb).
								LabelBefore(),
							switchField("required", fieldNotifications, m.Notifications, "Required notifications", send).
								Invalid(!m.Notifications).
								Description("Turn this on before continuing"),
							ui.Switch("disabled-on", true, "Disabled selected").Disabled(true),
							ui.Switch("disabled-off", false, "Disabled").Disabled(true),
							ui.Row(
								switchField("control-only", fieldPower, m.Power, "", send).
									Thumb(powerThumb),
								ui.Text("Control only").Size(14),
							).Gap(12).AlignMiddle(),
						).Gap(14),
					),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func switchField(key string, field Field, checked bool, label string, send ui.Send[Msg]) ui.SwitchWidget {
	return ui.Switch(key, checked, label).
		OnChange(func(checked bool) {
			send(Msg{
				Field:   field,
				Checked: checked,
			})
		})
}

func powerThumb(checked bool) ui.Widget {
	if checked {
		return ui.Text("I").Size(10)
	}
	return ui.Text("O").Size(10)
}

func switchStateText(checked bool) string {
	if checked {
		return "on"
	}
	return "off"
}

func main() {
	ui.Run(Model{
		Notifications: true,
		Newsletter:    true,
		Marketing:     true,
		Medium:        true,
		Power:         true,
	}, Update, View,
		ui.Title("FlowUI Switch"),
		ui.Size(900, 720),
	)
}
