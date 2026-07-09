package main

import ui "github.com/qianniancn/FlowUI"

type Model struct {
	Terms     bool
	Email     bool
	Reports   bool
	TwoFactor bool
	Archived  bool
	Invalid   bool
	Compact   bool
}

type Field string

const (
	fieldTerms     Field = "terms"
	fieldEmail     Field = "email"
	fieldReports   Field = "reports"
	fieldTwoFactor Field = "two-factor"
	fieldArchived  Field = "archived"
	fieldInvalid   Field = "invalid"
	fieldCompact   Field = "compact"
)

type Msg struct {
	Field   Field
	Checked bool
}

func Update(m *Model, msg Msg) {
	switch msg.Field {
	case fieldTerms:
		m.Terms = msg.Checked
	case fieldEmail:
		m.Email = msg.Checked
	case fieldReports:
		m.Reports = msg.Checked
	case fieldTwoFactor:
		m.TwoFactor = msg.Checked
	case fieldArchived:
		m.Archived = msg.Checked
	case fieldInvalid:
		m.Invalid = msg.Checked
	case fieldCompact:
		m.Compact = msg.Checked
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("checkboxes",
				ui.Column(
					ui.Text("FlowUI Checkboxes").Size(24),
					ui.Text("Account").Size(18),
					ui.Column(
						checkbox("terms", fieldTerms, m.Terms, "Accept terms", send).
							Invalid(!m.Terms),
						checkbox("email", fieldEmail, m.Email, "Email updates", send),
						checkbox("reports", fieldReports, m.Reports, "Weekly reports", send),
						checkbox("two-factor", fieldTwoFactor, m.TwoFactor, "Two-factor prompts", send),
					).Gap(14),
					ui.Divider(),
					ui.Text("States").Size(18),
					ui.Column(
						checkbox("archived", fieldArchived, m.Archived, "Archive item", send),
						checkbox("invalid-selected", fieldInvalid, m.Invalid, "Always invalid", send).
							Invalid(true),
						ui.Checkbox("checked-disabled", true, "Disabled checked").Disabled(true),
						ui.Checkbox("disabled", false, "Disabled").Disabled(true),
						ui.Row(
							checkbox("compact", fieldCompact, m.Compact, "", send),
							ui.Text("Control only").Size(14),
						).Gap(12).AlignMiddle(),
					).Gap(14),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(640).Padding(24),
	)
}

func checkbox(key string, field Field, checked bool, label string, send ui.Send[Msg]) ui.CheckboxWidget {
	return ui.Checkbox(key, checked, label).
		OnChange(func(checked bool) {
			send(Msg{
				Field:   field,
				Checked: checked,
			})
		})
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Checkboxes"),
		ui.Size(900, 640),
	)
}
