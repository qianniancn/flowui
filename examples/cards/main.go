package main

import "github.com/qianniancn/flowui/ui"

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

	return ui.Center(
		ui.Box(
			ui.Scroll("cards",
				ui.Column(
					ui.Text("FlowUI Cards").Size(24),
					ui.Text(status),
					ui.Divider(),
					ui.Text("Variants").Size(18),
					ui.Wrap(
						variantCard("Transparent", "Minimal prominence for nested content.", ui.CardTransparent),
						variantCard("Default", "The standard surface for most content.", ui.CardDefault),
						variantCard("Secondary", "A medium-prominence grouped surface.", ui.CardSecondary),
						variantCard("Tertiary", "A stronger surface for featured content.", ui.CardTertiary),
					).Gap(16).LineGap(16),
					ui.Text("Composition").Size(18),
					ui.Box(featureCard(send)).Style(ui.FillWidth()),
				).Gap(16),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func variantCard(title, description string, variant ui.CardVariant) ui.Widget {
	return ui.Box(
		ui.Card(
			cardHeader(title, description),
			ui.Text("Cards inherit semantic foreground colors from their surface."),
		).Variant(variant),
	).Style(ui.Width(320))
}

func featureCard(send ui.Send[Msg]) ui.Widget {
	return ui.Card(
		cardHeader("Create a focused workspace", "Group related content and actions without introducing application state into the component."),
		ui.Text("Card keeps its content composable, so controls continue to use the normal FlowUI MVU message path."),
		ui.Row(
			ui.Button("card-later", ui.Text("Later")).
				Size(ui.ButtonSmall).
				Variant(ui.ButtonGhost).
				OnClick(func() { send(Msg{Action: "Later"}) }),
			ui.Button("card-create", ui.Text("Create workspace")).
				Size(ui.ButtonSmall).
				OnClick(func() { send(Msg{Action: "Create workspace"}) }),
		).Gap(8),
	).Variant(ui.CardDefault)
}

func cardHeader(title, description string) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), ui.Text(description).Size(14)).Gap(4)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), ui.Title("FlowUI Cards"),
		ui.Size(900, 680),
	)
}
