package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Items           []ui.TreeItem
	Selected        string
	MultiSelected   []string
	Expanded        []string
	AsyncItems      []ui.TreeItem
	AsyncSelected   string
	AsyncExpanded   []string
	SurfaceSelected string
	SurfaceExpanded []string
	ContextTree     string
	ContextKey      string
	RenameTree      string
	RenameTarget    string
	RenameRevision  uint64
	LastAction      string
	Connectors      bool
	Dashed          bool
}

type Msg struct {
	Tree           string
	Selected       string
	SelectedKeys   []string
	SetSelection   bool
	Expanded       []string
	Action         string
	Connectors     bool
	SetConnectors  bool
	Dashed         bool
	SetDashed      bool
	Drop           ui.TreeDropEvent
	LoadChildren   string
	ChildrenLoaded string
	RenameKey      string
	RenameLabel    string
	RenameTree     string
	RenameRequest  string
	ContextTree    string
	ContextKey     string
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Expanded != nil {
		if msg.Tree == "async" {
			model.AsyncExpanded = append([]string(nil), msg.Expanded...)
		} else if msg.Tree == "surface" {
			model.SurfaceExpanded = append([]string(nil), msg.Expanded...)
		} else {
			model.Expanded = append([]string(nil), msg.Expanded...)
		}
	}
	if msg.Selected != "" {
		if msg.Tree == "async" {
			model.AsyncSelected = msg.Selected
		} else if msg.Tree == "surface" {
			model.SurfaceSelected = msg.Selected
		} else {
			model.Selected = msg.Selected
		}
	}
	if msg.SetSelection {
		model.MultiSelected = append([]string(nil), msg.SelectedKeys...)
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
	if msg.Drop.SourceKey != "" && moveTreeItems(&model.Items, msg.Drop) {
		model.LastAction = fmt.Sprintf("Moved: %d item(s)", len(treeDropSources(msg.Drop)))
		if msg.Drop.Position == ui.TreeDropInside && !containsKey(model.Expanded, msg.Drop.TargetKey) {
			model.Expanded = append(model.Expanded, msg.Drop.TargetKey)
		}
	}
	if msg.LoadChildren != "" && setTreeChildrenState(model.AsyncItems, msg.LoadChildren, ui.TreeChildrenLoading, "") {
		return loadTreeChildren(msg.LoadChildren)
	}
	if msg.ChildrenLoaded != "" {
		setTreeChildren(model.AsyncItems, msg.ChildrenLoaded, loadedTreeChildren(msg.ChildrenLoaded))
	}
	if msg.RenameKey != "" && renameTreeItem(model.Items, msg.RenameKey, msg.RenameLabel) {
		model.LastAction = fmt.Sprintf("Renamed: %s", msg.RenameLabel)
	}
	if msg.ContextKey != "" {
		model.ContextTree = msg.ContextTree
		model.ContextKey = msg.ContextKey
		model.LastAction = fmt.Sprintf("Context menu: %s", msg.ContextKey)
	}
	if msg.RenameRequest != "" {
		model.RenameTree = msg.RenameTree
		model.RenameTarget = msg.RenameRequest
		model.RenameRevision++
	}
	return nil
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
	fileRename, compactRename := "", ""
	if model.RenameTree == "files" {
		fileRename = model.RenameTarget
	} else if model.RenameTree == "compact" {
		compactRename = model.RenameTarget
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("trees",
				ui.Column(
					ui.Text("FlowUI Tree").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					section("File browser",
						ui.Box(controlledTree("files", model.Selected, model.Expanded, items, send).
							RequestRename(fileRename, model.RenameRevision).
							ContextMenu(treeContextMenu("files", model.ContextKey, model.ContextTree == "files" && model.ContextKey == model.Selected, send)).
							OnContextMenu(func(key string) { send(Msg{ContextTree: "files", ContextKey: key}) }).
							OnRename(func(key, label string) { send(Msg{RenameKey: key, RenameLabel: label}) })).
							Style(ui.Width(520)),
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
							ui.Box(controlledMultiTree("compact", model.MultiSelected, model.Expanded, items, send).
								Size(ui.TreeSmall).
								RequestRename(compactRename, model.RenameRevision).
								ContextMenu(treeContextMenu("compact", model.ContextKey, model.ContextTree == "compact" && containsKey(model.MultiSelected, model.ContextKey), send)).
								OnContextMenu(func(key string) { send(Msg{ContextTree: "compact", ContextKey: key}) }).
								ExpandOnRowClick(true).
								Guides(true).
								GuideConnectors(model.Connectors).
								GuideStyle(guideStyle).
								CanDrop(func(event ui.TreeDropEvent) bool {
									return event.TargetKey != "flowui" || event.Position != ui.TreeDropBefore
								}).
								OnDrop(func(event ui.TreeDropEvent) { send(Msg{Drop: event}) })).
								Style(ui.Width(520)),
						).Gap(8),
					),
					section("Async children",
						ui.Box(controlledTree("async", model.AsyncSelected, model.AsyncExpanded, model.AsyncItems, send).
							OnLoadChildren(func(key string) { send(Msg{LoadChildren: key}) })).
							Style(ui.Width(520)),
					),
					section("Surface",
						ui.Box(controlledTree("surface", model.SurfaceSelected, model.SurfaceExpanded, workspaceItems(), send).
							Variant(ui.TreeSurface)).
							Style(ui.Width(520)),
					),
					section("Action only",
						ui.Box(ui.Tree("actions", "", actionItems()).
							SelectionMode(ui.TreeSelectionNone).
							ExpandedKeys([]string{"build"}).
							OnAction(func(key string) { send(Msg{Action: key}) })).
							Style(ui.Width(520)),
					),
					section("Scrollable",
						ui.Box(ui.Tree("packages", "", packageItems()).
							MaxHeight(180).
							DisabledKeys([]string{"package-7"})).
							Style(ui.Width(520)),
					),
				).Gap(20),
			).Vertical(),
		).Style(ui.FillWidth().MaxWidth(760).Padding(24)),
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

func controlledMultiTree(key string, selected, expanded []string, items []ui.TreeItem, send ui.Send[Msg]) ui.TreeWidget {
	return ui.Tree(key, "", items).
		SelectionMode(ui.TreeSelectionMultiple).
		SelectedKeys(selected).
		ExpandedKeys(expanded).
		OnSelectionChange(func(keys []string) {
			send(Msg{SelectedKeys: keys, SetSelection: true})
		}).
		OnExpandedChange(func(keys []string) {
			send(Msg{Tree: key, Expanded: keys})
		}).
		OnRename(func(itemKey, label string) {
			send(Msg{RenameKey: itemKey, RenameLabel: label})
		})
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func treeIcon(data lucide.Data) ui.Widget {
	return ui.Icon(data).Size(16)
}

func treeContextMenu(treeKey, target string, renameEnabled bool, send ui.Send[Msg]) ui.MenuWidget {
	return ui.Menu("file-actions", []ui.MenuItem{
		{Key: "open", Label: "Open", Leading: ui.Icon(lucide.FolderOpen).Size(16)},
		{Key: "copy", Label: "Copy", Leading: ui.Icon(lucide.Copy).Size(16)},
		{Key: "rename", Label: "Rename", Shortcut: "F2", Leading: ui.Icon(lucide.Pencil).Size(16), Disabled: !renameEnabled},
		ui.MenuSeparator(),
		{Key: "delete", Label: "Delete", Variant: ui.MenuItemDanger, Leading: ui.Icon(lucide.Trash2).Size(16)},
	}).Compact(true).Style(ui.Width(168)).OnAction(func(action string) {
		if action == "rename" {
			send(Msg{RenameTree: treeKey, RenameRequest: target})
			return
		}
		send(Msg{Action: fmt.Sprintf("%s: %s", action, target)})
	})
}

func moveTreeItems(items *[]ui.TreeItem, event ui.TreeDropEvent) bool {
	sources := treeDropSources(event)
	if len(sources) == 0 {
		return false
	}
	_, targetOK := findTreeItem(*items, event.TargetKey)
	validPosition := event.Position == ui.TreeDropBefore || event.Position == ui.TreeDropInside || event.Position == ui.TreeDropAfter
	if !targetOK || !validPosition {
		return false
	}
	for _, sourceKey := range sources {
		source, sourceOK := findTreeItem(*items, sourceKey)
		if !sourceOK || sourceKey == event.TargetKey {
			return false
		}
		if _, descendant := findTreeItem(source.Children, event.TargetKey); descendant {
			return false
		}
	}
	working := cloneTreeItems(*items)
	moved := make([]ui.TreeItem, 0, len(sources))
	for _, sourceKey := range sources {
		item, ok := takeTreeItem(&working, sourceKey)
		if !ok {
			return false
		}
		moved = append(moved, item)
	}
	if !insertTreeItems(&working, event.TargetKey, event.Position, moved) {
		return false
	}
	*items = working
	return true
}

func treeDropSources(event ui.TreeDropEvent) []string {
	keys := event.SourceKeys
	if len(keys) == 0 && event.SourceKey != "" {
		keys = []string{event.SourceKey}
	}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		if key != "" && !containsKey(result, key) {
			result = append(result, key)
		}
	}
	return result
}

func cloneTreeItems(items []ui.TreeItem) []ui.TreeItem {
	result := append([]ui.TreeItem(nil), items...)
	for index := range result {
		result[index].Children = cloneTreeItems(result[index].Children)
	}
	return result
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

func insertTreeItems(items *[]ui.TreeItem, target string, position ui.TreeDropPosition, moved []ui.TreeItem) bool {
	for index := range *items {
		if (*items)[index].Key == target {
			switch position {
			case ui.TreeDropInside:
				(*items)[index].Children = append((*items)[index].Children, moved...)
			case ui.TreeDropBefore:
				next := make([]ui.TreeItem, 0, len(*items)+len(moved))
				next = append(next, (*items)[:index]...)
				next = append(next, moved...)
				next = append(next, (*items)[index:]...)
				*items = next
			case ui.TreeDropAfter:
				next := make([]ui.TreeItem, 0, len(*items)+len(moved))
				next = append(next, (*items)[:index+1]...)
				next = append(next, moved...)
				next = append(next, (*items)[index+1:]...)
				*items = next
			}
			return true
		}
		if insertTreeItems(&(*items)[index].Children, target, position, moved) {
			return true
		}
	}
	return false
}

func setTreeChildrenState(items []ui.TreeItem, key string, state ui.TreeChildrenState, message string) bool {
	for index := range items {
		if items[index].Key == key {
			items[index].ChildrenState = state
			items[index].LoadError = message
			return true
		}
		if setTreeChildrenState(items[index].Children, key, state, message) {
			return true
		}
	}
	return false
}

func setTreeChildren(items []ui.TreeItem, key string, children []ui.TreeItem) bool {
	for index := range items {
		if items[index].Key == key {
			items[index].Children = children
			items[index].ChildrenState = ui.TreeChildrenLoaded
			items[index].LoadError = ""
			return true
		}
		if setTreeChildren(items[index].Children, key, children) {
			return true
		}
	}
	return false
}

func renameTreeItem(items []ui.TreeItem, key, label string) bool {
	for index := range items {
		if items[index].Key == key {
			items[index].Label = label
			return true
		}
		if renameTreeItem(items[index].Children, key, label) {
			return true
		}
	}
	return false
}

func loadTreeChildren(key string) ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		timer := time.NewTimer(900 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			send(Msg{ChildrenLoaded: key})
			return nil
		}
	})
}

func loadedTreeChildren(parent string) []ui.TreeItem {
	return []ui.TreeItem{
		{Key: parent + "-report", Label: "report.json", Leading: treeIcon(lucide.FileCode), Renamable: true},
		{Key: parent + "-logs", Label: "logs", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true},
	}
}

func asyncItems() []ui.TreeItem {
	return []ui.TreeItem{
		{Key: "remote", Label: "Remote workspace", Leading: treeIcon(lucide.Cloud), ChildrenState: ui.TreeChildrenUnloaded, AcceptsChildren: true},
		{Key: "backup", Label: "Backup", Leading: treeIcon(lucide.DatabaseBackup), ChildrenState: ui.TreeChildrenError, LoadError: "Connection failed", AcceptsChildren: true},
	}
}

func containsKey(keys []string, key string) bool {
	return slices.Contains(keys, key)
}

func fileItems() []ui.TreeItem {
	return []ui.TreeItem{
		{
			Key: "flowui", Label: "FlowUI", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true,
			Children: []ui.TreeItem{
				{
					Key: "components", Label: "components", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true,
					Children: []ui.TreeItem{
						{Key: "button", Label: "button.go", Leading: treeIcon(lucide.FileCode), Renamable: true},
						{Key: "tree", Label: "tree.go", Leading: treeIcon(lucide.FileCode), Trailing: ui.Chip("New").Color(ui.ChipAccent).Size(ui.ChipSmall), Renamable: true},
					},
				},
				{Key: "assets", Label: "assets", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true, Children: []ui.TreeItem{
					{Key: "logo", Label: "logo.png", Leading: treeIcon(lucide.Image), Renamable: true},
				}},
				{Key: "readme", Label: "README.md", Description: "Project documentation", Leading: treeIcon(lucide.FileText), Renamable: true},
			},
		},
		{Key: "settings", Label: "Settings", Leading: treeIcon(lucide.Settings), Renamable: true},
		{Key: "archive", Label: "Archive", Description: "Read-only files", Leading: treeIcon(lucide.Archive), Disabled: true},
	}
}

func workspaceItems() []ui.TreeItem {
	return []ui.TreeItem{
		{Key: "app", Label: "Application", Description: "Desktop client", Leading: treeIcon(lucide.Package), Trailing: ui.Chip("12").Size(ui.ChipSmall), Children: []ui.TreeItem{
			{Key: "source", Label: "Source", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []ui.TreeItem{
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
	ui.RunCmd(
		Model{
			Items:           fileItems(),
			Selected:        "tree",
			MultiSelected:   []string{"button", "tree"},
			Expanded:        []string{"flowui", "components"},
			AsyncItems:      asyncItems(),
			SurfaceSelected: "main",
			SurfaceExpanded: []string{"app", "source"},
		},
		Update,
		View,
		ui.Title("FlowUI Tree"),
		ui.Size(900, 820),
	)
}
