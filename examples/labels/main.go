package main

import "github.com/qianniancn/flowui/ui"

type Model struct {
	Email    string
	Language string
	State    string
}

type Msg any

type SetEmail string
type SetLanguage string
type SetState string

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case SetEmail:
		model.Email = string(msg)
	case SetLanguage:
		model.Language = string(msg)
	case SetState:
		model.State = string(msg)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("labels",
				ui.Column(
					ui.Text("FlowUI Label").Size(24),
					ui.Divider(),
					section("States",
						ui.Row(
							ui.Label("Default"),
							ui.Label("Required").Required(true),
							ui.Label("Invalid").Invalid(true),
							ui.Label("Disabled").Disabled(true),
						).Gap(24).AlignMiddle(),
					),
					section("Field association",
						ui.Wrap(
							field(
								ui.Label("Email").For("email").Required(true),
								ui.Input("email", model.Email).
									Hint("name@example.com").
									OnChange(func(value string) { send(SetEmail(value)) }).
									FullWidth(),
							),
							field(
								ui.Label("Language").For("language"),
								ui.ComboBox("language", model.Language, languages()).
									Hint("Choose a language").
									OnChange(func(key string) { send(SetLanguage(key)) }).
									FullWidth(),
							),
							field(
								ui.Label("State").For("state").Invalid(model.State == ""),
								ui.Select("state", model.State, states()).
									Placeholder("Select a state").
									Invalid(model.State == "").
									OnChange(func(key string) { send(SetState(key)) }).
									FullWidth(),
							),
						).Gap(16).LineGap(16),
					),
					section("Select compatibility",
						ui.Box(
							ui.Select("built-in-label", model.State, states()).
								Label("Built-in Select label").
								Required(true).
								OnChange(func(key string) { send(SetState(key)) }).
								FullWidth(),
						).Style(ui.Width(300)),
					),
				).Gap(20),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func field(label, control ui.Widget) ui.Widget {
	return ui.Box(ui.Column(label, control).Gap(6)).Style(ui.Width(300))
}

func languages() []ui.ComboBoxItem {
	return []ui.ComboBoxItem{
		{Key: "go", Label: "Go"},
		{Key: "rust", Label: "Rust"},
		{Key: "typescript", Label: "TypeScript"},
	}
}

func states() []ui.SelectItem {
	return []ui.SelectItem{
		{Key: "california", Label: "California"},
		{Key: "new-york", Label: "New York"},
		{Key: "texas", Label: "Texas"},
	}
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Labels"),
		ui.Size(900, 640),
	)
}
