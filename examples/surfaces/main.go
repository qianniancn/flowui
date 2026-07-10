package main

import ui "github.com/qianniancn/FlowUI"

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Surface").Size(24),
				ui.Text("Semantic background levels for non-overlay content."),
				ui.Divider(),
				ui.Wrap(
					surfaceExample(
						"Default",
						"The standard content surface. It can use the restrained surface shadow when elevation is needed.",
						ui.SurfaceDefault,
						true,
					),
					surfaceExample(
						"Secondary",
						"A medium-prominence layer for grouped controls, settings panels, and nested content.",
						ui.SurfaceSecondary,
						false,
					),
					surfaceExample(
						"Tertiary",
						"A stronger layer for content that needs additional visual separation.",
						ui.SurfaceTertiary,
						false,
					),
					surfaceExample(
						"Transparent",
						"Keeps surface foreground semantics without painting a background.",
						ui.SurfaceTransparent,
						false,
					),
				).Gap(16).LineGap(16),
			).Gap(16),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func surfaceExample(title, description string, variant ui.SurfaceVariant, shadow bool) ui.Widget {
	content := ui.Box(
		ui.Column(
			ui.Text(title).Size(18),
			ui.Text(description).Size(14),
		).Gap(8),
	).Width(320).MinHeight(132).Padding(20)

	return ui.Surface(content).
		Variant(variant).
		Radius(24).
		Shadow(shadow)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Surface"),
		ui.Size(900, 620),
	)
}
