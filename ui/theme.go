package ui

import "github.com/qianniancn/FlowUI/internal/theme"

type Theme = theme.Theme
type Palette = theme.Palette
type Typography = theme.Typography
type Shape = theme.Shape
type Spacing = theme.Spacing
type MotionTheme = theme.MotionTheme
type ComponentsTheme = theme.ComponentsTheme
type DescriptionTheme = theme.DescriptionTheme
type LabelTheme = theme.LabelTheme
type ButtonTheme = theme.ButtonTheme
type ToggleButtonTheme = theme.ToggleButtonTheme
type CloseButtonTheme = theme.CloseButtonTheme
type ChipTheme = theme.ChipTheme
type AvatarTheme = theme.AvatarTheme
type BadgeTheme = theme.BadgeTheme
type CardTheme = theme.CardTheme
type AlertTheme = theme.AlertTheme
type AlertDialogTheme = theme.AlertDialogTheme
type InputTheme = theme.InputTheme
type TextAreaTheme = theme.TextAreaTheme
type InputGroupTheme = theme.InputGroupTheme
type CheckboxTheme = theme.CheckboxTheme
type SwitchTheme = theme.SwitchTheme
type SwitchGroupTheme = theme.SwitchGroupTheme
type RadioGroupTheme = theme.RadioGroupTheme
type ProgressBarTheme = theme.ProgressBarTheme
type MeterTheme = theme.MeterTheme
type SpinnerTheme = theme.SpinnerTheme
type SliderTheme = theme.SliderTheme
type ListBoxTheme = theme.ListBoxTheme
type TreeTheme = theme.TreeTheme
type TableTheme = theme.TableTheme
type MenuTheme = theme.MenuTheme
type DropdownTheme = theme.DropdownTheme
type MenubarTheme = theme.MenubarTheme
type LineChartTheme = theme.LineChartTheme
type BarChartTheme = theme.BarChartTheme
type PieChartTheme = theme.PieChartTheme
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
