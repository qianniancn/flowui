package ui

import "github.com/qianniancn/flowui/internal/components/menu"

type MenuWidget = menu.Widget

type ContextMenuWidget = menu.ContextMenuWidget

// MenuItem describes one menu entry.
type MenuItem = menu.Item

// MenuSection groups related menu items.
type MenuSection = menu.Section

// MenuItemKind identifies the behavior of a menu entry.
type MenuItemKind = menu.ItemKind

// MenuItemVariant selects a menu entry's visual treatment.
type MenuItemVariant = menu.ItemVariant

// MenuSelectionMode controls how many menu entries may be selected.
type MenuSelectionMode = menu.SelectionMode

// MenuIndicatorType selects the marker shown beside a selected entry.
type MenuIndicatorType = menu.IndicatorType

const (
	MenuItemAction     = menu.ItemAction
	MenuItemCheckbox   = menu.ItemCheckbox
	MenuItemRadio      = menu.ItemRadio
	MenuItemSubmenu    = menu.ItemSubmenu
	MenuItemSeparator  = menu.ItemSeparator
	MenuItemGroupLabel = menu.ItemGroupLabel

	MenuItemDefault = menu.ItemDefault
	MenuItemDanger  = menu.ItemDanger

	MenuSelectionNone     = menu.SelectionNone
	MenuSelectionSingle   = menu.SelectionSingle
	MenuSelectionMultiple = menu.SelectionMultiple

	MenuIndicatorNone      = menu.IndicatorNone
	MenuIndicatorCheckmark = menu.IndicatorCheckmark
	MenuIndicatorDot       = menu.IndicatorDot
)

// Menu creates a menu from its items.
func Menu(key string, items []MenuItem) MenuWidget {
	return menu.Menu(key, items)
}

// MenuSections creates a menu grouped into sections.
func MenuSections(key string, sections []MenuSection) MenuWidget {
	return menu.MenuSections(key, sections)
}

// MenuSeparator returns a non-selectable separator item.
func MenuSeparator() MenuItem {
	return menu.MenuSeparator()
}

// MenuGroupLabel returns a non-selectable group heading.
func MenuGroupLabel(label string) MenuItem {
	return menu.MenuGroupLabel(label)
}

// ContextMenu opens content from a context-menu trigger, normally on right click.
func ContextMenu(key string, trigger Widget, content MenuWidget) ContextMenuWidget {
	return menu.ContextMenu(key, trigger, content)
}
