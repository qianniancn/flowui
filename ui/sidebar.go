package ui

import "github.com/qianniancn/FlowUI/internal/components/sidebar"

type SidebarWidget = sidebar.Widget
type SidebarItem = sidebar.Item
type SidebarSection = sidebar.Section

func Sidebar(key, selectedKey string, items []SidebarItem) SidebarWidget {
	return sidebar.New(key, selectedKey, items)
}

func SidebarSections(key, selectedKey string, sections []SidebarSection) SidebarWidget {
	return sidebar.NewSections(key, selectedKey, sections)
}
