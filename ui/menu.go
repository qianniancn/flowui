package ui

import "github.com/qianniancn/FlowUI/internal/components/menu"

type MenuWidget = menu.Widget
type ContextMenuWidget = menu.ContextMenuWidget
type MenuItem = menu.Item
type MenuSection = menu.Section
type MenuItemKind = menu.ItemKind
type MenuItemVariant = menu.ItemVariant

const (
	MenuItemAction     = menu.ItemAction
	MenuItemCheckbox   = menu.ItemCheckbox
	MenuItemRadio      = menu.ItemRadio
	MenuItemSubmenu    = menu.ItemSubmenu
	MenuItemSeparator  = menu.ItemSeparator
	MenuItemGroupLabel = menu.ItemGroupLabel

	MenuItemDefault = menu.ItemDefault
	MenuItemDanger  = menu.ItemDanger
)

func Menu(key string, items []MenuItem) MenuWidget {
	return menu.Menu(key, items)
}

func MenuSections(key string, sections []MenuSection) MenuWidget {
	return menu.MenuSections(key, sections)
}

func MenuSeparator() MenuItem {
	return menu.MenuSeparator()
}

func MenuGroupLabel(label string) MenuItem {
	return menu.MenuGroupLabel(label)
}

func ContextMenu(key string, trigger Widget, content MenuWidget) ContextMenuWidget {
	return menu.ContextMenu(key, trigger, content)
}
