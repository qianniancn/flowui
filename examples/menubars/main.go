package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	LastAction string
	Toolbar    bool
	Layout     string
}

type Msg any
type SetAction string
type SetToolbar bool
type SetLayout string

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case SetAction:
		model.LastAction = string(msg)
	case SetToolbar:
		model.Toolbar = bool(msg)
	case SetLayout:
		model.Layout = string(msg)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "No command selected"
	if model.LastAction != "" {
		status = "Last command: " + model.LastAction
	}
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Menubar").Size(24),
				surface.New(
					layout.Box(applicationMenubar(model, send)).Style(style.New().Padding(4)),
				).Variant(surface.Secondary),
				layout.Divider(),
				text.New(status).Size(14),
				text.New("Layout: "+model.Layout).Size(14),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func applicationMenubar(model Model, send ui.Send[Msg]) menubar.Widget {
	return menubar.New("application-menu", []menubar.Item{
		menubar.NewMenu("file", "File", []menu.Item{
			{Key: "new", Label: "New", Shortcut: "Ctrl+N", Leading: icon.New(lucide.FilePlus).Size(16)},
			{Key: "open", Label: "Open", Shortcut: "Ctrl+O", Leading: icon.New(lucide.FolderOpen).Size(16)},
			{Key: "save", Label: "Save", Shortcut: "Ctrl+S", Leading: icon.New(lucide.Save).Size(16)},
			menu.Separator(),
			{
				Key: "export", Label: "Export", Kind: menu.ItemSubmenu,
				Children: []menu.Item{
					{Key: "export-pdf", Label: "PDF", Leading: icon.New(lucide.FileText).Size(16)},
					{Key: "export-png", Label: "PNG", Leading: icon.New(lucide.Image).Size(16)},
				},
			},
			menu.Separator(),
			{Key: "print", Label: "Print", Shortcut: "Ctrl+P", Leading: icon.New(lucide.Printer).Size(16)},
		}).OnActionEvent(func(event menu.ActionEvent) { send(SetAction(event.Key)) }),
		menubar.NewMenu("edit", "Edit", []menu.Item{
			{Key: "cut", Label: "Cut", Shortcut: "Ctrl+X", Leading: icon.New(lucide.Scissors).Size(16)},
			{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: icon.New(lucide.Copy).Size(16)},
			{Key: "paste", Label: "Paste", Shortcut: "Ctrl+V", Leading: icon.New(lucide.ClipboardPaste).Size(16)},
		}).OnActionEvent(func(event menu.ActionEvent) { send(SetAction(event.Key)) }),
		menubar.NewMenu("view", "View", []menu.Item{
			{Key: "zoom-in", Label: "Zoom in", Shortcut: "Ctrl++", Leading: icon.New(lucide.ZoomIn).Size(16)},
			{Key: "zoom-out", Label: "Zoom out", Shortcut: "Ctrl+-", Leading: icon.New(lucide.ZoomOut).Size(16)},
			menu.Separator(),
			{
				Key:      "toolbar",
				Label:    "Show toolbar",
				Kind:     menu.ItemCheckbox,
				Checked:  model.Toolbar,
				KeepOpen: true,
			},
			{
				Key: "layout", Label: "Layout", Kind: menu.ItemSubmenu,
				Children: []menu.Item{
					{Key: "single", Label: "Single column", Kind: menu.ItemRadio, RadioGroup: "layout", Value: "single", Checked: model.Layout == "single", Leading: icon.New(lucide.Rows2).Size(16)},
					{Key: "columns", Label: "Two columns", Kind: menu.ItemRadio, RadioGroup: "layout", Value: "columns", Checked: model.Layout == "columns", Leading: icon.New(lucide.Columns2).Size(16)},
				},
			},
		}).
			OnActionEvent(func(event menu.ActionEvent) { send(SetAction(event.Key)) }).
			OnCheckedChange(func(_ string, checked bool) { send(SetToolbar(checked)) }).
			OnRadioChange(func(_ string, value string) { send(SetLayout(value)) }),
		menubar.NewMenu("help", "Help", nil).Disabled(true),
	}).Alt("Application menu")
}

func main() {
	ui.Run(ui.NewProgram(Model{Toolbar: true, Layout: "single"},
		Update, View), window.Title("FlowUI Menubar"),
		window.Size(820, 480),
		window.Theme(base.Light()),
	)
}
