package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct {
	LastAction string
}

type Msg struct {
	Action string
}

func Update(model *Model, msg Msg) {
	model.LastAction = msg.Action
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
					ui.Box(featureCard(send)).FillWidth(),
				).Gap(16),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func variantCard(title, description string, variant ui.CardVariant) ui.Widget {
	return ui.Box(
		ui.Card(
			ui.CardHeader(
				ui.CardTitle(title),
				ui.CardDescription(description),
			),
			ui.CardContent(
				ui.Text("Cards inherit semantic foreground colors from their surface."),
			),
		).Variant(variant),
	).Width(320)
}

func featureCard(send ui.Send[Msg]) ui.Widget {
	return ui.Card(
		ui.CardHeader(
			ui.CardTitle("Create a focused workspace").Size(16),
			ui.CardDescription("Group related content and actions without introducing application state into the component."),
		).Gap(4),
		ui.CardContent(
			ui.Text("Card keeps its content composable, so controls continue to use the normal FlowUI MVU message path."),
		),
		ui.CardFooter(
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

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Cards"),
		ui.Size(900, 680),
	)
}
