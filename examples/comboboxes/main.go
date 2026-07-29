package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Language string
	Editor   string
	Runtime  string
	Animal   string
	Last     string
}

type Field string

const (
	fieldLanguage Field = "language"
	fieldEditor   Field = "editor"
	fieldRuntime  Field = "runtime"
	fieldAnimal   Field = "animal"
)

type Msg struct {
	Field Field
	Key   string
	Text  string
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.Field {
	case fieldLanguage:
		m.Language = msg.Key
	case fieldEditor:
		m.Editor = msg.Key
	case fieldRuntime:
		m.Runtime = msg.Key
	case fieldAnimal:
		m.Animal = msg.Key
	}
	if msg.Key != "" {
		m.Last = fmt.Sprintf("%s selected %s", msg.Field, msg.Key)
		return nil
	}
	if msg.Text != "" {
		m.Last = fmt.Sprintf("%s search %q", msg.Field, msg.Text)
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No selection"
	if m.Last != "" {
		status = m.Last
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("comboboxes",
				ui.Column(
					ui.Text("FlowUI ComboBox").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Variants",
						ui.Column(
							ui.Box(combo("language", fieldLanguage, m.Language, "Search language", languages, send)).
								Style(ui.Width(320)),
							ui.Box(combo("editor", fieldEditor, m.Editor, "Search editor", editors, send).
								Variant(ui.InputSecondary)).
								Style(ui.Width(320)),
						).Gap(12),
					),
					section("States",
						ui.Column(
							ui.Box(combo("runtime", fieldRuntime, m.Runtime, "Required runtime", runtimes, send).
								Invalid(m.Runtime == "")).
								Style(ui.Width(320)),
							ui.Box(ui.ComboBox("disabled", "", languages).
								Hint("Disabled").
								Disabled(true)).
								Style(ui.Width(320)),
						).Gap(12),
					),
					section("Default selection",
						ui.Box(combo("animal", fieldAnimal, m.Animal, "Search animal", animals, send)).
							Style(ui.Width(320)),
					),
					section("Full width",
						combo("full-width", fieldLanguage, m.Language, "Full width", languages, send).
							Variant(ui.InputSecondary).
							FullWidth(),
					),
				).Gap(18),
			).Vertical(),
		).Style(ui.FillWidth().MaxWidth(720).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func combo(key string, field Field, selected string, hint string, items []ui.ComboBoxItem, send ui.Send[Msg]) ui.ComboBoxWidget {
	return ui.ComboBox(key, selected, items).
		Hint(hint).
		OnChange(func(selected string) {
			send(Msg{
				Field: field,
				Key:   selected,
			})
		}).
		OnInputChange(func(text string) {
			send(Msg{
				Field: field,
				Text:  text,
			})
		})
}

var languages = []ui.ComboBoxItem{
	{Key: "go", Label: "Go", Description: "Fast builds and simple deployment"},
	{Key: "rust", Label: "Rust", Description: "Memory safety without a runtime"},
	{Key: "zig", Label: "Zig", Description: "Small toolchain and explicit control"},
	{Key: "swift", Label: "Swift", Description: "Native apps and server work"},
	{Key: "kotlin", Label: "Kotlin", Description: "Concise JVM and Android code"},
}

var editors = []ui.ComboBoxItem{
	{Key: "zed", Label: "Zed"},
	{Key: "vscode", Label: "VS Code"},
	{Key: "goland", Label: "GoLand"},
	{Key: "vim", Label: "Vim"},
	{Key: "helix", Label: "Helix"},
}

var runtimes = []ui.ComboBoxItem{
	{Key: "native", Label: "Native"},
	{Key: "wasm", Label: "WebAssembly"},
	{Key: "server", Label: "Server"},
	{Key: "mobile", Label: "Mobile", Disabled: true},
}

var animals = []ui.ComboBoxItem{
	{Key: "cat", Label: "Cat"},
	{Key: "dog", Label: "Dog"},
	{Key: "fox", Label: "Fox"},
	{Key: "panda", Label: "Panda"},
	{Key: "rabbit", Label: "Rabbit"},
}

func main() {
	ui.Run(ui.NewProgram(Model{Animal: "dog"},
		Update, View), ui.Title("FlowUI ComboBox"),
		ui.Size(900, 680),
	)
}
