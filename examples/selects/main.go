package main

import (
	"strings"

	"github.com/qianniancn/FlowUI/ui"
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

func Update(model *Model, msg Msg) {
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
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Select").Size(24),
				ui.Text("HeroUI-aligned field, selection, overlay, and keyboard behavior."),
				ui.Divider(),
				section("Basic",
					ui.Wrap(
						selectWidth(
							ui.Select("state", model.State, stateItems()).
								Label("State").
								Placeholder("Select one").
								Description("Choose a shipping destination.").
								OnChange(func(key string) { send(SetState(key)) }),
						),
						selectWidth(
							ui.SelectMultiple("countries", model.Countries, countryItems()).
								Label("Countries to visit").
								Placeholder("Select countries").
								OnSelectionChange(func(keys []string) { send(SetCountries(keys)) }),
						),
					).Gap(16).LineGap(16),
				),
				section("Sections and disabled options",
					selectWidth(
						ui.SelectSections("framework", model.Framework, frameworkSections()).
							Label("Framework").
							Placeholder("Choose a framework").
							Indicator(ui.Text("v").Size(12)).
							DisabledKeys([]string{"angular"}).
							OnChange(func(key string) { send(SetFramework(key)) }),
					),
				),
				section("Required and controlled",
					ui.Wrap(
						selectWidth(
							ui.Select("required", model.RequiredChoice, stateItems()).
								Label("Required state").
								Placeholder("Select one").
								Required(true).
								Invalid(model.RequiredChoice == "").
								ErrorMessage("Select a state to continue.").
								OnChange(func(key string) { send(SetRequiredChoice(key)) }),
						),
						selectWidth(
							ui.Select("controlled", model.Controlled, stateItems()).
								Label("Controlled open state").
								Placeholder("Select one").
								Open(model.ControlledOpen).
								OnOpenChange(func(open bool) { send(SetControlledOpen(open)) }).
								OnChange(func(key string) { send(SetControlled(key)) }),
						),
					).Gap(16).LineGap(16),
				),
				section("On Surface",
					ui.Surface(
						ui.Box(
							ui.Select("surface-choice", model.SurfaceChoice, stateItems()).
								Label("Secondary variant").
								Placeholder("Select one").
								Variant(ui.SelectSecondary).
								ValueText(customValue(model.SurfaceChoice)).
								OnChange(func(key string) { send(SetSurfaceChoice(key)) }).
								FullWidth(),
						).Width(300).Padding(20),
					).Variant(ui.SurfaceDefault).Style(ui.Radius(24).Shadow(ui.ShadowSurface)),
				),
			).Gap(18),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func selectWidth(selectWidget ui.SelectWidget) ui.Widget {
	return ui.Box(selectWidget.FullWidth()).Width(300)
}

func customValue(key string) string {
	if key == "" {
		return ""
	}
	return "Selected: " + strings.ToUpper(key)
}

func stateItems() []ui.SelectItem {
	return []ui.SelectItem{
		{Key: "california", Label: "California"},
		{Key: "florida", Label: "Florida"},
		{Key: "new-york", Label: "New York"},
		{Key: "texas", Label: "Texas"},
		{Key: "washington", Label: "Washington"},
	}
}

func countryItems() []ui.SelectItem {
	names := []string{"Argentina", "France", "Iceland", "Italy", "Japan", "New Zealand", "Spain", "Thailand"}
	items := make([]ui.SelectItem, 0, len(names))
	for _, name := range names {
		items = append(items, ui.SelectItem{Key: strings.ToLower(strings.ReplaceAll(name, " ", "-")), Label: name})
	}
	return items
}

func frameworkSections() []ui.SelectSection {
	return []ui.SelectSection{
		{Title: "Frontend", Items: []ui.SelectItem{
			{Key: "react", Label: "React", Description: "Component-based UI"},
			{Key: "vue", Label: "Vue", Description: "Progressive framework"},
			{Key: "angular", Label: "Angular", Description: "Disabled option"},
		}},
		{Title: "Backend", Items: []ui.SelectItem{
			{Key: "go", Label: "Go", Description: "Simple and concurrent"},
			{Key: "rust", Label: "Rust", Description: "Fast and memory safe"},
		}},
	}
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Select"),
		ui.Size(920, 760),
	)
}
