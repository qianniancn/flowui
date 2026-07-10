package ui

import (
	"github.com/qianniancn/FlowUI/internal/components/checkbox"
	"github.com/qianniancn/FlowUI/internal/components/radiogroup"
	"github.com/qianniancn/FlowUI/internal/components/switches"
)

type CheckboxWidget = checkbox.CheckboxWidget
type SwitchWidget = switches.SwitchWidget
type SwitchSize = switches.SwitchSize
type SwitchGroupWidget = switches.SwitchGroupWidget
type RadioItem = radiogroup.RadioItem
type RadioGroupWidget = radiogroup.RadioGroupWidget
type RadioGroupVariant = radiogroup.RadioGroupVariant

const (
	SwitchMedium = switches.SwitchMedium
	SwitchSmall  = switches.SwitchSmall
	SwitchLarge  = switches.SwitchLarge

	RadioPrimary   = radiogroup.RadioPrimary
	RadioSecondary = radiogroup.RadioSecondary
)

func Checkbox(key string, checked bool, label string) CheckboxWidget {
	return checkbox.Checkbox(key, checked, label)
}

func Switch(key string, checked bool, label string) SwitchWidget {
	return switches.Switch(key, checked, label)
}

func SwitchGroup(children ...Widget) SwitchGroupWidget {
	return switches.SwitchGroup(children...)
}

func RadioGroup(key, selectedKey string, items []RadioItem) RadioGroupWidget {
	return radiogroup.RadioGroup(key, selectedKey, items)
}
