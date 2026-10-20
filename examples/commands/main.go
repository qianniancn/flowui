package main

import (
	"gioui.org/font"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toolbar"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Bold bool
	Last string
}

type Msg any
type NewDocument struct{}
type SaveDocument struct{}
type ToggleBold struct{}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
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
	return nil
}

type applicationCommands struct {
	all         []command.Command
	newDocument command.Command
	save        command.Command
	bold        command.Command
}

func commands(model Model, send ui.Send[Msg]) applicationCommands {
	newDocument := command.New("new", "New").
		Icon(icon.New(lucide.FilePlus).Size(18)).
		Shortcut(command.KeyShortcut("N", command.ShortcutPrimary)).
		OnExecute(func() { send(NewDocument{}) })
	save := command.New("save", "Save").
		Icon(icon.New(lucide.Save).Size(18)).
		Shortcut(command.KeyShortcut("S", command.ShortcutPrimary)).
		OnExecute(func() { send(SaveDocument{}) })
	bold := command.New("bold", "Bold").
		Icon(icon.New(lucide.Bold).Size(18)).
		Shortcut(command.KeyShortcut("B", command.ShortcutPrimary)).
		Toggle(model.Bold).
		KeepOpen(true).
		OnExecute(func() { send(ToggleBold{}) })
	return applicationCommands{
		all:         []command.Command{newDocument, save, bold},
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
	preview := text.New("FlowUI command system").Size(18)
	if model.Bold {
		preview = preview.Weight(font.Bold)
	}

	content := layout.Column(
		surface.New(
			layout.Box(applicationMenubar(commands)).Style(style.New().Padding(4)),
		).Variant(surface.Secondary),
		layout.Row(
			toolbar.New(
				command.Button("toolbar-new", commands.newDocument),
				command.Button("toolbar-save", commands.save),
				toolbar.Separator(),
				command.Button("toolbar-bold", commands.bold),
			).Attached(true).Alt("Document tools"),
			dropdown.New(
				"more-actions",
				button.New("More").Variant(button.Secondary),
				[]dropdown.Item{
					command.MenuItem(commands.newDocument),
					command.MenuItem(commands.save),
					menu.Separator(),
					command.MenuItem(commands.bold),
				},
			),
		).AlignMiddle().Gap(12),
		(surface.New(
			layout.Box(
				layout.Column(
					text.New("Untitled document").Size(14),
					preview,
				).Gap(12),
			).Style(style.New().FillWidth().Padding(24)),
		).Variant(surface.Secondary)).Style(style.New().Radius(8)),
		text.New(status).Size(13),
	).Gap(16)

	return command.Scope(
		commands.all,
		layout.Center(layout.Box(content).Style(style.New().FillWidth().MaxWidth(720).Padding(24))),
	)
}

func applicationMenubar(commands applicationCommands) menubar.Widget {
	return menubar.New("application-menu", []menubar.Item{
		menubar.NewMenu("file", "File", []menu.Item{
			command.MenuItem(commands.newDocument),
			command.MenuItem(commands.save),
		}),
		menubar.NewMenu("format", "Format", []menu.Item{
			command.MenuItem(commands.bold),
		}),
	}).Alt("Application menu")
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Commands"), window.Size(820, 520), window.Theme(base.Light()))
}
