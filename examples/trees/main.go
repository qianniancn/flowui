package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Items           []ui.TreeItem
	Selected        string
	Expanded        []string
	SurfaceSelected string
	SurfaceExpanded []string
	LastAction      string
	Connectors      bool
	Dashed          bool
}

type Msg struct {
	Tree          string
	Selected      string
	Expanded      []string
	Action        string
	Connectors    bool
	SetConnectors bool
	Dashed        bool
	SetDashed     bool
	Drop          ui.TreeDropEvent
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
		model.LastAction = fmt.Sprintf("Activated: %s", msg.Action)
	}
	if msg.SetConnectors {
		model.Connectors = msg.Connectors
	}
	if msg.SetDashed {
		model.Dashed = msg.Dashed
	}
	if msg.Drop.SourceKey != "" && moveTreeItem(&model.Items, msg.Drop) {
		model.LastAction = fmt.Sprintf("Moved: %s", msg.Drop.SourceKey)
		if msg.Drop.Position == ui.TreeDropInside && !containsKey(model.Expanded, msg.Drop.TargetKey) {
			model.Expanded = append(model.Expanded, msg.Drop.TargetKey)
		}
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "Select a file or folder"
	if model.LastAction != "" {
		status = model.LastAction
	}
	guideStyle := ui.TreeGuideSolid
	if model.Dashed {
		guideStyle = ui.TreeGuideDashed
	}
	items := model.Items
	if len(items) == 0 {
		items = fileItems()
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("trees",
				ui.Column(
					ui.Text("FlowUI Tree").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					section("File browser",
						ui.Box(controlledTree("files", model.Selected, model.Expanded, items, send)).
							Width(520),
					),
					section("Compact file tree",
						ui.Column(
							ui.SwitchGroup(
								ui.Switch("tree-connectors", model.Connectors, "Branch connectors").
									Size(ui.SwitchSmall).
									OnChange(func(value bool) { send(Msg{Connectors: value, SetConnectors: true}) }),
								ui.Switch("tree-dashed", model.Dashed, "Dashed guides").
									Size(ui.SwitchSmall).
									OnChange(func(value bool) { send(Msg{Dashed: value, SetDashed: true}) }),
							).Horizontal(),
							ui.Box(controlledTree("compact", model.Selected, model.Expanded, items, send).
								Size(ui.TreeSmall).
								Guides(true).
								GuideConnectors(model.Connectors).
								GuideStyle(guideStyle).
								OnDrop(func(event ui.TreeDropEvent) { send(Msg{Drop: event}) })).
								Width(520),
						).Gap(8),
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

func moveTreeItem(items *[]ui.TreeItem, event ui.TreeDropEvent) bool {
	source, sourceOK := findTreeItem(*items, event.SourceKey)
	_, targetOK := findTreeItem(*items, event.TargetKey)
	validPosition := event.Position == ui.TreeDropBefore || event.Position == ui.TreeDropInside || event.Position == ui.TreeDropAfter
	if !sourceOK || !targetOK || !validPosition || event.SourceKey == event.TargetKey {
		return false
	}
	if _, descendant := findTreeItem(source.Children, event.TargetKey); descendant {
		return false
	}
	item, ok := takeTreeItem(items, event.SourceKey)
	return ok && insertTreeItem(items, event.TargetKey, event.Position, item)
}

func findTreeItem(items []ui.TreeItem, key string) (ui.TreeItem, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
		if found, ok := findTreeItem(item.Children, key); ok {
			return found, true
		}
	}
	return ui.TreeItem{}, false
}

func takeTreeItem(items *[]ui.TreeItem, key string) (ui.TreeItem, bool) {
	for index := range *items {
		if (*items)[index].Key == key {
			item := (*items)[index]
			*items = append((*items)[:index], (*items)[index+1:]...)
			return item, true
		}
		if item, ok := takeTreeItem(&(*items)[index].Children, key); ok {
			return item, true
		}
	}
	return ui.TreeItem{}, false
}

func insertTreeItem(items *[]ui.TreeItem, target string, position ui.TreeDropPosition, item ui.TreeItem) bool {
	for index := range *items {
		if (*items)[index].Key == target {
			switch position {
			case ui.TreeDropInside:
				(*items)[index].Children = append((*items)[index].Children, item)
			case ui.TreeDropBefore:
				*items = append(*items, ui.TreeItem{})
				copy((*items)[index+1:], (*items)[index:])
				(*items)[index] = item
			case ui.TreeDropAfter:
				index++
				*items = append(*items, ui.TreeItem{})
				copy((*items)[index+1:], (*items)[index:])
				(*items)[index] = item
			}
			return true
		}
		if insertTreeItem(&(*items)[index].Children, target, position, item) {
			return true
		}
	}
	return false
}

func containsKey(keys []string, key string) bool {
	for _, current := range keys {
		if current == key {
			return true
		}
	}
	return false
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
			Items:           fileItems(),
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
