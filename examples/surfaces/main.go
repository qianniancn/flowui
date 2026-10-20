package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Surface").Size(24),
				text.New("Semantic background levels for non-overlay content."),
				layout.Divider(),
				layout.Wrap(
					surfaceExample(
						"Default",
						"The standard content surface. It can use the restrained surface shadow when elevation is needed.",
						surface.Default,
						true,
					),
					surfaceExample(
						"Secondary",
						"A medium-prominence layer for grouped controls, settings panels, and nested content.",
						surface.Secondary,
						false,
					),
					surfaceExample(
						"Tertiary",
						"A stronger layer for content that needs additional visual separation.",
						surface.Tertiary,
						false,
					),
					surfaceExample(
						"Transparent",
						"Keeps surface foreground semantics without painting a background.",
						surface.Transparent,
						false,
					),
				).Gap(24).LineGap(24),
			).Gap(24),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func surfaceExample(title, description string, variant surface.Variant, shadow bool) ui.Widget {
	content := layout.Box(
		layout.Column(
			text.New(title).Size(18),
			text.New(description).Size(14),
		).Gap(8),
	).Style(style.New().Width(320).MinHeight(132).Padding(20))

	declaration := style.New().Radius(24)
	if shadow {
		declaration = declaration.Shadow(style.ShadowSurface)
	}
	if variant == surface.Secondary {
		declaration = declaration.
			BorderWidth(2).
			BorderColor(style.RGB(0x9333ea))
	}
	return (surface.New(content).Variant(variant)).Style(declaration)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Surface"),
		window.Size(900, 620),
		window.Theme(base.Light()),
	)
}
