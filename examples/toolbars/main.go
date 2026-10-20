package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	iconpkg "github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/widget/toolbar"
	"github.com/qianniancn/flowui/widget/tooltip"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Bold      bool
	Italic    bool
	Underline bool
	Last      string
}

type Msg any
type Action string
type Format struct {
	Key      string
	Selected bool
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Action:
		model.Last = string(msg)
	case Format:
		switch msg.Key {
		case "bold":
			model.Bold = msg.Selected
		case "italic":
			model.Italic = msg.Selected
		case "underline":
			model.Underline = msg.Selected
		}
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	last := "Ready"
	if model.Last != "" {
		last = "Last action: " + model.Last
	}
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Toolbar").Size(24),
				text.New(last).Size(13),
				layout.Divider(),
				section("Default", historyToolbar("default", send)),
				section("Attached", formattingToolbar(model, send).Attached(true)),
				section("Vertical", layout.Row(
					alignmentToolbar(send),
					(surface.New(
						layout.Box(
							layout.Column(
								text.New("Editor preview").Size(14),
								text.New("Use Tab to enter a toolbar and arrow keys to move between its tools.").Size(14),
							).Gap(10),
						).Style(style.New().Padding(16).FillWidth()),
					).Variant(surface.Secondary)).Style(style.New().Radius(12)),
				).AlignMiddle().Gap(20)),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func historyToolbar(prefix string, send ui.Send[Msg]) toolbar.Widget {
	return toolbar.New(
		toolButton(prefix+"-undo", "Undo", lucide.Undo2, send),
		toolButton(prefix+"-redo", "Redo", lucide.Redo2, send),
		toolbar.Separator(),
		toolButton(prefix+"-copy", "Copy", lucide.Copy, send),
		toolButton(prefix+"-cut", "Cut", lucide.Scissors, send),
	).Alt("History and clipboard tools")
}

func formattingToolbar(model Model, send ui.Send[Msg]) toolbar.Widget {
	return toolbar.New(
		formatButton("bold", "Bold", lucide.Bold, model.Bold, send),
		formatButton("italic", "Italic", lucide.Italic, model.Italic, send),
		formatButton("underline", "Underline", lucide.Underline, model.Underline, send),
		toolbar.Separator(),
		toolButton("format-copy", "Copy", lucide.Copy, send),
	).Alt("Text formatting tools")
}

func alignmentToolbar(send ui.Send[Msg]) toolbar.Widget {
	return toolbar.New(
		toolButton("align-left", "Align left", lucide.TextAlignStart, send),
		toolButton("align-center", "Align center", lucide.TextAlignCenter, send),
		toolButton("align-right", "Align right", lucide.TextAlignEnd, send),
	).Orientation(toolbar.Vertical).Attached(true).Alt("Text alignment tools")
}

func toolButton(key, label string, icon []byte, send ui.Send[Msg]) ui.Widget {
	return tooltip.New(key+"-tooltip", button.New("").Key(key).Content(iconpkg.New(icon).Size(18)).
		Variant(button.Secondary).
		IconOnly().
		OnClick(func() { send(Action(label)) }), label).Delay(0)
}

func formatButton(key, label string, icon []byte, selected bool, send ui.Send[Msg]) ui.Widget {
	return tooltip.New(key+"-tooltip", togglebutton.New("").Key(key).Selected(selected).Content(iconpkg.New(icon).Size(18)).
		IconOnly().
		Label(label).
		OnChange(func(selected bool) { send(Format{Key: key, Selected: selected}) }), label).Delay(0)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(16), child).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Toolbar"), window.Size(860, 620), window.Theme(base.Light()))
}
