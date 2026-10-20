package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/combobox"
	descriptionpkg "github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/input"
	labelpkg "github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/selectfield"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
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
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Description").Size(24),
				layout.Divider(),
				section("States",
					layout.Wrap(
						layout.Box(descriptionpkg.New("Muted supporting text for a field.")).Style(style.New().Width(300)),
						layout.Box(descriptionpkg.New("Disabled supporting text.").Disabled(true)).Style(style.New().Width(300)),
						layout.Box(descriptionpkg.New("Long descriptions wrap naturally when the available field width becomes constrained.")).Style(style.New().Width(300)),
					).Gap(16).LineGap(12),
				),
				section("Field association",
					layout.Wrap(
						field(
							labelpkg.New("Email").For("email"),
							input.New(model.Email).Key("email").
								Hint("name@example.com").
								OnChange(func(value string) { send(SetEmail(value)) }).
								FullWidth(),
							descriptionpkg.New("Used for account notifications.").For("email"),
						),
						field(
							labelpkg.New("Language").For("language"),
							combobox.New(languages()).Key("language").SelectedKey(model.Language).
								Hint("Choose a language").
								OnChange(func(key string) { send(SetLanguage(key)) }).
								FullWidth(),
							descriptionpkg.New("Search or choose from the available languages.").For("language"),
						),
						field(
							labelpkg.New("State").For("state"),
							selectfield.New(states()).Key("state").SelectedKey(model.State).
								Placeholder("Select a state").
								OnChange(func(key string) { send(SetState(key)) }).
								FullWidth(),
							descriptionpkg.New("Used to determine regional availability.").For("state"),
						),
					).Gap(16).LineGap(16),
				),
				section("Component compatibility",
					layout.Wrap(
						layout.Box(
							selectfield.New(states()).Key("select-description").SelectedKey(model.State).
								Label("Select description").
								Description("Rendered by the shared Description component.").
								OnChange(func(key string) { send(SetState(key)) }).
								FullWidth(),
						).Style(style.New().Width(300)),
						layout.Box(
							switchcontrol.New("Marketing updates").Key("marketing").Checked(model.Marketing).
								Description("Receive occasional product announcements.").
								OnChange(func(value bool) { send(SetMarketing(value)) }),
						).Style(style.New().Width(300)),
					).Gap(16).LineGap(16),
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

func field(label, control, description ui.Widget) ui.Widget {
	return layout.Box(layout.Column(label, control, description).Gap(6)).Style(style.New().Width(300))
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
		Update, View), window.Title("FlowUI Descriptions"),
		window.Size(900, 680),
		window.Theme(base.Light()),
	)
}
