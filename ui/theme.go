package ui

import "github.com/qianniancn/FlowUI/internal/theme"

type Theme = theme.Theme
type Palette = theme.Palette
type Typography = theme.Typography
type Shape = theme.Shape
type Spacing = theme.Spacing
type ComponentsTheme = theme.ComponentsTheme
type DescriptionTheme = theme.DescriptionTheme
type LabelTheme = theme.LabelTheme
type ButtonTheme = theme.ButtonTheme
type ToggleButtonTheme = theme.ToggleButtonTheme
type CloseButtonTheme = theme.CloseButtonTheme
type ChipTheme = theme.ChipTheme
type CardTheme = theme.CardTheme
type AlertTheme = theme.AlertTheme
type AlertDialogTheme = theme.AlertDialogTheme
type InputTheme = theme.InputTheme
type InputGroupTheme = theme.InputGroupTheme
type CheckboxTheme = theme.CheckboxTheme
type SwitchTheme = theme.SwitchTheme
type SwitchGroupTheme = theme.SwitchGroupTheme
type RadioGroupTheme = theme.RadioGroupTheme
type ProgressBarTheme = theme.ProgressBarTheme
type SpinnerTheme = theme.SpinnerTheme
type SliderTheme = theme.SliderTheme
type ListBoxTheme = theme.ListBoxTheme
type TreeTheme = theme.TreeTheme
type TabsTheme = theme.TabsTheme
type SelectTheme = theme.SelectTheme
type PopoverTheme = theme.PopoverTheme
type TooltipTheme = theme.TooltipTheme
type ToastTheme = theme.ToastTheme
type ModalTheme = theme.ModalTheme
type ComboBoxTheme = theme.ComboBoxTheme
type DatePickerTheme = theme.DatePickerTheme

func DefaultTheme() Theme {
	return theme.DefaultTheme()
}

func DarkTheme() Theme {
	return theme.DarkTheme()
}

func syncMaterialTheme(activeTheme *Theme) {
	theme.SyncMaterialTheme(activeTheme)
}
