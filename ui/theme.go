package ui

import (
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/theme"
)

// Theme is the complete FlowUI design-token tree used by a window.
type Theme = theme.Theme

// Palette contains semantic colors used by FlowUI components.
type Palette = theme.Palette

// Typography contains the default text families and sizes for the theme.
type Typography = theme.Typography

// FontConfig controls bundled and system font sources for a theme.
type FontConfig = theme.FontConfig

// Shape contains shared corner-radius values.
type Shape = theme.Shape

// Spacing contains shared control sizes and spacing values.
type Spacing = theme.Spacing

// ShadowLayerTheme describes one layer of a component shadow.
type ShadowLayerTheme = theme.ShadowLayerTheme

// ShadowTheme contains the layers that make up one shadow profile.
type ShadowTheme = theme.ShadowTheme

// ShadowsTheme contains the shadow profiles used by FlowUI surfaces and controls.
type ShadowsTheme = theme.ShadowsTheme

// MotionTheme controls whether animations run and how long they take.
type MotionTheme = theme.MotionTheme

// ComponentsTheme groups the theme values for FlowUI components.
type ComponentsTheme = theme.ComponentsTheme

// DescriptionTheme contains the text settings for descriptions.
type DescriptionTheme = theme.DescriptionTheme

// LabelTheme contains the text and layout settings for labels.
type LabelTheme = theme.LabelTheme

// ButtonTheme contains the sizing and interaction settings for buttons.
type ButtonTheme = theme.ButtonTheme

// ButtonGroupTheme contains the separator settings for grouped buttons.
type ButtonGroupTheme = theme.ButtonGroupTheme

// ToggleButtonTheme contains the sizing and interaction settings for toggle buttons.
type ToggleButtonTheme = theme.ToggleButtonTheme

// CloseButtonTheme contains the sizing and interaction settings for close buttons.
type CloseButtonTheme = theme.CloseButtonTheme

// ChipTheme contains the sizing and text settings for chips.
type ChipTheme = theme.ChipTheme

// AvatarTheme contains the size, shape, and text settings for avatars.
type AvatarTheme = theme.AvatarTheme

// BadgeTheme contains the size, shape, and text settings for badges.
type BadgeTheme = theme.BadgeTheme

// SurfaceTheme contains the border settings for surfaces.
type SurfaceTheme = theme.SurfaceTheme

// CardTheme contains the padding and shape settings for cards.
type CardTheme = theme.CardTheme

// AlertTheme contains the layout and typography settings for alerts.
type AlertTheme = theme.AlertTheme

// AlertDialogTheme contains the layout and typography settings for alert dialogs.
type AlertDialogTheme = theme.AlertDialogTheme

// InputTheme contains the sizing and focus settings for single-line inputs.
type InputTheme = theme.InputTheme

// TextAreaTheme contains the sizing and focus settings for text areas.
type TextAreaTheme = theme.TextAreaTheme

// InputGroupTheme contains the sizing and focus settings for grouped inputs.
type InputGroupTheme = theme.InputGroupTheme

// CheckboxTheme contains the sizing and indicator settings for checkboxes.
type CheckboxTheme = theme.CheckboxTheme

// SwitchTheme contains the sizing and interaction settings for switches.
type SwitchTheme = theme.SwitchTheme

// SwitchGroupTheme contains the spacing settings for groups of switches.
type SwitchGroupTheme = theme.SwitchGroupTheme

// RadioGroupTheme contains the sizing and interaction settings for radio groups.
type RadioGroupTheme = theme.RadioGroupTheme

// ProgressBarTheme contains the sizing and typography settings for progress bars.
type ProgressBarTheme = theme.ProgressBarTheme

// ProgressCircleTheme contains the size and stroke settings for progress circles.
type ProgressCircleTheme = theme.ProgressCircleTheme

// SpinnerTheme contains the size and stroke settings for spinners.
type SpinnerTheme = theme.SpinnerTheme

// SliderTheme contains the track, thumb, and label settings for sliders.
type SliderTheme = theme.SliderTheme

// ListBoxTheme contains the sizing and item layout settings for list boxes.
type ListBoxTheme = theme.ListBoxTheme

// TreeTheme contains the row, indentation, and drag-preview settings for trees.
type TreeTheme = theme.TreeTheme

// SidebarTheme contains the sizing and item layout settings for sidebars.
type SidebarTheme = theme.SidebarTheme

// ScrollbarTheme contains the sizing and opacity settings for scrollbars.
type ScrollbarTheme = theme.ScrollbarTheme

// SplitPaneTheme contains the divider and handle settings for split panes.
type SplitPaneTheme = theme.SplitPaneTheme

// TitleBarTheme contains the sizing and control colors for title bars.
type TitleBarTheme = theme.TitleBarTheme

// ToolbarTheme contains the spacing and separator settings for toolbars.
type ToolbarTheme = theme.ToolbarTheme

// TableTheme contains the sizing, spacing, and selection settings for tables.
type TableTheme = theme.TableTheme

// PaginationTheme contains the sizing and spacing settings for pagination controls.
type PaginationTheme = theme.PaginationTheme

// MenuTheme contains the colors, layout, and interaction settings for menus.
type MenuTheme = theme.MenuTheme

// DropdownTheme contains the trigger and panel settings for dropdowns.
type DropdownTheme = theme.DropdownTheme

// MenubarTheme contains the trigger and spacing settings for menu bars.
type MenubarTheme = theme.MenubarTheme

// LineChartTheme contains the layout and series settings for line charts.
type LineChartTheme = theme.LineChartTheme

// BarChartTheme contains the layout and series settings for bar charts.
type BarChartTheme = theme.BarChartTheme

// PieChartTheme contains the layout and series settings for pie charts.
type PieChartTheme = theme.PieChartTheme

// CandlestickChartTheme contains the layout and series settings for candlestick charts.
type CandlestickChartTheme = theme.CandlestickChartTheme

// TabsTheme contains the sizing, spacing, content, and motion tokens for tabs.
type TabsTheme = theme.TabsTheme

// NodeGraphTheme contains canvas, node, edge, and selection colors for node graphs.
type NodeGraphTheme = theme.NodeGraphTheme

// WorkbenchTheme contains shell-level sidebar, editor, panel, divider, tab,
// and density tokens for editor-like workbenches.
type WorkbenchTheme = theme.WorkbenchTheme

// CollapsibleTheme contains the spacing and animation settings for collapsibles.
type CollapsibleTheme = theme.CollapsibleTheme

// SelectTheme contains the trigger, panel, and animation settings for selects.
type SelectTheme = theme.SelectTheme

// PopoverTheme contains the layout and animation settings for popovers.
type PopoverTheme = theme.PopoverTheme

// TooltipTheme contains the layout, timing, and animation settings for tooltips.
type TooltipTheme = theme.TooltipTheme

// ToastTheme contains the layout, timing, and animation settings for toasts.
type ToastTheme = theme.ToastTheme

// ModalTheme contains the sizing, spacing, and animation settings for modals.
type ModalTheme = theme.ModalTheme

// ComboBoxTheme contains the sizing and item layout settings for combo boxes.
type ComboBoxTheme = theme.ComboBoxTheme

// DatePickerTheme contains the field and calendar settings for date pickers.
type DatePickerTheme = theme.DatePickerTheme

// ColorPickerTheme contains the trigger and panel settings for color pickers.
type ColorPickerTheme = theme.ColorPickerTheme

// ColorWheelTheme contains the sizing and focus settings for color wheels.
type ColorWheelTheme = theme.ColorWheelTheme

// ColorAreaTheme contains the sizing and focus settings for color areas.
type ColorAreaTheme = theme.ColorAreaTheme

// ColorFieldTheme contains the spacing settings for color fields.
type ColorFieldTheme = theme.ColorFieldTheme

// ColorSliderTheme contains the sizing and focus settings for color sliders.
type ColorSliderTheme = theme.ColorSliderTheme

// ColorSwatchTheme contains the size and shape settings for color swatches.
type ColorSwatchTheme = theme.ColorSwatchTheme

// ColorSwatchPickerTheme contains the layout and selection settings for swatch pickers.
type ColorSwatchPickerTheme = theme.ColorSwatchPickerTheme

// ShadowLayerCount is the number of layers in each shadow profile.
const ShadowLayerCount = theme.ShadowLayerCount

// DefaultTheme returns FlowUI's default light theme.
func DefaultTheme() Theme {
	return theme.DefaultTheme()
}

// DarkTheme returns FlowUI's default dark theme.
func DarkTheme() Theme {
	return theme.DarkTheme()
}

// DefaultShadows returns FlowUI's default shadow profiles.
func DefaultShadows() ShadowsTheme {
	return theme.DefaultShadows()
}

func syncMaterialTheme(activeTheme *Theme) {
	theme.SyncMaterialTheme(activeTheme)
}

// MaterialOf returns the Gio Material theme bridge associated with t, if any.
func MaterialOf(t *Theme) *material.Theme {
	return theme.MaterialOf(t)
}

// DetachMaterial removes the Material theme link from a FlowUI theme.
func DetachMaterial(t *Theme) {
	theme.DetachMaterial(t)
}

// SyncMaterialTheme synchronizes Material theme values into a FlowUI theme.
func SyncMaterialTheme(t *Theme) {
	theme.SyncMaterialTheme(t)
}
