package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/tooltip"
	"github.com/qianniancn/flowui/widget/tree"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Selected string
	Expanded []string
}

type Msg any

type SelectFile string
type ExpandFiles []string

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case SelectFile:
		model.Selected = string(msg)
	case ExpandFiles:
		model.Expanded = append([]string(nil), msg...)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	workspace := layout.SplitPane(explorer(model, send),
		layout.SplitPane(editorPane(model.Selected),
			outputPane(),
		).
			Vertical().
			DefaultRatio(.72).
			MinFirst(220).
			MinSecond(120).
			Label("Resize editor and output"),
	).
		DefaultRatio(.26).
		MinFirst(210).
		MinSecond(480).
		Label("Resize explorer and editor")

	return surface.New(
		layout.Column(
			appBar(),
			layout.Divider(),
			layout.Expanded(workspace),
		),
	).Variant(surface.Default)
}

func appBar() ui.Widget {
	return layout.Box(
		layout.Row(
			icon.New(lucide.PanelsTopLeft).Size(18),
			text.New("FlowUI Studio").Size(15),
			layout.Expanded(layout.Spacer(0, 0)),
			tooltip.New("search-help", button.New("").Content(icon.New(lucide.Search).Size(16)).Variant(button.Ghost).Size(button.Small).IconOnly(), "Search").Delay(0),
			tooltip.New("run-help", button.New("").Content(icon.New(lucide.Play).Size(16)).Variant(button.Primary).Size(button.Small).IconOnly(), "Run").Delay(0),
		).AlignMiddle().Gap(10),
	).Style(style.New().Padding(10).FillWidth())
}

func explorer(model Model, send ui.Send[Msg]) ui.Widget {
	return surface.New(
		layout.Box(
			layout.Column(
				layout.Row(
					icon.New(lucide.Files).Size(16),
					text.New("Explorer").Size(14),
				).AlignMiddle().Gap(8),
				layout.Divider(),
				layout.Expanded(
					tree.New("project-files", model.Selected, projectFiles()).
						ExpandedKeys(model.Expanded).
						OnChange(func(key string) { send(SelectFile(key)) }).
						OnExpandedChange(func(keys []string) { send(ExpandFiles(keys)) }).
						Variant(tree.VariantDefault),
				),
			).Gap(10),
		).Style(style.New().Padding(12).FillWidth().FillHeight()),
	).Variant(surface.Secondary)
}

func editorPane(selected string) ui.Widget {
	name := "main.go"
	if selected != "" {
		name = selected
	}
	return surface.New(
		layout.Box(
			layout.Column(
				layout.Row(
					icon.New(lucide.FileCode).Size(15),
					text.New(name).Size(13),
					layout.Expanded(layout.Spacer(0, 0)),
					text.New("Go").Size(12),
				).AlignMiddle().Gap(8),
				layout.Divider(),
				layout.Expanded(
					layout.Box(
						layout.Column(
							text.New("package main").Size(14),
							text.New("import \"fmt\"").Size(14),
							layout.Spacer(0, 8),
							text.New("func main() {").Size(14),
							text.New("    fmt.Println(\"FlowUI workspace\")").Size(14),
							text.New("}").Size(14),
						).Gap(7),
					).Style(style.New().Padding(18).FillWidth().FillHeight()),
				),
			).Gap(10),
		).Style(style.New().Padding(12).FillWidth().FillHeight()),
	).Variant(surface.Default)
}

func outputPane() ui.Widget {
	return surface.New(
		layout.Box(
			layout.Column(
				layout.Row(
					icon.New(lucide.Terminal).Size(15),
					text.New("Output").Size(13),
					layout.Expanded(layout.Spacer(0, 0)),
					text.New("Ready").Size(12),
				).AlignMiddle().Gap(8),
				layout.Divider(),
				text.New("> go run .").Size(13),
				text.New("FlowUI workspace").Size(13),
			).Gap(9),
		).Style(style.New().Padding(12).FillWidth().FillHeight()),
	).Variant(surface.Tertiary)
}

func projectFiles() []tree.Item {
	return []tree.Item{
		{
			Key: "cmd", Label: "cmd", Leading: icon.New(lucide.Folder).Size(15),
			Children: []tree.Item{
				{Key: "main.go", Label: "main.go", Leading: icon.New(lucide.FileCode).Size(15)},
			},
		},
		{
			Key: "internal", Label: "internal", Leading: icon.New(lucide.Folder).Size(15),
			Children: []tree.Item{
				{Key: "app.go", Label: "app.go", Leading: icon.New(lucide.FileCode).Size(15)},
				{Key: "view.go", Label: "view.go", Leading: icon.New(lucide.FileCode).Size(15)},
			},
		},
		{Key: "go.mod", Label: "go.mod", Leading: icon.New(lucide.FileText).Size(15)},
		{Key: "README.md", Label: "README.md", Leading: icon.New(lucide.BookOpenText).Size(15)},
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{Selected: "main.go", Expanded: []string{"cmd", "internal"}},
		Update, View), window.Title("FlowUI SplitPane"),
		window.Size(1080, 720),
		window.Theme(base.Light()),
	)
}
