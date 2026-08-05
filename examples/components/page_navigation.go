package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

func tabsPaginationPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Tabs & pagination",
		demoSection{Title: "Tabs", Content: demoPanel(
			ui.Tabs("catalog-tabs", model.TabValue, []ui.TabItem{
				{Key: "preview", Label: "Preview", Panel: tabPanel("Preview")},
				{Key: "code", Label: "Code", Panel: tabPanel("Code")},
				{Key: "accessibility", Label: "Accessibility", Panel: tabPanel("Accessibility")},
			}).Separators(true).OnChange(func(value string) { send(func(model *Model) { model.TabValue = value }) }),
		)},
		demoSection{Title: "Pagination", Content: demoPanel(
			ui.Pagination("catalog-pagination", model.PaginationPage, 12).
				Summary(ui.Text("61 to 72 of 144 results").Size(12)).
				OnChange(func(value int) { send(func(model *Model) { model.PaginationPage = value }) }),
		)},
	)
}

func sidebarTreePage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	guideStyle := ui.TreeGuideSolid
	if model.TreeDashed {
		guideStyle = ui.TreeGuideDashed
	}
	return demoPage("Sidebar & tree",
		demoSection{Title: "Sidebar", Content: demoPanel(
			ui.Box(ui.Sidebar("catalog-sidebar-demo", model.SidebarValue, []ui.SidebarItem{
				{Key: "overview", Label: "Overview", Leading: ui.Icon(lucide.LayoutDashboard).Size(16)},
				{Key: "projects", Label: "Projects", Leading: ui.Icon(lucide.FolderKanban).Size(16), Trailing: ui.Chip("8").Size(ui.ChipSmall)},
				{Key: "settings", Label: "Settings", Leading: ui.Icon(lucide.Settings).Size(16)},
			}).OnChange(func(value string) { send(func(model *Model) { model.SidebarValue = value }) })).Style(ui.Width(320).Height(260)),
		)},
		demoSection{Title: "Tree", Content: demoPanel(ui.Column(
			ui.SwitchGroup(
				ui.Switch("catalog-tree-connectors", model.TreeConnectors, "Branch connectors").
					Size(ui.SwitchSmall).
					OnChange(func(value bool) { send(func(model *Model) { model.TreeConnectors = value }) }),
				ui.Switch("catalog-tree-dashed", model.TreeDashed, "Dashed guides").
					Size(ui.SwitchSmall).
					OnChange(func(value bool) { send(func(model *Model) { model.TreeDashed = value }) }),
			).Horizontal(),
			ui.Box(ui.Tree("catalog-tree", model.TreeValue, catalogTreeItems).
				ExpandedKeys(model.TreeExpanded).
				Guides(true).
				GuideConnectors(model.TreeConnectors).
				GuideStyle(guideStyle).
				OnChange(func(value string) { send(func(model *Model) { model.TreeValue = value }) }).
				OnExpandedChange(func(values []string) {
					send(func(model *Model) { model.TreeExpanded = append([]string(nil), values...) })
				})).Style(ui.Width(440)),
		).Gap(8))},
		demoSection{Title: "Tree drag and drop", Content: demoPanel(ui.Column(
			ui.Box(ui.Tree("catalog-tree-drag", model.TreeDragValue, model.TreeDragItems).
				ExpandedKeys(model.TreeDragExpanded).
				Variant(ui.TreeSurface).
				Guides(true).
				GuideConnectors(true).
				GuideStyle(ui.TreeGuideDashed).
				MaxHeight(300).
				OnChange(func(value string) {
					send(func(model *Model) { model.TreeDragValue = value })
				}).
				OnExpandedChange(func(values []string) {
					send(func(model *Model) { model.TreeDragExpanded = append([]string(nil), values...) })
				}).
				CanDrop(func(event ui.TreeDropEvent) bool {
					return catalogTreeCanDrop(model.TreeDragItems, event)
				}).
				OnDrop(func(event ui.TreeDropEvent) {
					send(func(model *Model) {
						if !catalogTreeMove(&model.TreeDragItems, event) {
							return
						}
						model.TreeDropMessage = fmt.Sprintf("Moved %s %s %s", event.SourceKey, treeDropPositionLabel(event.Position), event.TargetKey)
						if event.Position == ui.TreeDropInside && !containsCatalogTreeKey(model.TreeDragExpanded, event.TargetKey) {
							model.TreeDragExpanded = append(model.TreeDragExpanded, event.TargetKey)
						}
					})
				}),
			).Style(ui.Width(440)),
			ui.Text(model.TreeDropMessage).Size(12),
		).Gap(10))},
	)
}

func menusPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	action := func(value string) { send(func(model *Model) { model.LastAction = value }) }
	menuItems := []ui.MenuItem{
		{Key: "new", Label: "New file", Shortcut: "Ctrl+N", Leading: ui.Icon(lucide.FilePlus).Size(16)},
		{Key: "open", Label: "Open", Shortcut: "Ctrl+O", Leading: ui.Icon(lucide.FolderOpen).Size(16)},
		ui.MenuSeparator(),
		{Key: "delete", Label: "Delete", Variant: ui.MenuItemDanger, Leading: ui.Icon(lucide.Trash2).Size(16)},
	}
	menuSections := []ui.MenuSection{
		{Title: "Create", Items: []ui.MenuItem{
			{Key: "new-file", Label: "New file", Leading: ui.Icon(lucide.FilePlus).Size(16)},
			{Key: "new-folder", Label: "New folder", Leading: ui.Icon(lucide.Folder).Size(16)},
		}},
		{Title: "Transfer", Items: []ui.MenuItem{
			{Key: "download", Label: "Download", Leading: ui.Icon(lucide.Download).Size(16)},
			{Key: "archive", Label: "Archive", Leading: ui.Icon(lucide.Archive).Size(16)},
		}},
	}
	return demoPage("Menus",
		demoSection{Title: "Menubar", Content: demoPanel(
			ui.Menubar("catalog-menubar", []ui.MenubarItem{
				ui.MenubarMenu("catalog-file-menu", "File", menuItems).OnActionEvent(func(event ui.MenuActionEvent) { action(event.Key) }),
				ui.MenubarMenu("catalog-edit-menu", "Edit", []ui.MenuItem{
					{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: ui.Icon(lucide.Copy).Size(16)},
					{Key: "paste", Label: "Paste", Shortcut: "Ctrl+V", Leading: ui.Icon(lucide.ClipboardPaste).Size(16)},
				}).OnActionEvent(func(event ui.MenuActionEvent) { action(event.Key) }),
			}).Alt("Catalog menu bar"),
		)},
		demoSection{Title: "Dropdown", Content: demoPanel(
			ui.Dropdown("catalog-dropdown",
				ui.Button("catalog-dropdown-trigger", ui.Row(ui.Text("Actions"), ui.Icon(lucide.ChevronDown).Size(14)).Gap(7)).Variant(ui.ButtonSecondary),
				[]ui.DropdownItem{
					{Key: "preview", Label: "Preview", Leading: ui.Icon(lucide.Eye).Size(16)},
					{Key: "download", Label: "Download", Leading: ui.Icon(lucide.Download).Size(16)},
					{Key: "archive", Label: "Archive", Leading: ui.Icon(lucide.Archive).Size(16)},
				},
			).OnActionEvent(func(event ui.DropdownActionEvent) { action(event.Key) }),
		)},
		demoSection{Title: "DropdownButton", Content: demoPanel(
			ui.DropdownButton("catalog-dropdown-button",
				ui.Button("catalog-dropdown-button-action", ui.Text("Create")),
				[]ui.DropdownItem{
					{Key: "project", Label: "New project", Leading: ui.Icon(lucide.FolderPlus).Size(16)},
					{Key: "file", Label: "New file", Leading: ui.Icon(lucide.FilePlus).Size(16)},
				},
			).Variant(ui.ButtonSecondary).
				OnClick(func() { action("Create") }).
				OnActionEvent(func(event ui.DropdownActionEvent) { action(event.Key) }),
		)},
		demoSection{Title: "DropdownSections & MenuSections", Content: demoPanel(demoRow(
			ui.DropdownSections("catalog-dropdown-sections",
				ui.Button("catalog-dropdown-sections-trigger", ui.Row(
					ui.Text("Sectioned actions"),
					ui.Icon(lucide.ChevronDown).Size(14),
				).Gap(7)).Variant(ui.ButtonSecondary),
				menuSections,
			).OnActionEvent(func(event ui.DropdownActionEvent) { action(event.Key) }),
			ui.Box(ui.MenuSections("catalog-menu-sections", menuSections).OnActionEvent(func(event ui.MenuActionEvent) { action(event.Key) })).Style(ui.Width(300)),
		))},
		demoSection{Title: "ContextMenu", Content: demoPanel(
			ui.ContextMenu("catalog-context-menu",
				contextMenuTrigger(),
				ui.Menu("catalog-context-content", menuItems).OnActionEvent(func(event ui.MenuActionEvent) { action(event.Key) }),
			),
		)},
		demoSection{Title: "Menu", Content: demoPanel(ui.Column(
			ui.Box(ui.Menu("catalog-inline-menu", menuItems).OnActionEvent(func(event ui.MenuActionEvent) { action(event.Key) })).Style(ui.Width(280)),
			ui.Text(model.LastAction).Size(12),
		).Gap(10))},
	)
}

func tabPanel(label string) ui.Widget {
	return ui.Surface(ui.Box(ui.Text(label + " panel")).Style(ui.FillWidth().Padding(16))).Variant(ui.SurfaceSecondary).Style(ui.Radius(8))
}

func treeIcon(data []byte) ui.Widget {
	return ui.Icon(data).Size(16)
}

var catalogTreeItems = []ui.TreeItem{
	{Key: "components", Label: "components", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []ui.TreeItem{
		{Key: "button", Label: "button.go", Leading: treeIcon(lucide.FileCode)},
		{Key: "surface", Label: "surface.go", Leading: treeIcon(lucide.FileCode)},
		{Key: "sidebar", Label: "sidebar.go", Leading: treeIcon(lucide.FileCode)},
	}},
	{Key: "readme", Label: "README.md", Leading: treeIcon(lucide.FileText)},
}

var catalogDragTreeItems = []ui.TreeItem{
	{Key: "workspace", Label: "workspace", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []ui.TreeItem{
		{Key: "workspace-src", Label: "src", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []ui.TreeItem{
			{Key: "workspace-app", Label: "app.go", Leading: treeIcon(lucide.FileCode)},
			{Key: "workspace-theme", Label: "theme.go", Leading: treeIcon(lucide.FileCode)},
		}},
		{Key: "workspace-tests", Label: "tests", Leading: treeIcon(lucide.Folder), AcceptsChildren: true},
		{Key: "workspace-readme", Label: "README.md", Leading: treeIcon(lucide.FileText)},
	}},
	{Key: "assets", Label: "assets", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Children: []ui.TreeItem{
		{Key: "assets-logo", Label: "logo.svg", Leading: treeIcon(lucide.FileImage)},
		{Key: "assets-preview", Label: "preview.png", Leading: treeIcon(lucide.FileImage)},
	}},
	{Key: "license", Label: "LICENSE", Leading: treeIcon(lucide.FileText)},
}
