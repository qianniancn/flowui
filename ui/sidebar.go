package ui

import "github.com/qianniancn/flowui/internal/components/sidebar"

type SidebarWidget = sidebar.Widget

// SidebarItem describes one navigation entry. Children form an inline
// navigation group controlled by SidebarWidget.OpenKeys.
type SidebarItem = sidebar.Item

// SidebarSection groups related navigation entries.
type SidebarSection = sidebar.Section

// SidebarExpandAction controls how a nested Sidebar item opens while expanded.
// Collapsed groups always open their flyout on hover.
type SidebarExpandAction = sidebar.ExpandAction

const (
	SidebarExpandOnClick = sidebar.ExpandActionClick
	SidebarExpandOnHover = sidebar.ExpandActionHover
)

// Sidebar creates a navigation sidebar with one selected item.
func Sidebar(key, selectedKey string, items []SidebarItem) SidebarWidget {
	return sidebar.New(key, selectedKey, items)
}

// SidebarSections creates a sectioned navigation sidebar.
func SidebarSections(key, selectedKey string, sections []SidebarSection) SidebarWidget {
	return sidebar.NewSections(key, selectedKey, sections)
}
