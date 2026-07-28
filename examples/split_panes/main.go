package main

import (
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Selected string
	Expanded []string
}

type Msg any

type SelectFile string
type ExpandFiles []string

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case SelectFile:
		model.Selected = string(msg)
	case ExpandFiles:
		model.Expanded = append([]string(nil), msg...)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	workspace := ui.SplitPane(
		"workspace",
		explorer(model, send),
		ui.SplitPane(
			"editor-output",
			editorPane(model.Selected),
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

	return ui.Surface(
		ui.Column(
			appBar(),
			ui.Divider(),
			ui.Expanded(workspace),
		),
	).Variant(ui.SurfaceDefault)
}

func appBar() ui.Widget {
	return ui.Box(
		ui.Row(
			ui.Icon(lucide.PanelsTopLeft).Size(18),
			ui.Text("FlowUI Studio").Size(15),
			ui.Expanded(ui.Spacer(0, 0)),
			ui.Tooltip(
				"search-help",
				ui.Button("search", ui.Icon(lucide.Search).Size(16)).Variant(ui.ButtonGhost).Size(ui.ButtonSmall).IconOnly(),
				ui.Text("Search"),
			).Delay(0),
			ui.Tooltip(
				"run-help",
				ui.Button("run", ui.Icon(lucide.Play).Size(16)).Variant(ui.ButtonPrimary).Size(ui.ButtonSmall).IconOnly(),
				ui.Text("Run"),
			).Delay(0),
		).AlignMiddle().Gap(10),
	).Style(ui.Padding(10).FillWidth())
}

func explorer(model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Icon(lucide.Files).Size(16),
					ui.Text("Explorer").Size(14),
				).AlignMiddle().Gap(8),
				ui.Divider(),
				ui.Expanded(
					ui.Tree("project-files", model.Selected, projectFiles()).
						ExpandedKeys(model.Expanded).
						OnChange(func(key string) { send(SelectFile(key)) }).
						OnExpandedChange(func(keys []string) { send(ExpandFiles(keys)) }).
						Variant(ui.TreeDefault),
				),
			).Gap(10),
		).Style(ui.Padding(12).FillWidth().FillHeight()),
	).Variant(ui.SurfaceSecondary)
}

func editorPane(selected string) ui.Widget {
	name := "main.go"
	if selected != "" {
		name = selected
	}
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Icon(lucide.FileCode).Size(15),
					ui.Text(name).Size(13),
					ui.Expanded(ui.Spacer(0, 0)),
					ui.Text("Go").Size(12),
				).AlignMiddle().Gap(8),
				ui.Divider(),
				ui.Expanded(
					ui.Box(
						ui.Column(
							ui.Text("package main").Size(14),
							ui.Text("import \"fmt\"").Size(14),
							ui.Spacer(0, 8),
							ui.Text("func main() {").Size(14),
							ui.Text("    fmt.Println(\"FlowUI workspace\")").Size(14),
							ui.Text("}").Size(14),
						).Gap(7),
					).Style(ui.Padding(18).FillWidth().FillHeight()),
				),
			).Gap(10),
		).Style(ui.Padding(12).FillWidth().FillHeight()),
	).Variant(ui.SurfaceDefault)
}

func outputPane() ui.Widget {
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Icon(lucide.Terminal).Size(15),
					ui.Text("Output").Size(13),
					ui.Expanded(ui.Spacer(0, 0)),
					ui.Text("Ready").Size(12),
				).AlignMiddle().Gap(8),
				ui.Divider(),
				ui.Text("> go run .").Size(13),
				ui.Text("FlowUI workspace").Size(13),
			).Gap(9),
		).Style(ui.Padding(12).FillWidth().FillHeight()),
	).Variant(ui.SurfaceTertiary)
}

func projectFiles() []ui.TreeItem {
	return []ui.TreeItem{
		{
			Key: "cmd", Label: "cmd", Leading: ui.Icon(lucide.Folder).Size(15),
			Children: []ui.TreeItem{
				{Key: "main.go", Label: "main.go", Leading: ui.Icon(lucide.FileCode).Size(15)},
			},
		},
		{
			Key: "internal", Label: "internal", Leading: ui.Icon(lucide.Folder).Size(15),
			Children: []ui.TreeItem{
				{Key: "app.go", Label: "app.go", Leading: ui.Icon(lucide.FileCode).Size(15)},
				{Key: "view.go", Label: "view.go", Leading: ui.Icon(lucide.FileCode).Size(15)},
			},
		},
		{Key: "go.mod", Label: "go.mod", Leading: ui.Icon(lucide.FileText).Size(15)},
		{Key: "README.md", Label: "README.md", Leading: ui.Icon(lucide.BookOpenText).Size(15)},
	}
}

func main() {
	ui.Run(
		Model{Selected: "main.go", Expanded: []string{"cmd", "internal"}},
		Update,
		View,
		ui.Title("FlowUI SplitPane"),
		ui.Size(1080, 720),
	)
}
