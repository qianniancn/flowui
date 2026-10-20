package main

import (
	"strings"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/selectfield"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	State          string
	Countries      []string
	Framework      string
	SurfaceChoice  string
	RequiredChoice string
	Controlled     string
	ControlledOpen bool
}

type Msg any

type SetState string
type SetCountries []string
type SetFramework string
type SetSurfaceChoice string
type SetRequiredChoice string
type SetControlled string
type SetControlledOpen bool

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case SetState:
		model.State = string(msg)
	case SetCountries:
		model.Countries = append([]string(nil), msg...)
	case SetFramework:
		model.Framework = string(msg)
	case SetSurfaceChoice:
		model.SurfaceChoice = string(msg)
	case SetRequiredChoice:
		model.RequiredChoice = string(msg)
	case SetControlled:
		model.Controlled = string(msg)
	case SetControlledOpen:
		model.ControlledOpen = bool(msg)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Select").Size(24),
				text.New("Field, selection, overlay, and keyboard behavior."),
				layout.Divider(),
				section("Basic",
					layout.Wrap(
						selectWidth(
							selectfield.New(stateItems()).Key("state").SelectedKey(model.State).
								Label("State").
								Placeholder("Select one").
								Description("Choose a shipping destination.").
								OnChange(func(key string) { send(SetState(key)) }),
						),
						selectWidth(
							selectfield.New(countryItems()).Key("countries").SelectedKeys(model.Countries).
								Label("Countries to visit").
								Placeholder("Select countries").
								OnSelectionChange(func(keys []string) { send(SetCountries(keys)) }),
						),
					).Gap(16).LineGap(16),
				),
				section("Sections and disabled options",
					selectWidth(
						selectfield.NewSections(frameworkSections()).Key("framework").SelectedKey(model.Framework).
							Label("Framework").
							Placeholder("Choose a framework").
							Indicator(text.New("v").Size(12)).
							DisabledKeys([]string{"angular"}).
							OnChange(func(key string) { send(SetFramework(key)) }),
					),
				),
				section("Required and controlled",
					layout.Wrap(
						selectWidth(
							selectfield.New(stateItems()).Key("required").SelectedKey(model.RequiredChoice).
								Label("Required state").
								Placeholder("Select one").
								Required(true).
								Invalid(model.RequiredChoice == "").
								ErrorMessage("Select a state to continue.").
								OnChange(func(key string) { send(SetRequiredChoice(key)) }),
						),
						selectWidth(
							selectfield.New(stateItems()).Key("controlled").SelectedKey(model.Controlled).
								Label("Controlled open state").
								Placeholder("Select one").
								Open(model.ControlledOpen).
								OnOpenChange(func(open bool) { send(SetControlledOpen(open)) }).
								OnChange(func(key string) { send(SetControlled(key)) }),
						),
					).Gap(16).LineGap(16),
				),
				section("On Surface",
					(surface.New(
						layout.Box(
							selectfield.New(stateItems()).Key("surface-choice").SelectedKey(model.SurfaceChoice).
								Label("Secondary variant").
								Placeholder("Select one").
								Variant(selectfield.SelectSecondary).
								ValueText(customValue(model.SurfaceChoice)).
								OnChange(func(key string) { send(SetSurfaceChoice(key)) }).
								FullWidth(),
						).Style(style.New().Width(300).Padding(20)),
					).Variant(surface.Default)).Style(style.New().Radius(24).Shadow(style.ShadowSurface)),
				),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(10)
}

func selectWidth(selectWidget selectfield.Widget) ui.Widget {
	return layout.Box(selectWidget.FullWidth()).Style(style.New().Width(300))
}

func customValue(key string) string {
	if key == "" {
		return ""
	}
	return "Selected: " + strings.ToUpper(key)
}

func stateItems() []selectfield.SelectItem {
	return []selectfield.SelectItem{
		{Key: "california", Label: "California"},
		{Key: "florida", Label: "Florida"},
		{Key: "new-york", Label: "New York"},
		{Key: "texas", Label: "Texas"},
		{Key: "washington", Label: "Washington"},
	}
}

func countryItems() []selectfield.SelectItem {
	names := []string{"Argentina", "France", "Iceland", "Italy", "Japan", "New Zealand", "Spain", "Thailand"}
	items := make([]selectfield.SelectItem, 0, len(names))
	for _, name := range names {
		items = append(items, selectfield.SelectItem{Key: strings.ToLower(strings.ReplaceAll(name, " ", "-")), Label: name})
	}
	return items
}

func frameworkSections() []selectfield.SelectSection {
	return []selectfield.SelectSection{
		{Title: "Frontend", Items: []selectfield.SelectItem{
			{Key: "react", Label: "React", Description: "Component-based UI"},
			{Key: "vue", Label: "Vue", Description: "Progressive framework"},
			{Key: "angular", Label: "Angular", Description: "Disabled option"},
		}},
		{Title: "Backend", Items: []selectfield.SelectItem{
			{Key: "go", Label: "Go", Description: "Simple and concurrent"},
			{Key: "rust", Label: "Rust", Description: "Fast and memory safe"},
		}},
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Select"),
		window.Size(920, 760),
		window.Theme(base.Light()),
	)
}
