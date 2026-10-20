package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	selected map[string]bool
	last     string
}

type SetSelected struct {
	key      string
	selected bool
}

func Update(model *Model, msg SetSelected) ui.Cmd[SetSelected] {
	if model.selected == nil {
		model.selected = make(map[string]bool)
	}
	model.selected[msg.key] = msg.selected
	model.last = fmt.Sprintf("%s: %t", msg.key, msg.selected)
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[SetSelected]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Toggle Buttons").Size(24),
				text.New(model.last).Size(12),
				layout.Divider(),
				section("Basic",
					buttonRow(toggle("like", selectedLabel(model, "like", "Liked", "Like"), model, send)),
				),
				section("Variants",
					buttonRow(
						toggle("default", "Default", model, send),
						toggle("ghost", "Ghost", model, send).Variant(togglebutton.Ghost),
					),
				),
				section("Sizes",
					buttonRow(
						toggle("small", "Small", model, send).Size(togglebutton.Small),
						toggle("medium", "Medium", model, send),
						toggle("large", "Large", model, send).Size(togglebutton.Large),
					),
				),
				section("Icon only",
					buttonRow(
						iconToggle("bold", "B", "Bold", model, send),
						iconToggle("italic", "I", "Italic", model, send).Variant(togglebutton.Ghost),
					),
				),
				section("States",
					buttonRow(
						toggle("selected", "Selected", model, send),
						togglebutton.New("").Key("disabled").Selected(false).Content(text.New("Disabled")).Disabled(true),
						togglebutton.New("").Key("disabled-selected").Selected(true).Content(text.New("Selected disabled")).Disabled(true),
					),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func toggle(key, label string, model Model, send ui.Send[SetSelected]) togglebutton.Widget {
	return togglebutton.New("").Key(key).Selected(model.selected[key]).Content(text.New(label)).OnChange(func(selected bool) {
		send(SetSelected{key: key, selected: selected})
	})
}

func iconToggle(key, glyph, label string, model Model, send ui.Send[SetSelected]) togglebutton.Widget {
	return togglebutton.New("").Key(key).Selected(model.selected[key]).Content(text.New(glyph)).
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
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func buttonRow(buttons ...ui.Widget) ui.Widget {
	return layout.Wrap(buttons...).Gap(10).LineGap(10).AlignMiddle()
}

func main() {
	ui.Run(ui.NewProgram(Model{selected: map[string]bool{"selected": true}},
		Update, View), window.Title("FlowUI Toggle Buttons"),
		window.Size(820, 640),
		window.Theme(base.Light()),
	)
}
