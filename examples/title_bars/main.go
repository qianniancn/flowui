package main

import (
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	LastAction string
}

type Msg string

func Update(model *Model, msg Msg) {
	model.LastAction = string(msg)
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "Ready"
	if model.LastAction != "" {
		status = "Command: " + model.LastAction
	}

	return ui.Column(
		ui.WindowTitleBar("workspace-title-bar", "main.go - FlowUI", applicationMenu(send)),
		ui.Expanded(
			ui.Row(
				ui.Surface(
					ui.Box(
						ui.Column(
							ui.Text("EXPLORER").Size(11),
							ui.Text("FLOWUI").Size(12),
							ui.Text("  examples").Size(13),
							ui.Text("    title_bars").Size(13),
							ui.Text("      main.go").Size(13),
						).Gap(10),
					).Style(ui.Width(210).FillHeight().Padding(14)),
				).Variant(ui.SurfaceSecondary),
				ui.Expanded(
					ui.Surface(
						ui.Box(
							ui.Column(
								ui.Text("main.go").Size(13),
								ui.Divider(),
								ui.Text("func main() {").Typeface("monospace").Size(14),
								ui.Text("    ui.Run(...)").Typeface("monospace").Size(14),
								ui.Text("}").Typeface("monospace").Size(14),
							).Gap(10),
						).Style(ui.FillWidth().FillHeight().Padding(20)),
					),
				),
			),
		),
		ui.StatusBar(ui.Text(status).Size(12), ui.Text("Ln 3, Col 16").Size(12)),
	)
}

func applicationMenu(send ui.Send[Msg]) ui.MenubarWidget {
	return ui.Menubar("workspace-menu", []ui.MenubarItem{
		ui.MenubarMenu("file", "File", []ui.MenuItem{
			{Key: "new-file", Label: "New File", Shortcut: "Ctrl+N", Leading: ui.Icon(lucide.FilePlus).Size(16)},
			{Key: "open-file", Label: "Open File", Shortcut: "Ctrl+O", Leading: ui.Icon(lucide.FolderOpen).Size(16)},
			{Key: "save", Label: "Save", Shortcut: "Ctrl+S", Leading: ui.Icon(lucide.Save).Size(16)},
		}).OnAction(func(key string) { send(Msg(key)) }),
		ui.MenubarMenu("edit", "Edit", []ui.MenuItem{
			{Key: "undo", Label: "Undo", Shortcut: "Ctrl+Z", Leading: ui.Icon(lucide.Undo2).Size(16)},
			{Key: "redo", Label: "Redo", Shortcut: "Ctrl+Y", Leading: ui.Icon(lucide.Redo2).Size(16)},
			ui.MenuSeparator(),
			{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: ui.Icon(lucide.Copy).Size(16)},
		}).OnAction(func(key string) { send(Msg(key)) }),
		ui.MenubarMenu("view", "View", []ui.MenuItem{
			{Key: "command-palette", Label: "Command Palette", Shortcut: "Ctrl+Shift+P"},
			{Key: "terminal", Label: "Terminal", Shortcut: "Ctrl+`", Leading: ui.Icon(lucide.Terminal).Size(16)},
		}).OnAction(func(key string) { send(Msg(key)) }),
		ui.MenubarMenu("help", "Help", []ui.MenuItem{
			{Key: "about", Label: "About FlowUI", Leading: ui.Icon(lucide.Info).Size(16)},
		}).OnAction(func(key string) { send(Msg(key)) }),
	}).Compact(true).Alt("Application menu")
}

func main() {
	ui.Run(
		Model{},
		Update,
		View,
		ui.Title("FlowUI Workspace"),
		ui.Size(1000, 680),
		ui.MinSize(640, 420),
		ui.Decorated(false),
	)
}
