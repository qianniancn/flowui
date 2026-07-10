package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Language string
	Editor   string
	Runtime  string
	Last     string
}

type Field string

const (
	fieldLanguage Field = "language"
	fieldEditor   Field = "editor"
	fieldRuntime  Field = "runtime"
)

type Msg struct {
	Field Field
	Key   string
	Text  string
}

func Update(m *Model, msg Msg) {
	switch msg.Field {
	case fieldLanguage:
		m.Language = msg.Key
	case fieldEditor:
		m.Editor = msg.Key
	case fieldRuntime:
		m.Runtime = msg.Key
	}
	if msg.Key != "" {
		m.Last = fmt.Sprintf("%s selected %s", msg.Field, msg.Key)
		return
	}
	if msg.Text != "" {
		m.Last = fmt.Sprintf("%s search %q", msg.Field, msg.Text)
	}
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
								Width(320),
							ui.Box(combo("editor", fieldEditor, m.Editor, "Search editor", editors, send).
								Variant(ui.InputSecondary)).
								Width(320),
						).Gap(12),
					),
					section("States",
						ui.Column(
							ui.Box(combo("runtime", fieldRuntime, m.Runtime, "Required runtime", runtimes, send).
								Invalid(m.Runtime == "")).
								Width(320),
							ui.Box(ui.ComboBox("disabled", "go", languages).
								Hint("Disabled").
								Disabled(true)).
								Width(320),
						).Gap(12),
					),
					section("Full width",
						combo("full-width", fieldLanguage, m.Language, "Full width", languages, send).
							Variant(ui.InputSecondary).
							FullWidth(),
					),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(720).Padding(24),
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

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI ComboBox"),
		ui.Size(900, 680),
	)
}
