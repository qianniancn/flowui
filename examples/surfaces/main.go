package main

import "github.com/qianniancn/flowui/ui"

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

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
				).Gap(24).LineGap(24),
			).Gap(24),
		).Style(ui.FillWidth().MaxWidth(760).Padding(24)),
	)
}

func surfaceExample(title, description string, variant ui.SurfaceVariant, shadow bool) ui.Widget {
	content := ui.Box(
		ui.Column(
			ui.Text(title).Size(18),
			ui.Text(description).Size(14),
		).Gap(8),
	).Style(ui.Width(320).MinHeight(132).Padding(20))

	declaration := ui.Radius(24)
	if shadow {
		declaration = declaration.Shadow(ui.ShadowSurface)
	}
	if variant == ui.SurfaceSecondary {
		declaration = declaration.
			BorderWidth(2).
			BorderColor(ui.RGB(0x9333ea))
	}
	return ui.Surface(content).Variant(variant).Style(declaration)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), ui.Title("FlowUI Surface"),
		ui.Size(900, 620),
	)
}
