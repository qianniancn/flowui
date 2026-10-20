package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	LastAction string
}

type Msg struct {
	Action string
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	model.LastAction = msg.Action
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "Choose an action from a card."
	if model.LastAction != "" {
		status = "Last action: " + model.LastAction
	}

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Cards").Size(24),
				text.New(status),
				layout.Divider(),
				text.New("Variants").Size(18),
				layout.Wrap(
					variantCard("Transparent", "Minimal prominence for nested content.", card.Transparent),
					variantCard("Default", "The standard surface for most content.", card.Default),
					variantCard("Secondary", "A medium-prominence grouped surface.", card.Secondary),
					variantCard("Tertiary", "A stronger surface for featured content.", card.Tertiary),
				).Gap(16).LineGap(16),
				text.New("Composition").Size(18),
				layout.Box(featureCard(send)).Style(style.New().FillWidth()),
			).Gap(16),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func variantCard(title, description string, variant card.Variant) ui.Widget {
	return layout.Box(
		card.New(
			cardHeader(title, description),
			text.New("Cards inherit semantic foreground colors from their surface."),
		).Variant(variant),
	).Style(style.New().Width(320))
}

func featureCard(send ui.Send[Msg]) ui.Widget {
	return card.New(
		cardHeader("Create a focused workspace", "Group related content and actions without introducing application state into the component."),
		text.New("Card keeps its content composable, so controls continue to use the normal FlowUI MVU message path."),
		layout.Row(
			button.New("Later").
				Size(button.Small).
				Variant(button.Ghost).
				OnClick(func() { send(Msg{Action: "Later"}) }),
			button.New("Create workspace").
				Size(button.Small).
				OnClick(func() { send(Msg{Action: "Create workspace"}) }),
		).Gap(8),
	).Variant(card.Default)
}

func cardHeader(title, description string) ui.Widget {
	return layout.Column(text.New(title).Size(16), text.New(description).Size(14)).Gap(4)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Cards"),
		window.Size(900, 680),
		window.Theme(base.Light()),
	)
}
