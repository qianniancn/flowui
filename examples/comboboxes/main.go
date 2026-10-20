package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/input"
	textpkg "github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				textpkg.New("FlowUI ComboBox").Size(24),
				textpkg.New(status).Size(16),
				layout.Divider(),
				section("Variants",
					layout.Column(
						layout.Box(combo("language", fieldLanguage, m.Language, "Search language", languages, send)).
							Style(style.New().Width(320)),
						layout.Box(combo("editor", fieldEditor, m.Editor, "Search editor", editors, send).
							Variant(input.Secondary)).
							Style(style.New().Width(320)),
					).Gap(12),
				),
				section("States",
					layout.Column(
						layout.Box(combo("runtime", fieldRuntime, m.Runtime, "Required runtime", runtimes, send).
							Invalid(m.Runtime == "")).
							Style(style.New().Width(320)),
						layout.Box(combobox.New(languages).Key("disabled").
							Hint("Disabled").
							Disabled(true)).
							Style(style.New().Width(320)),
					).Gap(12),
				),
				section("Default selection",
					layout.Box(combo("animal", fieldAnimal, m.Animal, "Search animal", animals, send)).
						Style(style.New().Width(320)),
				),
				section("Full width",
					combo("full-width", fieldLanguage, m.Language, "Full width", languages, send).
						Variant(input.Secondary).
						FullWidth(),
				),
			).Gap(18),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		textpkg.New(title).Size(18),
		child,
	).Gap(10)
}

func combo(key string, field Field, selected string, hint string, items []combobox.Item, send ui.Send[Msg]) combobox.Widget {
	return combobox.New(items).Key(key).SelectedKey(selected).
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

var languages = []combobox.Item{
	{Key: "go", Label: "Go", Description: "Fast builds and simple deployment"},
	{Key: "rust", Label: "Rust", Description: "Memory safety without a runtime"},
	{Key: "zig", Label: "Zig", Description: "Small toolchain and explicit control"},
	{Key: "swift", Label: "Swift", Description: "Native apps and server work"},
	{Key: "kotlin", Label: "Kotlin", Description: "Concise JVM and Android code"},
}

var editors = []combobox.Item{
	{Key: "zed", Label: "Zed"},
	{Key: "vscode", Label: "VS Code"},
	{Key: "goland", Label: "GoLand"},
	{Key: "vim", Label: "Vim"},
	{Key: "helix", Label: "Helix"},
}

var runtimes = []combobox.Item{
	{Key: "native", Label: "Native"},
	{Key: "wasm", Label: "WebAssembly"},
	{Key: "server", Label: "Server"},
	{Key: "mobile", Label: "Mobile", Disabled: true},
}

var animals = []combobox.Item{
	{Key: "cat", Label: "Cat"},
	{Key: "dog", Label: "Dog"},
	{Key: "fox", Label: "Fox"},
	{Key: "panda", Label: "Panda"},
	{Key: "rabbit", Label: "Rabbit"},
}

func main() {
	ui.Run(ui.NewProgram(Model{Animal: "dog"},
		Update, View), window.Title("FlowUI ComboBox"),
		window.Size(900, 680),
		window.Theme(base.Light()),
	)
}
