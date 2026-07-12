package ui

import (
	"time"

	"github.com/qianniancn/FlowUI/internal/components/combobox"
	"github.com/qianniancn/FlowUI/internal/components/datepicker"
	"github.com/qianniancn/FlowUI/internal/components/input"
	"github.com/qianniancn/FlowUI/internal/components/listbox"
	"github.com/qianniancn/FlowUI/internal/components/select"
)

type InputWidget = input.InputWidget
type TextAreaWidget = input.TextAreaWidget
type InputGroupWidget = input.InputGroupWidget
type InputVariant = input.InputVariant
type TextAreaVariant = input.TextAreaVariant
type InputType = input.InputType
type ComboBoxItem = combobox.ComboBoxItem
type ComboBoxWidget = combobox.ComboBoxWidget
type ListBoxItem = listbox.ListBoxItem
type ListBoxSection = listbox.ListBoxSection
type ListBoxItemVariant = listbox.ListBoxItemVariant
type ListBoxSelectionMode = listbox.ListBoxSelectionMode
type ListBoxWidget = listbox.ListBoxWidget
type SelectItem = selects.SelectItem
type SelectSection = selects.SelectSection
type SelectVariant = selects.SelectVariant
type SelectSelectionMode = selects.SelectSelectionMode
type SelectWidget = selects.SelectWidget
type DatePickerWidget = datepicker.DatePickerWidget
type DatePickerLocale = datepicker.DatePickerLocale

const (
	InputPrimary   = input.InputPrimary
	InputSecondary = input.InputSecondary
	InputText      = input.InputText
	InputEmail     = input.InputEmail
	InputNumber    = input.InputNumber
	InputPassword  = input.InputPassword

	TextAreaPrimary   = input.TextAreaPrimary
	TextAreaSecondary = input.TextAreaSecondary

	ListBoxItemDefault = listbox.ListBoxItemDefault
	ListBoxItemDanger  = listbox.ListBoxItemDanger

	ListBoxSelectionSingle   = listbox.ListBoxSelectionSingle
	ListBoxSelectionMultiple = listbox.ListBoxSelectionMultiple
	ListBoxSelectionNone     = listbox.ListBoxSelectionNone

	SelectPrimary   = selects.SelectPrimary
	SelectSecondary = selects.SelectSecondary

	SelectSelectionSingle   = selects.SelectSelectionSingle
	SelectSelectionMultiple = selects.SelectSelectionMultiple
)

func Input(key, value string) InputWidget {
	return input.Input(key, value)
}

func TextArea(key, value string) TextAreaWidget {
	return input.TextArea(key, value)
}

func InputGroup(field InputWidget) InputGroupWidget {
	return input.InputGroup(field)
}

func InputGroupTextArea(field TextAreaWidget) InputGroupWidget {
	return input.InputGroupTextArea(field)
}

func ComboBox(key, selectedKey string, items []ComboBoxItem) ComboBoxWidget {
	return combobox.ComboBox(key, selectedKey, items)
}

func ListBox(key, selectedKey string, items []ListBoxItem) ListBoxWidget {
	return listbox.ListBox(key, selectedKey, items)
}

func ListBoxMultiple(key string, selectedKeys []string, items []ListBoxItem) ListBoxWidget {
	return listbox.ListBoxMultiple(key, selectedKeys, items)
}

func ListBoxSections(key, selectedKey string, sections []ListBoxSection) ListBoxWidget {
	return listbox.ListBoxSections(key, selectedKey, sections)
}

func ListBoxMultipleSections(key string, selectedKeys []string, sections []ListBoxSection) ListBoxWidget {
	return listbox.ListBoxMultipleSections(key, selectedKeys, sections)
}

func Select(key, selectedKey string, items []SelectItem) SelectWidget {
	return selects.Select(key, selectedKey, items)
}

func SelectMultiple(key string, selectedKeys []string, items []SelectItem) SelectWidget {
	return selects.SelectMultiple(key, selectedKeys, items)
}

func SelectSections(key, selectedKey string, sections []SelectSection) SelectWidget {
	return selects.SelectSections(key, selectedKey, sections)
}

func SelectMultipleSections(key string, selectedKeys []string, sections []SelectSection) SelectWidget {
	return selects.SelectMultipleSections(key, selectedKeys, sections)
}

func DatePickerEnglish() DatePickerLocale {
	return datepicker.DatePickerEnglish()
}

func DatePickerChinese() DatePickerLocale {
	return datepicker.DatePickerChinese()
}

func DatePicker(key string, value time.Time) DatePickerWidget {
	return datepicker.DatePicker(key, value)
}
