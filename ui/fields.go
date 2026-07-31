package ui

import (
	"time"

	"github.com/qianniancn/flowui/internal/components/combobox"
	"github.com/qianniancn/flowui/internal/components/datepicker"
	"github.com/qianniancn/flowui/internal/components/input"
	"github.com/qianniancn/flowui/internal/components/listbox"
	"github.com/qianniancn/flowui/internal/components/select"
)

type InputWidget = input.InputWidget

type TextAreaWidget = input.TextAreaWidget

type InputGroupWidget = input.InputGroupWidget

// InputVariant selects an input's visual treatment.
type InputVariant = input.InputVariant

// TextAreaVariant selects a text area's visual treatment.
type TextAreaVariant = input.TextAreaVariant

// InputType controls the kind of value accepted by an input.
type InputType = input.InputType

// ComboBoxItem describes one option in a combo box.
type ComboBoxItem = combobox.ComboBoxItem

type ComboBoxWidget = combobox.ComboBoxWidget

// ListBoxItem describes one option in a list box.
type ListBoxItem = listbox.ListBoxItem

// ListBoxSection groups related list-box items.
type ListBoxSection = listbox.ListBoxSection

// ListBoxItemVariant selects a list-box item's visual treatment.
type ListBoxItemVariant = listbox.ListBoxItemVariant

// ListBoxSelectionMode controls how many list-box items may be selected.
type ListBoxSelectionMode = listbox.ListBoxSelectionMode

type ListBoxWidget = listbox.ListBoxWidget

// SelectItem describes one option in a select field.
type SelectItem = selects.SelectItem

// SelectSection groups related select items.
type SelectSection = selects.SelectSection

// SelectVariant selects a select field's visual treatment.
type SelectVariant = selects.SelectVariant

// SelectSelectionMode controls how many select items may be selected.
type SelectSelectionMode = selects.SelectSelectionMode

type SelectWidget = selects.SelectWidget

type DatePickerWidget = datepicker.DatePickerWidget

// DatePickerLocale contains the localized labels used by a date picker.
type DatePickerLocale = datepicker.DatePickerLocale

// DatePart identifies a part of a date that can be edited.
type DatePart = datepicker.DatePart

type DateFieldWidget = datepicker.DateFieldWidget

// DateRange stores the beginning and end of a selected date range.
type DateRange = datepicker.DateRange

type DateRangePickerWidget = datepicker.DateRangePickerWidget

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

const (
	DatePartYear  = datepicker.DatePartYear
	DatePartMonth = datepicker.DatePartMonth
	DatePartDay   = datepicker.DatePartDay
)

// Input creates a single-line controlled text field initialized with value.
func Input(key, value string) InputWidget {
	return input.Input(key, value)
}

// TextArea creates a multiline controlled text field initialized with value.
func TextArea(key, value string) TextAreaWidget {
	return input.TextArea(key, value)
}

// InputGroup wraps an input field with the theme's grouped field layout.
func InputGroup(field InputWidget) InputGroupWidget {
	return input.InputGroup(field)
}

// InputGroupTextArea wraps a text area with the theme's grouped field layout.
func InputGroupTextArea(field TextAreaWidget) InputGroupWidget {
	return input.InputGroupTextArea(field)
}

// ComboBox creates an editable field with a popup list of items.
func ComboBox(key, selectedKey string, items []ComboBoxItem) ComboBoxWidget {
	return combobox.ComboBox(key, selectedKey, items)
}

// ListBox creates a single-selection list box.
func ListBox(key, selectedKey string, items []ListBoxItem) ListBoxWidget {
	return listbox.ListBox(key, selectedKey, items)
}

// ListBoxMultiple creates a list box that accepts multiple selected keys.
func ListBoxMultiple(key string, selectedKeys []string, items []ListBoxItem) ListBoxWidget {
	return listbox.ListBoxMultiple(key, selectedKeys, items)
}

// ListBoxSections creates a single-selection list box grouped into sections.
func ListBoxSections(key, selectedKey string, sections []ListBoxSection) ListBoxWidget {
	return listbox.ListBoxSections(key, selectedKey, sections)
}

// ListBoxMultipleSections creates a multi-selection list box grouped into sections.
func ListBoxMultipleSections(key string, selectedKeys []string, sections []ListBoxSection) ListBoxWidget {
	return listbox.ListBoxMultipleSections(key, selectedKeys, sections)
}

// Select creates a single-selection popup field.
func Select(key, selectedKey string, items []SelectItem) SelectWidget {
	return selects.Select(key, selectedKey, items)
}

// SelectMultiple creates a popup field that accepts multiple selected keys.
func SelectMultiple(key string, selectedKeys []string, items []SelectItem) SelectWidget {
	return selects.SelectMultiple(key, selectedKeys, items)
}

// SelectSections creates a sectioned single-selection popup field.
func SelectSections(key, selectedKey string, sections []SelectSection) SelectWidget {
	return selects.SelectSections(key, selectedKey, sections)
}

// SelectMultipleSections creates a sectioned multi-selection popup field.
func SelectMultipleSections(key string, selectedKeys []string, sections []SelectSection) SelectWidget {
	return selects.SelectMultipleSections(key, selectedKeys, sections)
}

// DatePickerEnglish returns the English date-picker locale.
func DatePickerEnglish() DatePickerLocale {
	return datepicker.DatePickerEnglish()
}

// DatePickerChinese returns the Chinese date-picker locale.
func DatePickerChinese() DatePickerLocale {
	return datepicker.DatePickerChinese()
}

// DatePicker creates a calendar picker initialized with value.
func DatePicker(key string, value time.Time) DatePickerWidget {
	return datepicker.DatePicker(key, value)
}

// DateField creates a text field with an attached calendar picker.
func DateField(key string, value time.Time) DateFieldWidget {
	return datepicker.DateField(key, value)
}

// DateRangePicker creates a calendar picker for selecting a date range.
func DateRangePicker(key string, value DateRange) DateRangePickerWidget {
	return datepicker.DateRangePicker(key, value)
}
