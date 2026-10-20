package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/pagination"
	"github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/tree"
)

func tabsPaginationPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Tabs & pagination",
		demoSection{Title: "Tabs", Content: demoPanel(
			tabs.New("catalog-tabs", model.TabValue, []tabs.TabItem{
				{Key: "preview", Label: "Preview", Panel: tabPanel("Preview")},
				{Key: "code", Label: "Code", Panel: tabPanel("Code")},
				{Key: "accessibility", Label: "Accessibility", Panel: tabPanel("Accessibility")},
			}).Separators(true).OnChange(func(value string) { send(func(model *Model) { model.TabValue = value }) }),
		)},
		demoSection{Title: "Pagination", Content: demoPanel(
			pagination.New("catalog-pagination", model.PaginationPage, 12).
				Summary(text.New("61 to 72 of 144 results").Size(12)).
				OnChange(func(value int) { send(func(model *Model) { model.PaginationPage = value }) }),
		)},
	)
}

func sidebarTreePage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	guideStyle := tree.GuideSolid
	if model.TreeDashed {
		guideStyle = tree.GuideDashed
	}
	return demoPage("Sidebar & tree",
		demoSection{Title: "Sidebar", Content: demoPanel(
			layout.Box(sidebar.New("catalog-sidebar-demo", model.SidebarValue, []sidebar.Item{
				{Key: "overview", Label: "Overview", Leading: icon.New(lucide.LayoutDashboard).Size(16)},
				{Key: "projects", Label: "Projects", Leading: icon.New(lucide.FolderKanban).Size(16), Trailing: chip.New("8").Size(chip.SizeSmall)},
				{Key: "settings", Label: "Settings", Leading: icon.New(lucide.Settings).Size(16)},
			}).OnChange(func(value string) { send(func(model *Model) { model.SidebarValue = value }) })).Style(style.New().Width(320).Height(260)),
		)},
		demoSection{Title: "Tree", Content: demoPanel(layout.Column(
			switchcontrol.SwitchGroup(
				switchcontrol.New("Branch connectors").Key("catalog-tree-connectors").Checked(model.TreeConnectors).
					Size(switchcontrol.SwitchSmall).
					OnChange(func(value bool) { send(func(model *Model) { model.TreeConnectors = value }) }),
				switchcontrol.New("Dashed guides").Key("catalog-tree-dashed").Checked(model.TreeDashed).
					Size(switchcontrol.SwitchSmall).
					OnChange(func(value bool) { send(func(model *Model) { model.TreeDashed = value }) }),
			).Horizontal(),
			layout.Box(tree.New("catalog-tree", model.TreeValue, catalogTreeItems).
				ExpandedKeys(model.TreeExpanded).
				Guides(true).
				GuideConnectors(model.TreeConnectors).
				GuideStyle(guideStyle).
				OnChange(func(value string) { send(func(model *Model) { model.TreeValue = value }) }).
				OnExpandedChange(func(values []string) {
					send(func(model *Model) { model.TreeExpanded = append([]string(nil), values...) })
				})).Style(style.New().Width(440)),
		).Gap(8))},
		demoSection{Title: "Tree drag and drop", Content: demoPanel(layout.Column(
			layout.Box(tree.New("catalog-tree-drag", model.TreeDragValue, model.TreeDragItems).
				ExpandedKeys(model.TreeDragExpanded).
				Variant(tree.VariantSurface).
				Guides(true).
				GuideConnectors(true).
				GuideStyle(tree.GuideDashed).
				MaxHeight(300).
				OnChange(func(value string) {
					send(func(model *Model) { model.TreeDragValue = value })
				}).
				OnExpandedChange(func(values []string) {
					send(func(model *Model) { model.TreeDragExpanded = append([]string(nil), values...) })
				}).
				CanDrop(func(event tree.DropEvent) bool {
					return catalogTreeCanDrop(model.TreeDragItems, event)
				}).
				OnDrop(func(event tree.DropEvent) {
					send(func(model *Model) {
						if !catalogTreeMove(&model.TreeDragItems, event) {
							return
						}
						model.TreeDropMessage = fmt.Sprintf("Moved %s %s %s", event.SourceKey, treeDropPositionLabel(event.Position), event.TargetKey)
						if event.Position == tree.DropInside && !containsCatalogTreeKey(model.TreeDragExpanded, event.TargetKey) {
							model.TreeDragExpanded = append(model.TreeDragExpanded, event.TargetKey)
						}
					})
				}),
			).Style(style.New().Width(440)),
			text.New(model.TreeDropMessage).Size(12),
		).Gap(10))},
	)
}

func menusPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	action := func(value string) { send(func(model *Model) { model.LastAction = value }) }
	menuItems := []menu.Item{
		{Key: "new", Label: "New file", Shortcut: "Ctrl+N", Leading: icon.New(lucide.FilePlus).Size(16)},
		{Key: "open", Label: "Open", Shortcut: "Ctrl+O", Leading: icon.New(lucide.FolderOpen).Size(16)},
		menu.Separator(),
		{Key: "delete", Label: "Delete", Variant: menu.ItemDanger, Leading: icon.New(lucide.Trash).Size(16)},
	}
	menuSections := []menu.Section{
		{Title: "Create", Items: []menu.Item{
			{Key: "new-file", Label: "New file", Leading: icon.New(lucide.FilePlus).Size(16)},
			{Key: "new-folder", Label: "New folder", Leading: icon.New(lucide.Folder).Size(16)},
		}},
		{Title: "Transfer", Items: []menu.Item{
			{Key: "download", Label: "Download", Leading: icon.New(lucide.Download).Size(16)},
			{Key: "archive", Label: "Archive", Leading: icon.New(lucide.Archive).Size(16)},
		}},
	}
	return demoPage("Menus",
		demoSection{Title: "Menubar", Content: demoPanel(
			menubar.New("catalog-menubar", []menubar.Item{
				menubar.NewMenu("catalog-file-menu", "File", menuItems).OnActionEvent(func(event menu.ActionEvent) { action(event.Key) }),
				menubar.NewMenu("catalog-edit-menu", "Edit", []menu.Item{
					{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: icon.New(lucide.Copy).Size(16)},
					{Key: "paste", Label: "Paste", Shortcut: "Ctrl+V", Leading: icon.New(lucide.ClipboardPaste).Size(16)},
				}).OnActionEvent(func(event menu.ActionEvent) { action(event.Key) }),
			}).Alt("Catalog menu bar"),
		)},
		demoSection{Title: "Dropdown", Content: demoPanel(
			dropdown.New("catalog-dropdown",
				button.New("").Content(layout.Row(text.New("Actions"), icon.New(lucide.ChevronDown).Size(14)).Gap(7)).Variant(button.Secondary),
				[]dropdown.Item{
					{Key: "preview", Label: "Preview", Leading: icon.New(lucide.Eye).Size(16)},
					{Key: "download", Label: "Download", Leading: icon.New(lucide.Download).Size(16)},
					{Key: "archive", Label: "Archive", Leading: icon.New(lucide.Archive).Size(16)},
				},
			).OnActionEvent(func(event dropdown.ActionEvent) { action(event.Key) }),
		)},
		demoSection{Title: "DropdownButton", Content: demoPanel(
			dropdown.Button("catalog-dropdown-button",
				button.New("Create"),
				[]dropdown.Item{
					{Key: "project", Label: "New project", Leading: icon.New(lucide.FolderPlus).Size(16)},
					{Key: "file", Label: "New file", Leading: icon.New(lucide.FilePlus).Size(16)},
				},
			).Variant(button.Secondary).
				OnClick(func() { action("Create") }).
				OnActionEvent(func(event dropdown.ActionEvent) { action(event.Key) }),
		)},
		demoSection{Title: "DropdownSections & MenuSections", Content: demoPanel(demoRow(
			dropdown.NewSections("catalog-dropdown-sections",
				button.New("").Content(layout.Row(
					text.New("Sectioned actions"),
					icon.New(lucide.ChevronDown).Size(14),
				).Gap(7)).Variant(button.Secondary),
				menuSections,
			).OnActionEvent(func(event dropdown.ActionEvent) { action(event.Key) }),
			layout.Box(menu.Sections("catalog-menu-sections", menuSections).OnActionEvent(func(event menu.ActionEvent) { action(event.Key) })).Style(style.New().Width(300)),
		))},
		demoSection{Title: "ContextMenu", Content: demoPanel(
			menu.ContextMenu("catalog-context-menu",
				contextMenuTrigger(),
				menu.New("catalog-context-content", menuItems).OnActionEvent(func(event menu.ActionEvent) { action(event.Key) }),
			),
		)},
		demoSection{Title: "Menu", Content: demoPanel(layout.Column(
			layout.Box(menu.New("catalog-inline-menu", menuItems).OnActionEvent(func(event menu.ActionEvent) { action(event.Key) })).Style(style.New().Width(280)),
			text.New(model.LastAction).Size(12),
		).Gap(10))},
	)
}

func tabPanel(label string) ui.Widget {
	return (surface.New(layout.Box(text.New(label + " panel")).Style(style.New().FillWidth().Padding(16))).Variant(surface.Secondary)).Style(style.New().Radius(8))
}

func treeIcon(data []byte) ui.Widget {
	return icon.New(data).Size(16)
}

var catalogTreeItems = []tree.Item{
	{Key: "components", Label: "components", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []tree.Item{
		{Key: "button", Label: "button.go", Leading: treeIcon(lucide.FileCode)},
		{Key: "surface", Label: "surface.go", Leading: treeIcon(lucide.FileCode)},
		{Key: "sidebar", Label: "sidebar.go", Leading: treeIcon(lucide.FileCode)},
	}},
	{Key: "readme", Label: "README.md", Leading: treeIcon(lucide.FileText)},
}

var catalogDragTreeItems = []tree.Item{
	{Key: "workspace", Label: "workspace", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []tree.Item{
		{Key: "workspace-src", Label: "src", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), Children: []tree.Item{
			{Key: "workspace-app", Label: "app.go", Leading: treeIcon(lucide.FileCode)},
			{Key: "workspace-theme", Label: "theme.go", Leading: treeIcon(lucide.FileCode)},
		}},
		{Key: "workspace-tests", Label: "tests", Leading: treeIcon(lucide.Folder), AcceptsChildren: true},
		{Key: "workspace-readme", Label: "README.md", Leading: treeIcon(lucide.FileText)},
	}},
	{Key: "assets", Label: "assets", Leading: treeIcon(lucide.Folder), ExpandedLeading: treeIcon(lucide.FolderOpen), AcceptsChildren: true, Children: []tree.Item{
		{Key: "assets-logo", Label: "logo.svg", Leading: treeIcon(lucide.FileImage)},
		{Key: "assets-preview", Label: "preview.png", Leading: treeIcon(lucide.FileImage)},
	}},
	{Key: "license", Label: "LICENSE", Leading: treeIcon(lucide.FileText)},
}
