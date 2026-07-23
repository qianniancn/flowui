package main

import (
	"gioui.org/font"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Bold bool
	Last string
}

type Msg any
type NewDocument struct{}
type SaveDocument struct{}
type ToggleBold struct{}

func Update(model *Model, msg Msg) {
	switch msg.(type) {
	case NewDocument:
		model.Bold = false
		model.Last = "New document"
	case SaveDocument:
		model.Last = "Saved"
	case ToggleBold:
		model.Bold = !model.Bold
		if model.Bold {
			model.Last = "Bold enabled"
		} else {
			model.Last = "Bold disabled"
		}
	}
}

type applicationCommands struct {
	all         []ui.Command
	newDocument ui.Command
	save        ui.Command
	bold        ui.Command
}

func commands(model Model, send ui.Send[Msg]) applicationCommands {
	newDocument := ui.NewCommand("new", "New").
		Icon(ui.Icon(lucide.FilePlus).Size(18)).
		Shortcut(ui.KeyShortcut("N", ui.ShortcutPrimary)).
		OnExecute(func() { send(NewDocument{}) })
	save := ui.NewCommand("save", "Save").
		Icon(ui.Icon(lucide.Save).Size(18)).
		Shortcut(ui.KeyShortcut("S", ui.ShortcutPrimary)).
		OnExecute(func() { send(SaveDocument{}) })
	bold := ui.NewCommand("bold", "Bold").
		Icon(ui.Icon(lucide.Bold).Size(18)).
		Shortcut(ui.KeyShortcut("B", ui.ShortcutPrimary)).
		Toggle(model.Bold).
		KeepOpen(true).
		OnExecute(func() { send(ToggleBold{}) })
	return applicationCommands{
		all:         []ui.Command{newDocument, save, bold},
		newDocument: newDocument,
		save:        save,
		bold:        bold,
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	commands := commands(model, send)
	status := model.Last
	if status == "" {
		status = "Ready"
	}
	preview := ui.Text("FlowUI command system").Size(18)
	if model.Bold {
		preview = preview.Weight(font.Bold)
	}

	content := ui.Column(
		ui.Surface(
			ui.Box(applicationMenubar(commands)).Padding(4),
		).Variant(ui.SurfaceSecondary),
		ui.Row(
			ui.Toolbar(
				ui.CommandButton("toolbar-new", commands.newDocument),
				ui.CommandButton("toolbar-save", commands.save),
				ui.ToolbarSeparator(),
				ui.CommandButton("toolbar-bold", commands.bold),
			).Attached(true).Alt("Document tools"),
			ui.Dropdown(
				"more-actions",
				ui.Button("more-actions-trigger", ui.Text("More")).Variant(ui.ButtonSecondary),
				[]ui.DropdownItem{
					ui.CommandMenuItem(commands.newDocument),
					ui.CommandMenuItem(commands.save),
					ui.MenuSeparator(),
					ui.CommandMenuItem(commands.bold),
				},
			),
		).AlignMiddle().Gap(12),
		ui.Surface(
			ui.Box(
				ui.Column(
					ui.Text("Untitled document").Size(14),
					preview,
				).Gap(12),
			).FillWidth().Padding(24),
		).Variant(ui.SurfaceSecondary).Style(ui.Radius(8)),
		ui.Text(status).Size(13),
	).Gap(16)

	return ui.CommandScope(
		commands.all,
		ui.Center(ui.Box(content).FillWidth().MaxWidth(720).Padding(24)),
	)
}

func applicationMenubar(commands applicationCommands) ui.MenubarWidget {
	return ui.Menubar("application-menu", []ui.MenubarItem{
		ui.MenubarMenu("file", "File", []ui.MenuItem{
			ui.CommandMenuItem(commands.newDocument),
			ui.CommandMenuItem(commands.save),
		}),
		ui.MenubarMenu("format", "Format", []ui.MenuItem{
			ui.CommandMenuItem(commands.bold),
		}),
	}).Alt("Application menu")
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI Commands"), ui.Size(820, 520))
}
