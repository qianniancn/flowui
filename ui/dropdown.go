package ui

import (
	"github.com/qianniancn/flowui/internal/components/dropdown"
)

type DropdownWidget = dropdown.Widget

// DropdownButtonWidget combines a primary button with a dropdown trigger.
type DropdownButtonWidget = dropdown.ButtonWidget

// DropdownItem describes one entry in a dropdown menu.
type DropdownItem = dropdown.Item

// DropdownItemKind identifies the behavior of a dropdown entry.
type DropdownItemKind = dropdown.ItemKind

// DropdownSection groups related dropdown items.
type DropdownSection = dropdown.Section

// DropdownSelectionMode controls how many dropdown items may be selected.
type DropdownSelectionMode = dropdown.SelectionMode

// DropdownItemVariant selects an item's visual treatment.
type DropdownItemVariant = dropdown.ItemVariant

// DropdownIndicatorType selects the marker shown beside a selected item.
type DropdownIndicatorType = dropdown.IndicatorType

// DropdownActionEvent describes an activated item and its submenu path.
type DropdownActionEvent = dropdown.ActionEvent

// DropdownTriggerMode controls which pointer gesture opens a dropdown.
type DropdownTriggerMode = dropdown.TriggerMode

// DropdownOpenChangeSource identifies what requested a dropdown state change.
type DropdownOpenChangeSource = dropdown.OpenChangeSource

// DropdownOpenChangeEvent describes a dropdown state change and its source.
type DropdownOpenChangeEvent = dropdown.OpenChangeEvent

const (
	DropdownItemAction     = dropdown.ItemAction
	DropdownItemCheckbox   = dropdown.ItemCheckbox
	DropdownItemRadio      = dropdown.ItemRadio
	DropdownItemSubmenu    = dropdown.ItemSubmenu
	DropdownItemSeparator  = dropdown.ItemSeparator
	DropdownItemGroupLabel = dropdown.ItemGroupLabel

	DropdownSelectionNone     = dropdown.SelectionNone
	DropdownSelectionSingle   = dropdown.SelectionSingle
	DropdownSelectionMultiple = dropdown.SelectionMultiple

	DropdownItemDefault = dropdown.ItemDefault
	DropdownItemDanger  = dropdown.ItemDanger

	DropdownIndicatorNone      = dropdown.IndicatorNone
	DropdownIndicatorCheckmark = dropdown.IndicatorCheckmark
	DropdownIndicatorDot       = dropdown.IndicatorDot

	DropdownTriggerPress       = dropdown.TriggerPress
	DropdownTriggerLongPress   = dropdown.TriggerLongPress
	DropdownTriggerHover       = dropdown.TriggerHover
	DropdownTriggerContextMenu = dropdown.TriggerContextMenu

	DropdownOpenChangeProgrammatic = dropdown.OpenChangeProgrammatic
	DropdownOpenChangeTrigger      = dropdown.OpenChangeTrigger
	DropdownOpenChangeMenu         = dropdown.OpenChangeMenu
	DropdownOpenChangeOutside      = dropdown.OpenChangeOutside
	DropdownOpenChangeKeyboard     = dropdown.OpenChangeKeyboard
	DropdownOpenChangePeer         = dropdown.OpenChangePeer
	DropdownOpenChangeContextMenu  = dropdown.OpenChangeContextMenu
)

// Dropdown creates a popup menu opened by trigger.
func Dropdown(key string, trigger Widget, items []DropdownItem) DropdownWidget {
	return dropdown.New(key, trigger, items)
}

// DropdownButton creates a split button with a primary action and an ellipsis
// dropdown trigger.
func DropdownButton(key string, action ButtonWidget, items []DropdownItem) DropdownButtonWidget {
	return dropdown.Button(key, action, items)
}

// DropdownSections creates a popup menu grouped into sections.
func DropdownSections(key string, trigger Widget, sections []DropdownSection) DropdownWidget {
	return dropdown.NewSections(key, trigger, sections)
}

// DropdownSeparator returns a non-selectable separator item.
func DropdownSeparator() DropdownItem {
	return dropdown.Separator()
}

// DropdownGroupLabel returns a non-selectable heading for a dropdown menu.
func DropdownGroupLabel(label string) DropdownItem {
	return dropdown.GroupLabel(label)
}
