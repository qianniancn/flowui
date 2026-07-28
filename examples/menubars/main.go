package main

import (
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui-icons-lucide"
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

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case SetAction:
		model.LastAction = string(msg)
	case SetToolbar:
		model.Toolbar = bool(msg)
	case SetLayout:
		model.Layout = string(msg)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "No command selected"
	if model.LastAction != "" {
		status = "Last command: " + model.LastAction
	}
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Menubar").Size(24),
				ui.Surface(
					ui.Box(applicationMenubar(model, send)).Style(ui.Padding(4)),
				).Variant(ui.SurfaceSecondary),
				ui.Divider(),
				ui.Text(status).Size(14),
				ui.Text("Layout: "+model.Layout).Size(14),
			).Gap(18),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(720)).Style(ui.Padding(24)),
	)
}

func applicationMenubar(model Model, send ui.Send[Msg]) ui.MenubarWidget {
	return ui.Menubar("application-menu", []ui.MenubarItem{
		ui.MenubarMenu("file", "File", []ui.MenuItem{
			{Key: "new", Label: "New", Shortcut: "Ctrl+N", Leading: ui.Icon(lucide.FilePlus).Size(16)},
			{Key: "open", Label: "Open", Shortcut: "Ctrl+O", Leading: ui.Icon(lucide.FolderOpen).Size(16)},
			{Key: "save", Label: "Save", Shortcut: "Ctrl+S", Leading: ui.Icon(lucide.Save).Size(16)},
			ui.MenuSeparator(),
			{
				Key: "export", Label: "Export", Kind: ui.MenuItemSubmenu,
				Children: []ui.MenuItem{
					{Key: "export-pdf", Label: "PDF", Leading: ui.Icon(lucide.FileText).Size(16)},
					{Key: "export-png", Label: "PNG", Leading: ui.Icon(lucide.Image).Size(16)},
				},
			},
			ui.MenuSeparator(),
			{Key: "print", Label: "Print", Shortcut: "Ctrl+P", Leading: ui.Icon(lucide.Printer).Size(16)},
		}).OnAction(func(key string) { send(SetAction(key)) }),
		ui.MenubarMenu("edit", "Edit", []ui.MenuItem{
			{Key: "cut", Label: "Cut", Shortcut: "Ctrl+X", Leading: ui.Icon(lucide.Scissors).Size(16)},
			{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: ui.Icon(lucide.Copy).Size(16)},
			{Key: "paste", Label: "Paste", Shortcut: "Ctrl+V", Leading: ui.Icon(lucide.ClipboardPaste).Size(16)},
		}).OnAction(func(key string) { send(SetAction(key)) }),
		ui.MenubarMenu("view", "View", []ui.MenuItem{
			{Key: "zoom-in", Label: "Zoom in", Shortcut: "Ctrl++", Leading: ui.Icon(lucide.ZoomIn).Size(16)},
			{Key: "zoom-out", Label: "Zoom out", Shortcut: "Ctrl+-", Leading: ui.Icon(lucide.ZoomOut).Size(16)},
			ui.MenuSeparator(),
			{
				Key:      "toolbar",
				Label:    "Show toolbar",
				Kind:     ui.MenuItemCheckbox,
				Checked:  model.Toolbar,
				KeepOpen: true,
			},
			{
				Key: "layout", Label: "Layout", Kind: ui.MenuItemSubmenu,
				Children: []ui.MenuItem{
					{Key: "single", Label: "Single column", Kind: ui.MenuItemRadio, RadioGroup: "layout", Value: "single", Checked: model.Layout == "single", Leading: ui.Icon(lucide.Rows2).Size(16)},
					{Key: "columns", Label: "Two columns", Kind: ui.MenuItemRadio, RadioGroup: "layout", Value: "columns", Checked: model.Layout == "columns", Leading: ui.Icon(lucide.Columns2).Size(16)},
				},
			},
		}).
			OnAction(func(key string) { send(SetAction(key)) }).
			OnCheckedChange(func(_ string, checked bool) { send(SetToolbar(checked)) }).
			OnRadioChange(func(_ string, value string) { send(SetLayout(value)) }),
		ui.MenubarMenu("help", "Help", nil).Disabled(true),
	}).Alt("Application menu")
}

func main() {
	ui.Run(
		Model{Toolbar: true, Layout: "single"},
		Update,
		View,
		ui.Title("FlowUI Menubar"),
		ui.Size(820, 480),
	)
}
