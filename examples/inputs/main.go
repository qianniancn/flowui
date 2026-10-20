package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.Kind {
	case submitted:
		m.Last = fmt.Sprintf("Submitted %s: %s", msg.Field, msg.Text)
		return nil
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
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No value submitted"
	if m.Last != "" {
		status = m.Last
	}
	emailInvalid := m.Email != "" && !containsAt(m.Email)

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Input").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				layout.Wrap(
					layout.Column(
						section("Variants",
							layout.Column(
								labeledInput("Primary", "primary",
									textField("primary", fieldName, m.Name, "Primary input", send)),
								labeledInput("Secondary", "secondary",
									textField("secondary", fieldSearch, m.Search, "Secondary input", send).
										Variant(input.Secondary)),
							).Gap(12),
						),
						section("Input types",
							layout.Column(
								labeledInput("Email", "email",
									textField("email", fieldEmail, m.Email, "jane@example.com", send).
										Type(input.Email).
										Invalid(emailInvalid)),
								labeledInput("Age", "age",
									textField("age", fieldAge, m.Age, "30", send).
										Type(input.Number).
										MaxLength(3)),
								labeledInput("Password", "password",
									textField("password", fieldPass, m.Password, "Enter password", send).
										Type(input.Password).
										MaxLength(32)),
							).Gap(12),
						),
					).Gap(18),
					layout.Column(
						section("States",
							layout.Column(
								labeledInput("Default value", "default-value",
									input.New("").DefaultValue("Initial value")),
								labeledInput("Read only", "read-only",
									input.New("Read-only value").ReadOnly(true)),
								labeledInput("Disabled", "disabled",
									input.New("Disabled value").Disabled(true)),
								labeledInput("Invalid", "invalid",
									input.New("Invalid value").Invalid(true)),
							).Gap(12),
						),
						section("Full width",
							layout.Box(
								textField("full-width", fieldFull, m.Full, "Full width input", send).
									Variant(input.Secondary).
									FullWidth(),
							).Style(style.New().Width(320)),
						),
					).Gap(18),
				).Gap(32).LineGap(24).AlignStart(),
			).Gap(18),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(10)
}

func labeledInput(title, key string, field input.Widget) ui.Widget {
	return layout.Column(
		label.New(title).For(key),
		layout.Box(field).Style(style.New().Width(320)),
	).Gap(6)
}

func textField(key string, field Field, value string, hint string, send ui.Send[Msg]) input.Widget {
	return input.New(value).Key(key).
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

func containsAt(value string) bool {
	for _, r := range value {
		if r == '@' {
			return true
		}
	}
	return false
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Password: "flowui",
	},
		Update, View), window.Title("FlowUI Inputs"),
		window.Size(900, 760),
		window.Theme(base.Light()),
	)
}
