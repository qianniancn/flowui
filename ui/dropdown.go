package ui

import (
	"github.com/qianniancn/FlowUI/internal/components/dropdown"
)

type DropdownWidget = dropdown.Widget
type DropdownItem = dropdown.Item
type DropdownSection = dropdown.Section
type DropdownSelectionMode = dropdown.SelectionMode
type DropdownItemVariant = dropdown.ItemVariant
type DropdownIndicatorType = dropdown.IndicatorType
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

func Dropdown(key string, trigger Widget, items []DropdownItem) DropdownWidget {
	return dropdown.New(key, trigger, items)
}

func DropdownSections(key string, trigger Widget, sections []DropdownSection) DropdownWidget {
	return dropdown.NewSections(key, trigger, sections)
}

func DropdownSeparator() DropdownItem {
	return dropdown.Separator()
}
