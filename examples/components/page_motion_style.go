package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
)

func motionStylePage(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	directionIcon := lucide.ArrowRight
	directionLabel := "Play forward"
	if model.MotionForward {
		directionIcon = lucide.ArrowLeft
		directionLabel = "Play backward"
	}
	active := ctx.Theme()
	semantic := active.Semantic()
	base := style.New().FontSize(14).TextColor(style.Solid(semantic.Tones[theme.TonePrimary].Fill))
	layered := layout.Column(
		text.New("Declarations merge with style.Join"),
		description.New("A later layer wins on the properties it sets."),
		text.New("Local override").Style(style.Join(base,
			style.New().TextColor(style.Solid(semantic.Tones[theme.ToneDanger].Fill)))),
	).Gap(8)

	return demoPage("Motion & style",
		demoSection{Title: "Animation primitives", Content: demoPanel(layout.Column(
			layout.Row(
				layout.Expanded(text.New("Tween, spring, timeline, layout and rectangle motion").Size(13)),
				button.New("").Content(icon.New(directionIcon).Size(16)).
					Variant(button.Ghost).Size(button.Small).IconOnly().Label(directionLabel).
					OnClick(func() { send(func(model *Model) { model.MotionForward = !model.MotionForward }) }),
			).AlignMiddle().Gap(6),
			layout.AutoGrid(300, motionDemoCards(model, send)...).ColumnGap(16).RowGap(16),
		).Gap(14))},
		demoSection{Title: "Easing curves", Content: demoPanel(layout.AutoGrid(200, easingCurveCards(model.MotionForward)...).ColumnGap(12).RowGap(12))},
		demoSection{Title: "Declaration layering", Content: demoPanel(layered)},
	)
}
