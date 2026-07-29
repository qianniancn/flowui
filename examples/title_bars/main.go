package main

import (
	"runtime"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

const (
	titleBarHeight       = 36
	titleBarInputHeight  = 26
	titleBarActionSize   = 28
	titleBarContentIcon  = 14
	titleBarLeadingIcon  = 16
	titleBarInputText    = 12
	titleBarInputLeading = 8
	titleBarInputGap     = 6
)

type Model struct {
	LastAction string
	Search     string
}

type Msg interface{ titleBarMsg() }

type actionMsg string
type searchChanged string

func (actionMsg) titleBarMsg()     {}
func (searchChanged) titleBarMsg() {}

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case searchChanged:
		model.Search = string(msg)
	case actionMsg:
		model.LastAction = string(msg)
	}
}

func applicationView(application *ui.Application) ui.View[Model, Msg] {
	return func(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
		status := "Ready"
		if model.LastAction != "" {
			status = "Command: " + model.LastAction
		}

		return ui.Column(
			applicationHeader(application, ctx.WindowState().TopMost, model, send),
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
									ui.Text("    application.Run(...)").Typeface("monospace").Size(14),
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
}

func applicationHeader(application *ui.Application, topMost bool, model Model, send ui.Send[Msg]) ui.Widget {
	center := ui.InputGroup(
		ui.Input("workspace-search", model.Search).
			Placeholder("Search").
			OnChange(func(value string) { send(searchChanged(value)) }),
	).
		Prefix(ui.Icon(lucide.Search).Size(titleBarContentIcon)).
		PrefixPadding(titleBarInputLeading, titleBarInputGap).
		Style(
			ui.Width(280).
				Height(titleBarInputHeight).
				MinHeight(titleBarInputHeight).
				FontSize(titleBarInputText).
				LineHeight(16).
				Outline(1, 0, ui.RGBA(0)).
				When(ui.Focused, ui.Outline(1, 0, ui.TokenFocus)),
		)
	leading := ui.Icon(lucide.AppWindow).Size(titleBarLeadingIcon)
	trailing := titleBarActions(application, topMost, send)
	return ui.WindowTitleBar("workspace-title-bar", "main.go - FlowUI", applicationMenu(send)).
		Leading(leading).
		Center(center).
		Trailing(trailing).
		ShowMinimize(true).
		ShowMaximize(true).
		ShowClose(true).
		Style(ui.Height(titleBarHeight).Background(ui.TokenSurfaceSecondary))
}

func titleBarActions(application *ui.Application, topMost bool, send ui.Send[Msg]) ui.Widget {
	actions := make([]ui.Widget, 0, 3)
	pinIcon := lucide.Pin
	pinLabel := "Keep on top"
	if topMost {
		pinIcon = lucide.PinOff
		pinLabel = "Stop keeping on top"
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		actions = append(actions, ui.ToggleButton("title-top-most", topMost, ui.Icon(pinIcon).Size(titleBarContentIcon)).
			Variant(ui.ToggleButtonGhost).
			Size(ui.ToggleButtonSmall).
			IconOnly().
			Label(pinLabel).
			Style(ui.Width(titleBarActionSize).Height(titleBarActionSize).Radius(4)).
			OnChange(func(enabled bool) {
				application.Configure("main", ui.TopMost(enabled))
				if enabled {
					send(actionMsg("window pinned"))
					return
				}
				send(actionMsg("window unpinned"))
			}))
	}
	actions = append(actions,
		ui.Button("title-notifications", ui.Icon(lucide.Bell).Size(titleBarContentIcon)).
			Variant(ui.ButtonGhost).
			Size(ui.ButtonSmall).
			IconOnly().
			Label("Notifications").
			Style(ui.Width(titleBarActionSize).Height(titleBarActionSize).Radius(4)).
			OnClick(func() { send(actionMsg("notifications")) }),
		ui.Button("title-settings", ui.Icon(lucide.Settings).Size(titleBarContentIcon)).
			Variant(ui.ButtonGhost).
			Size(ui.ButtonSmall).
			IconOnly().
			Label("Settings").
			Style(ui.Width(titleBarActionSize).Height(titleBarActionSize).Radius(4)).
			OnClick(func() { send(actionMsg("settings")) }),
	)
	return ui.Row(actions...).Gap(2)
}

func applicationMenu(send ui.Send[Msg]) ui.MenubarWidget {
	return ui.Menubar("workspace-menu", []ui.MenubarItem{
		ui.MenubarMenu("file", "File", []ui.MenuItem{
			{Key: "new-file", Label: "New File", Shortcut: "Ctrl+N", Leading: ui.Icon(lucide.FilePlus).Size(16)},
			{Key: "open-file", Label: "Open File", Shortcut: "Ctrl+O", Leading: ui.Icon(lucide.FolderOpen).Size(16)},
			{Key: "save", Label: "Save", Shortcut: "Ctrl+S", Leading: ui.Icon(lucide.Save).Size(16)},
		}).OnAction(func(key string) { send(actionMsg(key)) }),
		ui.MenubarMenu("edit", "Edit", []ui.MenuItem{
			{Key: "undo", Label: "Undo", Shortcut: "Ctrl+Z", Leading: ui.Icon(lucide.Undo2).Size(16)},
			{Key: "redo", Label: "Redo", Shortcut: "Ctrl+Y", Leading: ui.Icon(lucide.Redo2).Size(16)},
			ui.MenuSeparator(),
			{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: ui.Icon(lucide.Copy).Size(16)},
		}).OnAction(func(key string) { send(actionMsg(key)) }),
		ui.MenubarMenu("view", "View", []ui.MenuItem{
			{Key: "command-palette", Label: "Command Palette", Shortcut: "Ctrl+Shift+P"},
			{Key: "terminal", Label: "Terminal", Shortcut: "Ctrl+`", Leading: ui.Icon(lucide.Terminal).Size(16)},
		}).OnAction(func(key string) { send(actionMsg(key)) }),
		ui.MenubarMenu("help", "Help", []ui.MenuItem{
			{Key: "about", Label: "About FlowUI", Leading: ui.Icon(lucide.Info).Size(16)},
		}).OnAction(func(key string) { send(actionMsg(key)) }),
	}).Compact(true).Alt("Application menu")
}

func main() {
	options := []ui.Option{
		ui.Title("FlowUI Workspace"),
		ui.Size(1000, 680),
		ui.MinSize(640, 420),
	}
	if ui.WindowTitleBarSupported() {
		options = append(options, ui.Decorated(false))
	}
	application := ui.NewApplication()
	window := ui.NewWindow(
		"main",
		func() Model { return Model{} },
		Update,
		applicationView(application),
		options...,
	)
	application.Run(window)
}
