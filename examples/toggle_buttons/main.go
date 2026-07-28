package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	selected map[string]bool
	last     string
}

type SetSelected struct {
	key      string
	selected bool
}

func Update(model *Model, msg SetSelected) {
	if model.selected == nil {
		model.selected = make(map[string]bool)
	}
	model.selected[msg.key] = msg.selected
	model.last = fmt.Sprintf("%s: %t", msg.key, msg.selected)
}

func View(_ *ui.Context, model Model, send ui.Send[SetSelected]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Toggle Buttons").Size(24),
				ui.Text(model.last).Size(12),
				ui.Divider(),
				section("Basic",
					buttonRow(toggle("like", selectedLabel(model, "like", "Liked", "Like"), model, send)),
				),
				section("Variants",
					buttonRow(
						toggle("default", "Default", model, send),
						toggle("ghost", "Ghost", model, send).Variant(ui.ToggleButtonGhost),
					),
				),
				section("Sizes",
					buttonRow(
						toggle("small", "Small", model, send).Size(ui.ToggleButtonSmall),
						toggle("medium", "Medium", model, send),
						toggle("large", "Large", model, send).Size(ui.ToggleButtonLarge),
					),
				),
				section("Icon only",
					buttonRow(
						iconToggle("bold", "B", "Bold", model, send),
						iconToggle("italic", "I", "Italic", model, send).Variant(ui.ToggleButtonGhost),
					),
				),
				section("States",
					buttonRow(
						toggle("selected", "Selected", model, send),
						ui.ToggleButton("disabled", false, ui.Text("Disabled")).Disabled(true),
						ui.ToggleButton("disabled-selected", true, ui.Text("Selected disabled")).Disabled(true),
					),
				),
			).Gap(20),
		).Style(ui.FillWidth().MaxWidth(720).Padding(24)),
	)
}

func toggle(key, label string, model Model, send ui.Send[SetSelected]) ui.ToggleButtonWidget {
	return ui.ToggleButton(key, model.selected[key], ui.Text(label)).OnChange(func(selected bool) {
		send(SetSelected{key: key, selected: selected})
	})
}

func iconToggle(key, glyph, label string, model Model, send ui.Send[SetSelected]) ui.ToggleButtonWidget {
	return ui.ToggleButton(key, model.selected[key], ui.Text(glyph)).
		IconOnly().
		Label(label).
		OnChange(func(selected bool) {
			send(SetSelected{key: key, selected: selected})
		})
}

func selectedLabel(model Model, key, selected, unselected string) string {
	if model.selected[key] {
		return selected
	}
	return unselected
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func buttonRow(buttons ...ui.Widget) ui.Widget {
	return ui.Wrap(buttons...).Gap(10).LineGap(10).AlignMiddle()
}

func main() {
	ui.Run(
		Model{selected: map[string]bool{"selected": true}},
		Update,
		View,
		ui.Title("FlowUI Toggle Buttons"),
		ui.Size(820, 640),
	)
}
