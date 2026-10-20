package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/tree"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Items           []tree.Item
	Selected        string
	Checked         []string
	HalfChecked     []string
	MultiSelected   []string
	Expanded        []string
	AsyncItems      []tree.Item
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
	Tree            string
	Selected        string
	SelectedKeys    []string
	HalfCheckedKeys []string
	SetSelection    bool
	SetChecked      bool
	Expanded        []string
	Action          string
	Connectors      bool
	SetConnectors   bool
	Dashed          bool
	SetDashed       bool
	Drop            tree.DropEvent
	LoadChildren    string
	ChildrenLoaded  string
	RenameKey       string
	RenameLabel     string
	RenameTree      string
	RenameRequest   string
	ContextTree     string
	ContextKey      string
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
	if msg.SetChecked {
		model.Checked = append([]string(nil), msg.SelectedKeys...)
		model.HalfChecked = append([]string(nil), msg.HalfCheckedKeys...)
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
		if msg.Drop.Position == tree.DropInside && !containsKey(model.Expanded, msg.Drop.TargetKey) {
			model.Expanded = append(model.Expanded, msg.Drop.TargetKey)
		}
	}
	if msg.LoadChildren != "" && setTreeChildrenState(model.AsyncItems, msg.LoadChildren, tree.ChildrenLoading, "") {
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
	guideStyle := tree.GuideSolid
	if model.Dashed {
		guideStyle = tree.GuideDashed
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

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Tree").Size(24),
				text.New(status).Size(14),
				layout.Divider(),
				section("File browser",
					layout.Box(controlledTree("files", model.Selected, model.Expanded, items, send).
						ExpandOnRowClick(true).
						RequestRename(fileRename, model.RenameRevision).
						ContextMenu(treeContextMenu("files", model.ContextKey, model.ContextTree == "files" && model.ContextKey == model.Selected, send)).
						OnContextMenu(func(key string) { send(Msg{ContextTree: "files", ContextKey: key}) }).
						OnRename(func(key, label string) { send(Msg{RenameKey: key, RenameLabel: label}) })).
						Style(style.New().Width(520)),
				),
				section("Checkable",
					layout.Box(controlledCheckTree("checks", model.Checked, model.HalfChecked, model.Expanded, items, send)).
						Style(style.New().Width(520)),
				),
				section("Compact file tree",
					layout.Column(
						switchcontrol.SwitchGroup(
							switchcontrol.New("Branch connectors").Key("tree-connectors").Checked(model.Connectors).
								Size(switchcontrol.SwitchSmall).
								OnChange(func(value bool) { send(Msg{Connectors: value, SetConnectors: true}) }),
							switchcontrol.New("Dashed guides").Key("tree-dashed").Checked(model.Dashed).
								Size(switchcontrol.SwitchSmall).
								OnChange(func(value bool) { send(Msg{Dashed: value, SetDashed: true}) }),
						).Horizontal(),
						layout.Box(controlledMultiTree("compact", model.MultiSelected, model.Expanded, items, send).
							Size(tree.SizeSmall).
							RequestRename(compactRename, model.RenameRevision).
							ContextMenu(treeContextMenu("compact", model.ContextKey, model.ContextTree == "compact" && containsKey(model.MultiSelected, model.ContextKey), send)).
							OnContextMenu(func(key string) { send(Msg{ContextTree: "compact", ContextKey: key}) }).
							ExpandOnRowClick(true).
							Guides(true).
							GuideConnectors(model.Connectors).
							GuideStyle(guideStyle).
							CanDrop(func(event tree.DropEvent) bool {
								return event.TargetKey != "flowui" || event.Position != tree.DropBefore
							}).
							OnDragStart(func(event tree.DragEvent) {
								send(Msg{Action: fmt.Sprintf("Dragging: %d item(s)", len(event.SourceKeys))})
							}).
							OnDragEnter(func(event tree.DragEvent) {
								send(Msg{Action: fmt.Sprintf("Drop %s %s", treeDropPositionText(event.Position), event.TargetKey)})
							}).
							OnDrop(func(event tree.DropEvent) { send(Msg{Drop: event}) })).
							Style(style.New().Width(520)),
					).Gap(8),
				),
				section("Async children",
					layout.Box(controlledTree("async", model.AsyncSelected, model.AsyncExpanded, model.AsyncItems, send).
						OnLoadChildrenEvent(func(event tree.LoadEvent) { send(Msg{LoadChildren: event.Key}) })).
						Style(style.New().Width(520)),
				),
				section("Flat data",
					layout.Box(tree.New("simple", "", tree.ItemsFromSimple(simpleTreeItems())).
						ExpandedKeys([]string{"simple-flowui", "simple-ui"}).
						Guides(true)).
						Style(style.New().Width(520)),
				),
				section("Custom filter",
					layout.Box(tree.New("filtered", "", fileItems()).
						FilterFunc(func(item tree.Item) bool { return item.Key == "tree" }).
						FilterVersion(1).
						ExpandedKeys([]string{"flowui", "components"}).
						Guides(true)).
						Style(style.New().Width(520)),
				),
				section("Surface",
					layout.Box(controlledTree("surface", model.SurfaceSelected, model.SurfaceExpanded, workspaceItems(), send).
						Variant(tree.VariantSurface)).
						Style(style.New().Width(520)),
				),
				section("Custom rows",
					layout.Box(tree.New("custom", "", customTreeItems()).
						ExpandedKeys([]string{"design"})).
						Style(style.New().Width(520)),
				),
				section("Action only",
					layout.Box(tree.New("actions", "", actionItems()).
						SelectionMode(tree.SelectionNone).
						ExpandedKeys([]string{"build"}).
						OnAction(func(key string) { send(Msg{Action: key}) })).
						Style(style.New().Width(520)),
				),
				section("Scrollable",
					layout.Box(tree.New("packages", "", packageItems()).
						MaxHeight(180).
						DisabledKeys([]string{"package-7"})).
						Style(style.New().Width(520)),
				),
			).Gap(20),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func controlledTree(key, selected string, expanded []string, items []tree.Item, send ui.Send[Msg]) tree.Widget {
	return tree.New(key, selected, items).
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

func controlledMultiTree(key string, selected, expanded []string, items []tree.Item, send ui.Send[Msg]) tree.Widget {
	return tree.New(key, "", items).
		SelectionMode(tree.SelectionMultiple).
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

func controlledCheckTree(key string, checked, halfChecked, expanded []string, items []tree.Item, send ui.Send[Msg]) tree.Widget {
	return tree.New(key, "", items).
		Checkable(true).
		CheckedKeys(checked).
		HalfCheckedKeys(halfChecked).
		ExpandedKeys(expanded).
		OnCheckedChange(func(keys, halfKeys []string) {
			send(Msg{SelectedKeys: keys, HalfCheckedKeys: halfKeys, SetChecked: true})
		}).
		OnExpandedChange(func(keys []string) {
			send(Msg{Tree: key, Expanded: keys})
		})
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func treeIcon(data lucide.Data) ui.Widget {
	return icon.New(data).Size(16)
}

func treeContextMenu(treeKey, target string, renameEnabled bool, send ui.Send[Msg]) menu.Widget {
	return menu.New("file-actions", []menu.Item{
		{Key: "open", Label: "Open", Leading: icon.New(lucide.FolderOpen).Size(16)},
		{Key: "copy", Label: "Copy", Leading: icon.New(lucide.Copy).Size(16)},
		{Key: "rename", Label: "Rename", Shortcut: "F2", Leading: icon.New(lucide.Pencil).Size(16), Disabled: !renameEnabled},
		menu.Separator(),
		{Key: "delete", Label: "Delete", Variant: menu.ItemDanger, Leading: icon.New(lucide.Trash).Size(16)},
	}).Compact(true).OnActionEvent(func(event menu.ActionEvent) {
		if event.Key == "rename" {
			send(Msg{RenameTree: treeKey, RenameRequest: target})
			return
		}
		send(Msg{Action: fmt.Sprintf("%s: %s", event.Key, target)})
	})
}

func customTreeItems() []tree.Item {
	return []tree.Item{
		{
			Key:              "design",
			Label:            "Design",
			Switcher:         icon.New(lucide.ChevronRight).Size(14),
			ExpandedSwitcher: icon.New(lucide.ChevronDown).Size(14),
			Children: []tree.Item{
				{
					Key:   "figma",
					Label: "Figma",
					Content: layout.Row(
						icon.New(lucide.FilePenLine).Size(16),
						layout.Column(text.New("Figma"), description.New("Shared design file")).Gap(2),
					).Gap(8).AlignMiddle(),
				},
			},
		},
	}
}

func moveTreeItems(items *[]tree.Item, event tree.DropEvent) bool {
	sources := treeDropSources(event)
	if len(sources) == 0 {
		return false
	}
	_, targetOK := findTreeItem(*items, event.TargetKey)
	validPosition := event.Position == tree.DropBefore || event.Position == tree.DropInside || event.Position == tree.DropAfter
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
	moved := make([]tree.Item, 0, len(sources))
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

func treeDropSources(event tree.DropEvent) []string {
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

func cloneTreeItems(items []tree.Item) []tree.Item {
	result := append([]tree.Item(nil), items...)
	for index := range result {
		result[index].Children = cloneTreeItems(result[index].Children)
	}
	return result
}

func findTreeItem(items []tree.Item, key string) (tree.Item, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
		if found, ok := findTreeItem(item.Children, key); ok {
			return found, true
		}
	}
	return tree.Item{}, false
}

func takeTreeItem(items *[]tree.Item, key string) (tree.Item, bool) {
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
	return tree.Item{}, false
}

func insertTreeItems(items *[]tree.Item, target string, position tree.DropPosition, moved []tree.Item) bool {
	for index := range *items {
		if (*items)[index].Key == target {
			switch position {
			case tree.DropInside:
				(*items)[index].Children = append((*items)[index].Children, moved...)
			case tree.DropBefore:
				next := make([]tree.Item, 0, len(*items)+len(moved))
				next = append(next, (*items)[:index]...)
				next = append(next, moved...)
				next = append(next, (*items)[index:]...)
				*items = next
			case tree.DropAfter:
				next := make([]tree.Item, 0, len(*items)+len(moved))
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

func setTreeChildrenState(items []tree.Item, key string, state tree.ChildrenState, message string) bool {
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

func setTreeChildren(items []tree.Item, key string, children []tree.Item) bool {
	for index := range items {
		if items[index].Key == key {
			items[index].Children = children
			items[index].ChildrenState = tree.ChildrenLoaded
			items[index].LoadError = ""
			return true
		}
		if setTreeChildren(items[index].Children, key, children) {
			return true
		}
	}
	return false
}

func renameTreeItem(items []tree.Item, key, label string) bool {
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

func loadedTreeChildren(parent string) []tree.Item {
	return []tree.Item{
		{Key: parent + "-report", Label: "report.json", Leading: treeIcon(lucide.FileCode), Renamable: true},
		{Key: parent + "-logs", Label: "logs", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true},
	}
}

func asyncItems() []tree.Item {
	return []tree.Item{
		{Key: "remote", Label: "Remote workspace", Leading: treeIcon(lucide.Cloud), ChildrenState: tree.ChildrenUnloaded, AcceptsChildren: true},
		{Key: "backup", Label: "Backup", Leading: treeIcon(lucide.DatabaseBackup), ChildrenState: tree.ChildrenError, LoadError: "Connection failed", AcceptsChildren: true},
	}
}

func simpleTreeItems() []tree.SimpleItem {
	return []tree.SimpleItem{
		tree.NewSimpleItem("simple-flowui", "", "FlowUI"),
		{Key: "simple-ui", ParentKey: "simple-flowui", Label: "ui", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen)},
		{Key: "simple-tree", ParentKey: "simple-ui", Label: "tree.go", Leading: treeIcon(lucide.FileCode)},
		{Key: "simple-readme", ParentKey: "simple-flowui", Label: "README.md", Leading: treeIcon(lucide.FileText)},
	}
}

func containsKey(keys []string, key string) bool {
	return slices.Contains(keys, key)
}

func treeDropPositionText(position tree.DropPosition) string {
	switch position {
	case tree.DropBefore:
		return "before"
	case tree.DropAfter:
		return "after"
	default:
		return "inside"
	}
}

func fileItems() []tree.Item {
	return []tree.Item{
		{
			Key: "flowui", Label: "FlowUI", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true,
			Children: []tree.Item{
				{
					Key: "components", Label: "components", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true,
					Children: []tree.Item{
						{Key: "button", Label: "button.go", Leading: treeIcon(lucide.FileCode), Renamable: true},
						{Key: "tree", Label: "tree.go", Leading: treeIcon(lucide.FileCode), Trailing: chip.New("New").Color(chip.ColorAccent).Size(chip.SizeSmall), Renamable: true},
					},
				},
				{Key: "assets", Label: "assets", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Renamable: true, Children: []tree.Item{
					{Key: "logo", Label: "logo.png", Leading: treeIcon(lucide.Image), Renamable: true},
				}},
				{Key: "readme", Label: "README.md", Description: "Project documentation", Leading: treeIcon(lucide.FileText), Renamable: true},
			},
		},
		{Key: "settings", Label: "Settings", Leading: treeIcon(lucide.Settings), Renamable: true},
		{Key: "archive", Label: "Archive", Description: "Read-only files", Leading: treeIcon(lucide.Archive), Disabled: true},
	}
}

func workspaceItems() []tree.Item {
	return []tree.Item{
		{Key: "app", Label: "Application", Description: "Desktop client", Leading: treeIcon(lucide.Package), Trailing: chip.New("12").Size(chip.SizeSmall), Children: []tree.Item{
			{Key: "source", Label: "Source", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []tree.Item{
				{Key: "main", Label: "main.go", Leading: treeIcon(lucide.FileCode)},
				{Key: "view", Label: "view.go", Leading: treeIcon(lucide.FileCode)},
			}},
			{Key: "database", Label: "Database", Leading: treeIcon(lucide.Database)},
		}},
		{Key: "private", Label: "Private", Description: "Restricted workspace", Leading: treeIcon(lucide.Lock), Disabled: true},
	}
}

func actionItems() []tree.Item {
	return []tree.Item{
		{Key: "build", Label: "Build", Leading: treeIcon(lucide.Package), Children: []tree.Item{
			{Key: "test", Label: "Run tests", Leading: treeIcon(lucide.TestTube)},
			{Key: "compile", Label: "Compile application", Leading: treeIcon(lucide.Code)},
		}},
	}
}

func packageItems() []tree.Item {
	items := make([]tree.Item, 14)
	for index := range items {
		items[index] = tree.Item{
			Key:      fmt.Sprintf("package-%d", index+1),
			Label:    fmt.Sprintf("Package %02d", index+1),
			Leading:  treeIcon(lucide.Package),
			Trailing: text.New(fmt.Sprintf("v1.%d", index)).Size(12),
		}
	}
	return items
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Items:           fileItems(),
		Selected:        "tree",
		MultiSelected:   []string{"button", "tree"},
		Expanded:        []string{"flowui", "components"},
		AsyncItems:      asyncItems(),
		SurfaceSelected: "main",
		SurfaceExpanded: []string{"app", "source"},
	},
		Update, View), window.Title("FlowUI Tree"),
		window.Size(900, 820),
		window.Theme(base.Light()),
	)
}
