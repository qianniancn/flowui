package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Last string
}

type Selected string

func Update(model *Model, msg Selected) ui.Cmd[Selected] {
	model.Last = string(msg)
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Selected]) ui.Widget {
	status := "Choose an action"
	if model.Last != "" {
		status = "Last action: " + model.Last
	}
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("ButtonGroup").Size(24),
				text.New(status).Size(14),
				layout.Divider(),
				section("Default",
					group("default", button.Primary, button.Medium, send),
				),
				section("Sizes",
					layout.Column(
						group("small", button.Secondary, button.Small, send),
						group("medium", button.Secondary, button.Medium, send),
						group("large", button.Secondary, button.Large, send),
					).Gap(12),
				),
				section("Variants",
					layout.Column(
						group("tertiary", button.Tertiary, button.Medium, send),
						group("outline", button.Outline, button.Medium, send),
						group("danger", button.Danger, button.Medium, send),
					).Gap(12),
				),
				section("Full width",
					layout.Box(
						button.Group(
							iconButton("align-left", "Align left", lucide.TextAlignStart, send),
							iconButton("align-center", "Align center", lucide.TextAlignCenter, send),
							iconButton("align-right", "Align right", lucide.TextAlignEnd, send),
						).Variant(button.Tertiary).FullWidth().Separators(true),
					).Style(style.New().Width(420)),
				),
				section("Vertical",
					button.Group(
						iconTextButton("photos", "Photos", lucide.Image, send),
						iconTextButton("videos", "Videos", lucide.Video, send),
						iconTextButton("more", "More", lucide.Ellipsis, send),
					).
						Orientation(button.GroupVertical).
						Variant(button.Tertiary).
						Separators(true),
				),
				section("Disabled",
					button.Group(
						textButton("disabled-first", "First", send),
						textButton("disabled-second", "Second", send),
						textButton("enabled-third", "Enabled", send).Disabled(false),
					).Variant(button.Secondary).Disabled(true).Separators(true),
				),
			).Gap(18),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func group(prefix string, variant button.Variant, size button.Size, send ui.Send[Selected]) button.GroupWidget {
	return button.Group(
		textButton(prefix+"-first", "First", send),
		textButton(prefix+"-second", "Second", send),
		textButton(prefix+"-third", "Third", send),
	).Variant(variant).Size(size).Separators(true)
}

func textButton(key, label string, send ui.Send[Selected]) button.Widget {
	return button.New(label).Key(key).OnClick(func() { send(Selected(label)) })
}

func iconButton(key, label string, data []byte, send ui.Send[Selected]) button.Widget {
	return button.New("").Key(key).Content(icon.New(data).Size(16)).
		IconOnly().
		Label(label).
		OnClick(func() { send(Selected(label)) })
}

func iconTextButton(key, label string, data []byte, send ui.Send[Selected]) button.Widget {
	return button.New("").Key(key).Content(layout.Row(icon.New(data).Size(16), text.New(label)).Gap(8).AlignMiddle()).OnClick(func() { send(Selected(label)) })
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(16), child).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI ButtonGroup"), window.Size(860, 720), window.Theme(base.Light()))
}
