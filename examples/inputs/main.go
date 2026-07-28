package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Name     string
	Email    string
	Age      string
	Password string
	Search   string
	Full     string
	Last     string
}

type Field string

const (
	fieldName   Field = "name"
	fieldEmail  Field = "email"
	fieldAge    Field = "age"
	fieldPass   Field = "password"
	fieldSearch Field = "search"
	fieldFull   Field = "full"
)

type Kind int

const (
	changed Kind = iota
	submitted
)

type Msg struct {
	Kind  Kind
	Field Field
	Text  string
}

func Update(m *Model, msg Msg) {
	if msg.Kind == submitted {
		m.Last = fmt.Sprintf("Submitted %s: %s", msg.Field, msg.Text)
		return
	}
	switch msg.Field {
	case fieldName:
		m.Name = msg.Text
	case fieldEmail:
		m.Email = msg.Text
	case fieldAge:
		m.Age = msg.Text
	case fieldPass:
		m.Password = msg.Text
	case fieldSearch:
		m.Search = msg.Text
	case fieldFull:
		m.Full = msg.Text
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No value submitted"
	if m.Last != "" {
		status = m.Last
	}
	emailInvalid := m.Email != "" && !containsAt(m.Email)

	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Input").Size(24),
				ui.Text(status).Size(16),
				ui.Divider(),
				ui.Wrap(
					ui.Column(
						section("Variants",
							ui.Column(
								labeledInput("Primary", "primary",
									input("primary", fieldName, m.Name, "Primary input", send)),
								labeledInput("Secondary", "secondary",
									input("secondary", fieldSearch, m.Search, "Secondary input", send).
										Variant(ui.InputSecondary)),
							).Gap(12),
						),
						section("Input types",
							ui.Column(
								labeledInput("Email", "email",
									input("email", fieldEmail, m.Email, "jane@example.com", send).
										Type(ui.InputEmail).
										Invalid(emailInvalid)),
								labeledInput("Age", "age",
									input("age", fieldAge, m.Age, "30", send).
										Type(ui.InputNumber).
										MaxLength(3)),
								labeledInput("Password", "password",
									input("password", fieldPass, m.Password, "Enter password", send).
										Type(ui.InputPassword).
										MaxLength(32)),
							).Gap(12),
						),
					).Gap(18),
					ui.Column(
						section("States",
							ui.Column(
								labeledInput("Read only", "read-only",
									ui.Input("read-only", "Read-only value").ReadOnly(true)),
								labeledInput("Disabled", "disabled",
									ui.Input("disabled", "Disabled value").Disabled(true)),
								labeledInput("Invalid", "invalid",
									ui.Input("invalid", "Invalid value").Invalid(true)),
							).Gap(12),
						),
						section("Full width",
							ui.Box(
								input("full-width", fieldFull, m.Full, "Full width input", send).
									Variant(ui.InputSecondary).
									FullWidth(),
							).Style(ui.Width(320)),
						),
					).Gap(18),
				).Gap(32).LineGap(24).AlignStart(),
			).Gap(18),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(720)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func labeledInput(label, key string, field ui.InputWidget) ui.Widget {
	return ui.Column(
		ui.Label(label).For(key),
		ui.Box(field).Style(ui.Width(320)),
	).Gap(6)
}

func input(key string, field Field, value string, hint string, send ui.Send[Msg]) ui.InputWidget {
	return ui.Input(key, value).
		Placeholder(hint).
		OnChange(func(text string) {
			send(Msg{
				Kind:  changed,
				Field: field,
				Text:  text,
			})
		}).
		OnSubmit(func(text string) {
			send(Msg{
				Kind:  submitted,
				Field: field,
				Text:  text,
			})
		})
}

func containsAt(text string) bool {
	for _, r := range text {
		if r == '@' {
			return true
		}
	}
	return false
}

func main() {
	ui.Run(Model{Password: "heroui"}, Update, View,
		ui.Title("FlowUI Inputs"),
		ui.Size(900, 760),
	)
}
