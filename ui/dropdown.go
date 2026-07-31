package ui

import (
	"github.com/qianniancn/flowui/internal/components/dropdown"
)

type DropdownWidget = dropdown.Widget

// DropdownItem describes one entry in a dropdown menu.
type DropdownItem = dropdown.Item

// DropdownSection groups related dropdown items.
type DropdownSection = dropdown.Section

// DropdownSelectionMode controls how many dropdown items may be selected.
type DropdownSelectionMode = dropdown.SelectionMode

// DropdownItemVariant selects an item's visual treatment.
type DropdownItemVariant = dropdown.ItemVariant

// DropdownIndicatorType selects the marker shown beside a selected item.
type DropdownIndicatorType = dropdown.IndicatorType

// DropdownTriggerMode controls which pointer gesture opens a dropdown.
type DropdownTriggerMode = dropdown.TriggerMode

const (
	DropdownSelectionNone     = dropdown.SelectionNone
	DropdownSelectionSingle   = dropdown.SelectionSingle
	DropdownSelectionMultiple = dropdown.SelectionMultiple

	DropdownItemDefault = dropdown.ItemDefault
	DropdownItemDanger  = dropdown.ItemDanger

	DropdownIndicatorNone      = dropdown.IndicatorNone
	DropdownIndicatorCheckmark = dropdown.IndicatorCheckmark
	DropdownIndicatorDot       = dropdown.IndicatorDot

	DropdownTriggerPress     = dropdown.TriggerPress
	DropdownTriggerLongPress = dropdown.TriggerLongPress
)

// Dropdown creates a popup menu opened by trigger.
func Dropdown(key string, trigger Widget, items []DropdownItem) DropdownWidget {
	return dropdown.New(key, trigger, items)
}

// DropdownSections creates a popup menu grouped into sections.
func DropdownSections(key string, trigger Widget, sections []DropdownSection) DropdownWidget {
	return dropdown.NewSections(key, trigger, sections)
}

// DropdownSeparator returns a non-selectable separator item.
func DropdownSeparator() DropdownItem {
	return dropdown.Separator()
}
