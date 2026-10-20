package main

import (
	"runtime"

	giolayout "gioui.org/layout"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/titlebar"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/window"
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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case searchChanged:
		model.Search = string(msg)
	case actionMsg:
		model.LastAction = string(msg)
	}
	return nil
}

func applicationView(application *ui.Application) ui.View[Model, Msg] {
	return func(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
		status := "Ready"
		if model.LastAction != "" {
			status = "Command: " + model.LastAction
		}

		return layout.Column(
			applicationHeader(application, ctx.WindowState().TopMost, model, send),
			layout.Expanded(
				layout.Row(
					surface.New(
						layout.Box(
							layout.Column(
								text.New("EXPLORER").Size(11),
								text.New("FLOWUI").Size(12),
								text.New("  examples").Size(13),
								text.New("    title_bars").Size(13),
								text.New("      main.go").Size(13),
							).Gap(10),
						).Style(style.New().Width(210).FillHeight().Padding(14)),
					).Variant(surface.Secondary),
					layout.Expanded(
						surface.New(
							layout.Box(
								layout.Column(
									text.New("main.go").Size(13),
									layout.Divider(),
									text.New("func main() {").Typeface("monospace").Size(14),
									text.New("    application.Run(...)").Typeface("monospace").Size(14),
									text.New("}").Typeface("monospace").Size(14),
								).Gap(10),
							).Style(style.New().FillWidth().FillHeight().Padding(20)),
						),
					),
				),
			),
			statusbar.New(text.New(status).Size(12), text.New("Ln 3, Col 16").Size(12)),
		)
	}
}

func applicationHeader(application *ui.Application, topMost bool, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		active := ctx.Theme()
		sem := active.Semantic()
		center := (input.Group(
			input.New(model.Search).Key("workspace-search").
				Placeholder("Search").
				OnChange(func(value string) { send(searchChanged(value)) }),
		).
			Prefix(icon.New(lucide.Search).Size(titleBarContentIcon))).Style(style.New().Width(280).
			Height(titleBarInputHeight).
			MinHeight(titleBarInputHeight).
			FontSize(titleBarInputText).
			LineHeight(16).
			Outline(1, 0, style.RGBA(0)).
			On(style.Focused, style.New().Outline(1, 0, style.Solid(sem.Ring))))
		leading := icon.New(lucide.AppWindow).Size(titleBarLeadingIcon)
		trailing := titleBarActions(application, topMost, send)
		bar := (titlebar.New("workspace-title-bar", "main.go - FlowUI", applicationMenu(send)).
			Leading(leading).
			Center(center).
			Trailing(trailing).
			ShowMinimize(true).
			ShowMaximize(true).
			ShowClose(true)).Style(style.New().Height(titleBarHeight).Background(style.Solid(sem.SurfaceSecondary)))
		return bar.Layout(ctx, gtx)
	})
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
		actions = append(actions, (togglebutton.New("").Key("title-top-most").Selected(topMost).Content(icon.New(pinIcon).Size(titleBarContentIcon)).
			Variant(togglebutton.Ghost).
			Size(togglebutton.Small).
			IconOnly().
			Label(pinLabel).
			OnChange(func(enabled bool) {
				application.Configure("main", window.TopMost(enabled))
				if enabled {
					send(actionMsg("window pinned"))
					return
				}
				send(actionMsg("window unpinned"))
			})).Style(style.New().Width(titleBarActionSize).Height(titleBarActionSize).Radius(4)))
	}
	actions = append(actions,
		(button.New("").Content(icon.New(lucide.Bell).Size(titleBarContentIcon)).
			Variant(button.Ghost).
			Size(button.Small).
			IconOnly().
			Label("Notifications").
			OnClick(func() { send(actionMsg("notifications")) })).Style(style.New().Width(titleBarActionSize).Height(titleBarActionSize).Radius(4)),
		(button.New("").Content(icon.New(lucide.Settings).Size(titleBarContentIcon)).
			Variant(button.Ghost).
			Size(button.Small).
			IconOnly().
			Label("Settings").
			OnClick(func() { send(actionMsg("settings")) })).Style(style.New().Width(titleBarActionSize).Height(titleBarActionSize).Radius(4)),
	)
	return layout.Row(actions...).Gap(2)
}

func applicationMenu(send ui.Send[Msg]) menubar.Widget {
	return menubar.New("workspace-menu", []menubar.Item{
		menubar.NewMenu("file", "File", []menu.Item{
			{Key: "new-file", Label: "New File", Shortcut: "Ctrl+N", Leading: icon.New(lucide.FilePlus).Size(16)},
			{Key: "open-file", Label: "Open File", Shortcut: "Ctrl+O", Leading: icon.New(lucide.FolderOpen).Size(16)},
			{Key: "save", Label: "Save", Shortcut: "Ctrl+S", Leading: icon.New(lucide.Save).Size(16)},
		}).OnActionEvent(func(event menu.ActionEvent) { send(actionMsg(event.Key)) }),
		menubar.NewMenu("edit", "Edit", []menu.Item{
			{Key: "undo", Label: "Undo", Shortcut: "Ctrl+Z", Leading: icon.New(lucide.Undo2).Size(16)},
			{Key: "redo", Label: "Redo", Shortcut: "Ctrl+Y", Leading: icon.New(lucide.Redo2).Size(16)},
			menu.Separator(),
			{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: icon.New(lucide.Copy).Size(16)},
		}).OnActionEvent(func(event menu.ActionEvent) { send(actionMsg(event.Key)) }),
		menubar.NewMenu("view", "View", []menu.Item{
			{Key: "command-palette", Label: "Command Palette", Shortcut: "Ctrl+Shift+P"},
			{Key: "terminal", Label: "Terminal", Shortcut: "Ctrl+`", Leading: icon.New(lucide.Terminal).Size(16)},
		}).OnActionEvent(func(event menu.ActionEvent) { send(actionMsg(event.Key)) }),
		menubar.NewMenu("help", "Help", []menu.Item{
			{Key: "about", Label: "About FlowUI", Leading: icon.New(lucide.Info).Size(16)},
		}).OnActionEvent(func(event menu.ActionEvent) { send(actionMsg(event.Key)) }),
	}).Compact(true).Alt("Application menu")
}

func main() {
	options := []window.Option{
		window.Title("FlowUI Workspace"),
		window.Size(1000, 680),
		window.MinSize(640, 420),
		window.Theme(base.Light()),
	}
	if titlebar.Supported() {
		options = append(options, window.Decorated(false))
	}
	application := ui.NewApplication()
	window := ui.NewWindow(
		"main", ui.NewProgram(Model{}, Update, applicationView(application)), options...,
	)
	application.Run(window)
}
