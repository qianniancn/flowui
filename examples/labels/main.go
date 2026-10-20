package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/input"
	labelpkg "github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/selectfield"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Email    string
	Language string
	State    string
}

type Msg any

type SetEmail string
type SetLanguage string
type SetState string

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case SetEmail:
		model.Email = string(msg)
	case SetLanguage:
		model.Language = string(msg)
	case SetState:
		model.State = string(msg)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Label").Size(24),
				layout.Divider(),
				section("States",
					layout.Row(
						labelpkg.New("Default"),
						labelpkg.New("Required").Required(true),
						labelpkg.New("Invalid").Invalid(true),
						labelpkg.New("Disabled").Disabled(true),
					).Gap(24).AlignMiddle(),
				),
				section("Field association",
					layout.Wrap(
						field(
							labelpkg.New("Email").For("email").Required(true),
							input.New(model.Email).Key("email").
								Hint("name@example.com").
								OnChange(func(value string) { send(SetEmail(value)) }).
								FullWidth(),
						),
						field(
							labelpkg.New("Language").For("language"),
							combobox.New(languages()).Key("language").SelectedKey(model.Language).
								Hint("Choose a language").
								OnChange(func(key string) { send(SetLanguage(key)) }).
								FullWidth(),
						),
						field(
							labelpkg.New("State").For("state").Invalid(model.State == ""),
							selectfield.New(states()).Key("state").SelectedKey(model.State).
								Placeholder("Select a state").
								Invalid(model.State == "").
								OnChange(func(key string) { send(SetState(key)) }).
								FullWidth(),
						),
					).Gap(16).LineGap(16),
				),
				section("Select compatibility",
					layout.Box(
						selectfield.New(states()).Key("built-in-label").SelectedKey(model.State).
							Label("Built-in Select label").
							Required(true).
							OnChange(func(key string) { send(SetState(key)) }).
							FullWidth(),
					).Style(style.New().Width(300)),
				),
			).Gap(20),
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

func field(label, control ui.Widget) ui.Widget {
	return layout.Box(layout.Column(label, control).Gap(6)).Style(style.New().Width(300))
}

func languages() []combobox.Item {
	return []combobox.Item{
		{Key: "go", Label: "Go"},
		{Key: "rust", Label: "Rust"},
		{Key: "typescript", Label: "TypeScript"},
	}
}

func states() []selectfield.SelectItem {
	return []selectfield.SelectItem{
		{Key: "california", Label: "California"},
		{Key: "new-york", Label: "New York"},
		{Key: "texas", Label: "Texas"},
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Labels"),
		window.Size(900, 640),
		window.Theme(base.Light()),
	)
}
