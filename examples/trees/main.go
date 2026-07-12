package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Selected        string
	Expanded        []string
	SurfaceSelected string
	SurfaceExpanded []string
	LastAction      string
}

type Msg struct {
	Tree     string
	Selected string
	Expanded []string
	Action   string
}

func Update(model *Model, msg Msg) {
	if msg.Expanded != nil {
		if msg.Tree == "surface" {
			model.SurfaceExpanded = append([]string(nil), msg.Expanded...)
		} else {
			model.Expanded = append([]string(nil), msg.Expanded...)
		}
	}
	if msg.Selected != "" {
		if msg.Tree == "surface" {
			model.SurfaceSelected = msg.Selected
		} else {
			model.Selected = msg.Selected
		}
	}
	if msg.Action != "" {
		model.LastAction = msg.Action
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "Select a file or folder"
	if model.LastAction != "" {
		status = fmt.Sprintf("Activated: %s", model.LastAction)
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("trees",
				ui.Column(
					ui.Text("FlowUI Tree").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					section("File browser",
						ui.Box(controlledTree("files", model.Selected, model.Expanded, fileItems(), send)).
							Width(520),
					),
					section("Surface",
						ui.Box(controlledTree("surface", model.SurfaceSelected, model.SurfaceExpanded, workspaceItems(), send).
							Variant(ui.TreeSurface)).
							Width(520),
					),
					section("Action only",
						ui.Box(ui.Tree("actions", "", actionItems()).
							SelectionMode(ui.TreeSelectionNone).
							ExpandedKeys([]string{"build"}).
							OnAction(func(key string) { send(Msg{Action: key}) })).
							Width(520),
					),
					section("Scrollable",
						ui.Box(ui.Tree("packages", "", packageItems()).
							MaxHeight(180).
							DisabledKeys([]string{"package-7"})).
							Width(520),
					),
				).Gap(20),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func controlledTree(key, selected string, expanded []string, items []ui.TreeItem, send ui.Send[Msg]) ui.TreeWidget {
	return ui.Tree(key, selected, items).
		ExpandedKeys(expanded).
		OnChange(func(selected string) {
			send(Msg{Tree: key, Selected: selected})
		}).
		OnExpandedChange(func(expanded []string) {
			send(Msg{Tree: key, Expanded: expanded})
		}).
		OnAction(func(action string) {
			send(Msg{Action: action})
		})
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func treeIcon(data lucide.Data) ui.Widget {
	return ui.Icon(data).Size(16)
}

func fileItems() []ui.TreeItem {
	return []ui.TreeItem{
		{
			Key: "flowui", Label: "FlowUI", Leading: treeIcon(lucide.Folder),
			Children: []ui.TreeItem{
				{
					Key: "components", Label: "components", Leading: treeIcon(lucide.Folder),
					Children: []ui.TreeItem{
						{Key: "button", Label: "button.go", Leading: treeIcon(lucide.FileCode)},
						{Key: "tree", Label: "tree.go", Leading: treeIcon(lucide.FileCode), Trailing: ui.Chip("New").Color(ui.ChipAccent).Size(ui.ChipSmall)},
					},
				},
				{Key: "assets", Label: "assets", Leading: treeIcon(lucide.Folder), Children: []ui.TreeItem{
					{Key: "logo", Label: "logo.png", Leading: treeIcon(lucide.Image)},
				}},
				{Key: "readme", Label: "README.md", Description: "Project documentation", Leading: treeIcon(lucide.FileText)},
			},
		},
		{Key: "settings", Label: "Settings", Leading: treeIcon(lucide.Settings)},
		{Key: "archive", Label: "Archive", Description: "Read-only files", Leading: treeIcon(lucide.Archive), Disabled: true},
	}
}

func workspaceItems() []ui.TreeItem {
	return []ui.TreeItem{
		{Key: "app", Label: "Application", Description: "Desktop client", Leading: treeIcon(lucide.Package), Trailing: ui.Chip("12").Size(ui.ChipSmall), Children: []ui.TreeItem{
			{Key: "source", Label: "Source", Leading: treeIcon(lucide.Folder), Children: []ui.TreeItem{
				{Key: "main", Label: "main.go", Leading: treeIcon(lucide.FileCode)},
				{Key: "view", Label: "view.go", Leading: treeIcon(lucide.FileCode)},
			}},
			{Key: "database", Label: "Database", Leading: treeIcon(lucide.Database)},
		}},
		{Key: "private", Label: "Private", Description: "Restricted workspace", Leading: treeIcon(lucide.Lock), Disabled: true},
	}
}

func actionItems() []ui.TreeItem {
	return []ui.TreeItem{
		{Key: "build", Label: "Build", Leading: treeIcon(lucide.Package), Children: []ui.TreeItem{
			{Key: "test", Label: "Run tests", Leading: treeIcon(lucide.TestTube)},
			{Key: "compile", Label: "Compile application", Leading: treeIcon(lucide.Code)},
		}},
	}
}

func packageItems() []ui.TreeItem {
	items := make([]ui.TreeItem, 14)
	for index := range items {
		items[index] = ui.TreeItem{
			Key:      fmt.Sprintf("package-%d", index+1),
			Label:    fmt.Sprintf("Package %02d", index+1),
			Leading:  treeIcon(lucide.Package),
			Trailing: ui.Text(fmt.Sprintf("v1.%d", index)).Size(12),
		}
	}
	return items
}

func main() {
	ui.Run(
		Model{
			Selected:        "tree",
			Expanded:        []string{"flowui", "components"},
			SurfaceSelected: "main",
			SurfaceExpanded: []string{"app", "source"},
		},
		Update,
		View,
		ui.Title("FlowUI Tree"),
		ui.Size(900, 820),
	)
}
