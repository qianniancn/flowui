package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct {
	Email     string
	Language  string
	State     string
	Marketing bool
}

type Msg any

type SetEmail string
type SetLanguage string
type SetState string
type SetMarketing bool

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case SetEmail:
		model.Email = string(msg)
	case SetLanguage:
		model.Language = string(msg)
	case SetState:
		model.State = string(msg)
	case SetMarketing:
		model.Marketing = bool(msg)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("descriptions",
				ui.Column(
					ui.Text("FlowUI Description").Size(24),
					ui.Divider(),
					section("States",
						ui.Wrap(
							ui.Box(ui.Description("Muted supporting text for a field.")).Width(300),
							ui.Box(ui.Description("Disabled supporting text.").Disabled(true)).Width(300),
							ui.Box(ui.Description("Long descriptions wrap naturally when the available field width becomes constrained.")).Width(300),
						).Gap(16).LineGap(12),
					),
					section("Field association",
						ui.Wrap(
							field(
								ui.Label("Email").For("email"),
								ui.Input("email", model.Email).
									Hint("name@example.com").
									OnChange(func(value string) { send(SetEmail(value)) }).
									FullWidth(),
								ui.Description("Used for account notifications.").For("email"),
							),
							field(
								ui.Label("Language").For("language"),
								ui.ComboBox("language", model.Language, languages()).
									Hint("Choose a language").
									OnChange(func(key string) { send(SetLanguage(key)) }).
									FullWidth(),
								ui.Description("Search or choose from the available languages.").For("language"),
							),
							field(
								ui.Label("State").For("state"),
								ui.Select("state", model.State, states()).
									Placeholder("Select a state").
									OnChange(func(key string) { send(SetState(key)) }).
									FullWidth(),
								ui.Description("Used to determine regional availability.").For("state"),
							),
						).Gap(16).LineGap(16),
					),
					section("Component compatibility",
						ui.Wrap(
							ui.Box(
								ui.Select("select-description", model.State, states()).
									Label("Select description").
									Description("Rendered by the shared Description component.").
									OnChange(func(key string) { send(SetState(key)) }).
									FullWidth(),
							).Width(300),
							ui.Box(
								ui.Switch("marketing", model.Marketing, "Marketing updates").
									Description("Receive occasional product announcements.").
									OnChange(func(value bool) { send(SetMarketing(value)) }),
							).Width(300),
						).Gap(16).LineGap(16),
					),
				).Gap(20),
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

func field(label, control, description ui.Widget) ui.Widget {
	return ui.Box(ui.Column(label, control, description).Gap(6)).Width(300)
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
		ui.Title("FlowUI Descriptions"),
		ui.Size(900, 680),
	)
}
