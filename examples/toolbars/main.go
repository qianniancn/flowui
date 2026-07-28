package main

import (
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
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

func Update(model *Model, msg Msg) {
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
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	last := "Ready"
	if model.Last != "" {
		last = "Last action: " + model.Last
	}
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Toolbar").Size(24),
				ui.Text(last).Size(13),
				ui.Divider(),
				section("Default", historyToolbar("default", send)),
				section("Attached", formattingToolbar(model, send).Attached(true)),
				section("Vertical", ui.Row(
					alignmentToolbar(send),
					ui.Surface(
						ui.Box(
							ui.Column(
								ui.Text("Editor preview").Size(14),
								ui.Text("Use Tab to enter a toolbar and arrow keys to move between its tools.").Size(14),
							).Gap(10),
						).Style(ui.Padding(16).FillWidth()),
					).Variant(ui.SurfaceSecondary).Style(ui.Radius(12)),
				).AlignMiddle().Gap(20)),
			).Gap(18),
		).Style(ui.FillWidth().MaxWidth(760).Padding(24)),
	)
}

func historyToolbar(prefix string, send ui.Send[Msg]) ui.ToolbarWidget {
	return ui.Toolbar(
		toolButton(prefix+"-undo", "Undo", lucide.Undo2, send),
		toolButton(prefix+"-redo", "Redo", lucide.Redo2, send),
		ui.ToolbarSeparator(),
		toolButton(prefix+"-copy", "Copy", lucide.Copy, send),
		toolButton(prefix+"-cut", "Cut", lucide.Scissors, send),
	).Alt("History and clipboard tools")
}

func formattingToolbar(model Model, send ui.Send[Msg]) ui.ToolbarWidget {
	return ui.Toolbar(
		formatButton("bold", "Bold", lucide.Bold, model.Bold, send),
		formatButton("italic", "Italic", lucide.Italic, model.Italic, send),
		formatButton("underline", "Underline", lucide.Underline, model.Underline, send),
		ui.ToolbarSeparator(),
		toolButton("format-copy", "Copy", lucide.Copy, send),
	).Alt("Text formatting tools")
}

func alignmentToolbar(send ui.Send[Msg]) ui.ToolbarWidget {
	return ui.Toolbar(
		toolButton("align-left", "Align left", lucide.TextAlignStart, send),
		toolButton("align-center", "Align center", lucide.TextAlignCenter, send),
		toolButton("align-right", "Align right", lucide.TextAlignEnd, send),
	).Orientation(ui.ToolbarVertical).Attached(true).Alt("Text alignment tools")
}

func toolButton(key, label string, icon []byte, send ui.Send[Msg]) ui.Widget {
	return ui.Tooltip(
		key+"-tooltip",
		ui.Button(key, ui.Icon(icon).Size(18)).
			Variant(ui.ButtonSecondary).
			IconOnly().
			OnClick(func() { send(Action(label)) }),
		ui.Text(label),
	).Delay(0)
}

func formatButton(key, label string, icon []byte, selected bool, send ui.Send[Msg]) ui.Widget {
	return ui.Tooltip(
		key+"-tooltip",
		ui.ToggleButton(key, selected, ui.Icon(icon).Size(18)).
			IconOnly().
			Label(label).
			OnChange(func(selected bool) { send(Format{Key: key, Selected: selected}) }),
		ui.Text(label),
	).Delay(0)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), child).Gap(10)
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI Toolbar"), ui.Size(860, 620))
}
