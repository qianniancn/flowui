package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

func motionStylePage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	directionIcon := lucide.ArrowRight
	directionLabel := "Play forward"
	if model.MotionForward {
		directionIcon = lucide.ArrowLeft
		directionLabel = "Play backward"
	}
	scope := ui.StyleScope(
		ui.FontSize(14).TextColor(ui.TokenAccent),
		ui.Column(
			ui.Text("Styles inherited from the scope"),
			ui.Description("Instance styles still have final precedence."),
			ui.Text("Local override").Style(ui.TextColor(ui.TokenDanger)),
		).Gap(8),
	)

	return demoPage("Motion & style",
		demoSection{Title: "Animation primitives", Content: demoPanel(ui.Column(
			ui.Row(
				ui.Expanded(ui.Text("Tween, spring, timeline, layout and rectangle motion").Size(13)),
				ui.Button("catalog-motion-direction", ui.Icon(directionIcon).Size(16)).
					Variant(ui.ButtonGhost).Size(ui.ButtonSmall).IconOnly().Label(directionLabel).
					OnClick(func() { send(func(model *Model) { model.MotionForward = !model.MotionForward }) }),
			).AlignMiddle().Gap(6),
			ui.AutoGrid(300, motionDemoCards(model, send)...).ColumnGap(16).RowGap(16),
		).Gap(14))},
		demoSection{Title: "Easing curves", Content: demoPanel(ui.AutoGrid(200, easingCurveCards(model.MotionForward)...).ColumnGap(12).RowGap(12))},
		demoSection{Title: "StyleScope", Content: demoPanel(scope)},
	)
}
