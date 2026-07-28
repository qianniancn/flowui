package main

import (
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Last string
}

type Selected string

func Update(model *Model, msg Selected) {
	model.Last = string(msg)
}

func View(_ *ui.Context, model Model, send ui.Send[Selected]) ui.Widget {
	status := "Choose an action"
	if model.Last != "" {
		status = "Last action: " + model.Last
	}
	return ui.Center(
		ui.Box(
			ui.Scroll("button-groups",
				ui.Column(
					ui.Text("ButtonGroup").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					section("Default",
						group("default", ui.ButtonPrimary, ui.ButtonMedium, send),
					),
					section("Sizes",
						ui.Column(
							group("small", ui.ButtonSecondary, ui.ButtonSmall, send),
							group("medium", ui.ButtonSecondary, ui.ButtonMedium, send),
							group("large", ui.ButtonSecondary, ui.ButtonLarge, send),
						).Gap(12),
					),
					section("Variants",
						ui.Column(
							group("tertiary", ui.ButtonTertiary, ui.ButtonMedium, send),
							group("outline", ui.ButtonOutline, ui.ButtonMedium, send),
							group("danger", ui.ButtonDanger, ui.ButtonMedium, send),
						).Gap(12),
					),
					section("Full width",
						ui.Box(
							ui.ButtonGroup(
								iconButton("align-left", "Align left", lucide.TextAlignStart, send),
								iconButton("align-center", "Align center", lucide.TextAlignCenter, send),
								iconButton("align-right", "Align right", lucide.TextAlignEnd, send),
							).Variant(ui.ButtonTertiary).FullWidth().Separators(true),
						).Style(ui.Width(420)),
					),
					section("Vertical",
						ui.ButtonGroup(
							iconTextButton("photos", "Photos", lucide.Image, send),
							iconTextButton("videos", "Videos", lucide.Video, send),
							iconTextButton("more", "More", lucide.Ellipsis, send),
						).
							Orientation(ui.ButtonGroupVertical).
							Variant(ui.ButtonTertiary).
							Separators(true),
					),
					section("Disabled",
						ui.ButtonGroup(
							textButton("disabled-first", "First", send),
							textButton("disabled-second", "Second", send),
							textButton("enabled-third", "Enabled", send).Disabled(false),
						).Variant(ui.ButtonSecondary).Disabled(true).Separators(true),
					),
				).Gap(18),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func group(prefix string, variant ui.ButtonVariant, size ui.ButtonSize, send ui.Send[Selected]) ui.ButtonGroupWidget {
	return ui.ButtonGroup(
		textButton(prefix+"-first", "First", send),
		textButton(prefix+"-second", "Second", send),
		textButton(prefix+"-third", "Third", send),
	).Variant(variant).Size(size).Separators(true)
}

func textButton(key, label string, send ui.Send[Selected]) ui.ButtonWidget {
	return ui.Button(key, ui.Text(label)).OnClick(func() { send(Selected(label)) })
}

func iconButton(key, label string, data []byte, send ui.Send[Selected]) ui.ButtonWidget {
	return ui.Button(key, ui.Icon(data).Size(16)).
		IconOnly().
		Label(label).
		OnClick(func() { send(Selected(label)) })
}

func iconTextButton(key, label string, data []byte, send ui.Send[Selected]) ui.ButtonWidget {
	return ui.Button(key,
		ui.Row(ui.Icon(data).Size(16), ui.Text(label)).Gap(8).AlignMiddle(),
	).OnClick(func() { send(Selected(label)) })
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), child).Gap(10)
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI ButtonGroup"), ui.Size(860, 720))
}
