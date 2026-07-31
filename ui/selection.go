package ui

import (
	"github.com/qianniancn/flowui/internal/components/checkbox"
	"github.com/qianniancn/flowui/internal/components/radiogroup"
	"github.com/qianniancn/flowui/internal/components/switches"
)

type CheckboxWidget = checkbox.CheckboxWidget

// CheckboxVariant selects a checkbox's visual treatment.
type CheckboxVariant = checkbox.CheckboxVariant

// CheckboxIndicatorState describes the checkbox indicator state.
type CheckboxIndicatorState = checkbox.IndicatorState

type SwitchWidget = switches.SwitchWidget

// SwitchSize selects the switch's size.
type SwitchSize = switches.SwitchSize

type SwitchGroupWidget = switches.SwitchGroupWidget

// RadioItem describes one option in a radio group.
type RadioItem = radiogroup.RadioItem

type RadioGroupWidget = radiogroup.RadioGroupWidget

// RadioGroupVariant selects a radio group's visual treatment.
type RadioGroupVariant = radiogroup.RadioGroupVariant

const (
	CheckboxPrimary   = checkbox.CheckboxPrimary
	CheckboxSecondary = checkbox.CheckboxSecondary

	SwitchMedium = switches.SwitchMedium
	SwitchSmall  = switches.SwitchSmall
	SwitchLarge  = switches.SwitchLarge

	RadioPrimary   = radiogroup.RadioPrimary
	RadioSecondary = radiogroup.RadioSecondary
)

// Checkbox creates a keyed check box initialized with checked.
func Checkbox(key string, checked bool, label string) CheckboxWidget {
	return checkbox.Checkbox(key, checked, label)
}

// Switch creates a keyed switch initialized with checked.
func Switch(key string, checked bool, label string) SwitchWidget {
	return switches.Switch(key, checked, label)
}

// SwitchGroup groups switches into one themed control group.
func SwitchGroup(children ...Widget) SwitchGroupWidget {
	return switches.SwitchGroup(children...)
}

// RadioGroup creates a single-selection group initialized with selectedKey.
func RadioGroup(key, selectedKey string, items []RadioItem) RadioGroupWidget {
	return radiogroup.RadioGroup(key, selectedKey, items)
}
