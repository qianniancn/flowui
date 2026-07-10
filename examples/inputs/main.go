package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Name   string
	Email  string
	Search string
	Full   string
	Last   string
}

type Field string

const (
	fieldName   Field = "name"
	fieldEmail  Field = "email"
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
	case fieldSearch:
		m.Search = msg.Text
	case fieldFull:
		m.Full = msg.Text
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "Press Enter in a field to submit."
	if m.Last != "" {
		status = m.Last
	}
	emailInvalid := m.Email != "" && !containsAt(m.Email)

	return ui.Center(
		ui.Box(
			ui.Scroll("inputs",
				ui.Column(
					ui.Text("FlowUI Inputs").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Variants",
						ui.Column(
							ui.Box(input("primary", fieldName, m.Name, "Primary input", send)).
								Width(320),
							ui.Box(input("secondary", fieldSearch, m.Search, "Secondary input", send).
								Variant(ui.InputSecondary)).
								Width(320),
						).Gap(12),
					),
					section("States",
						ui.Column(
							ui.Box(input("email", fieldEmail, m.Email, "Email", send).
								Invalid(emailInvalid)).
								Width(320),
							ui.Box(ui.Input("disabled", "Disabled value").
								Disabled(true)).
								Width(320),
						).Gap(12),
					),
					section("Full width",
						input("full-width", fieldFull, m.Full, "Full width input", send).
							Variant(ui.InputSecondary).
							FullWidth(),
					),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(720).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func input(key string, field Field, value string, hint string, send ui.Send[Msg]) ui.InputWidget {
	return ui.Input(key, value).
		Hint(hint).
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
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Inputs"),
		ui.Size(900, 640),
	)
}
