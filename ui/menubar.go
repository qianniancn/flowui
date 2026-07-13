package ui

import "github.com/qianniancn/FlowUI/internal/components/menubar"

type MenubarWidget = menubar.Widget
type MenubarItem = menubar.Item
type MenubarOrientation = menubar.Orientation

const (
	MenubarHorizontal = menubar.Horizontal
	MenubarVertical   = menubar.Vertical
)

func Menubar(key string, items []MenubarItem) MenubarWidget {
	return menubar.New(key, items)
}

func MenubarMenu(key, label string, items []MenuItem) MenubarItem {
	return menubar.NewMenu(key, label, items)
}

func MenubarMenuSections(key, label string, sections []MenuSection) MenubarItem {
	return menubar.NewMenuSections(key, label, sections)
}

func MenubarMenuContent(key, label string, content MenuWidget) MenubarItem {
	return menubar.NewMenuContent(key, label, content)
}
