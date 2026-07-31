package ui

import "github.com/qianniancn/flowui/internal/components/menubar"

type MenubarWidget = menubar.Widget

// MenubarItem describes one entry in a menu bar.
type MenubarItem = menubar.Item

// MenubarOrientation controls the direction of a menu bar.
type MenubarOrientation = menubar.Orientation

const (
	MenubarHorizontal = menubar.Horizontal
	MenubarVertical   = menubar.Vertical
)

// Menubar creates a horizontal or vertical menu bar.
func Menubar(key string, items []MenubarItem) MenubarWidget {
	return menubar.New(key, items)
}

// MenubarMenu creates a menu-bar item backed by menu items.
func MenubarMenu(key, label string, items []MenuItem) MenubarItem {
	return menubar.NewMenu(key, label, items)
}

// MenubarMenuSections creates a menu-bar item backed by menu sections.
func MenubarMenuSections(key, label string, sections []MenuSection) MenubarItem {
	return menubar.NewMenuSections(key, label, sections)
}

// MenubarMenuContent creates a menu-bar item backed by custom content.
func MenubarMenuContent(key, label string, content MenuWidget) MenubarItem {
	return menubar.NewMenuContent(key, label, content)
}
