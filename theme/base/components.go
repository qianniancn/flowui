package base

import "github.com/qianniancn/flowui/theme"

// Keep the default design here so external themes can still start from the
// framework board without inheriting this design's density and corner sizes.
func configureComponents(c *theme.ComponentsTheme) {
	c.Surface.BorderWidth = 1
	c.ToggleButton.SmallHeight, c.ToggleButton.MediumHeight, c.ToggleButton.LargeHeight = 24, 32, 40
	c.ToggleButton.SmallPaddingX, c.ToggleButton.MediumPaddingX, c.ToggleButton.LargePaddingX = 8, 12, 16
	c.ToggleButton.Radius = 4
	c.ToggleButton.PressedScaleSmall, c.ToggleButton.PressedScaleMedium, c.ToggleButton.PressedScaleLarge = 1, 1, 1
	c.CloseButton.Radius, c.CloseButton.PressedScale = 4, 1
	c.Chip.Radius = 4
	c.Badge.SmallRadius, c.Badge.MediumRadius, c.Badge.LargeRadius = 3, 4, 6
	c.Card.Radius, c.Card.Padding, c.Card.Gap = 8, 16, 16
	c.Alert.Radius, c.Alert.PaddingX, c.Alert.PaddingY, c.Alert.Gap = 8, 12, 12, 12
	c.AlertDialog.TitleSize = 20

	c.Input.Height, c.Input.SmallHeight, c.Input.LargeHeight = 32, 24, 40
	c.Input.Radius, c.Input.SmallRadius, c.Input.LargeRadius = 4, 4, 6
	c.Input.PaddingX, c.Input.SmallPaddingX, c.Input.LargePaddingX = 8, 6, 12
	c.Input.BorderWidth, c.Input.FocusGlowBlur = 1, 0
	c.Input.FocusHaloBlur, c.Input.FocusHaloOpacity = 5, .60
	c.Input.FocusShadowSpread, c.Input.ShadowOpacity = 0, 0
	c.Input.FocusRingWidth, c.Input.InvalidOutlineWidth = 1, 0
	c.TextArea.MinHeight, c.TextArea.Radius = 64, 4
	c.TextArea.PaddingX, c.TextArea.PaddingY = 8, 8
	c.TextArea.BorderWidth, c.TextArea.FocusGlowBlur = 1, 0
	c.TextArea.FocusHaloBlur, c.TextArea.FocusHaloOpacity = 5, .60
	c.TextArea.FocusShadowSpread, c.TextArea.ShadowOpacity = 0, 0
	c.TextArea.FocusRingWidth, c.TextArea.InvalidOutlineWidth = 1, 0
	c.InputGroup.MinHeight, c.InputGroup.Radius, c.InputGroup.PaddingX = 32, 4, 8
	c.InputGroup.TextAreaMinHeight, c.InputGroup.TextAreaPaddingY = 64, 8
	c.InputGroup.BorderWidth, c.InputGroup.ShadowOpacity = 1, 0
	c.InputGroup.FocusGlowBlur = 0
	c.InputGroup.FocusHaloBlur, c.InputGroup.FocusHaloOpacity = 5, .60
	c.InputGroup.FocusShadowSpread = 0
	c.InputGroup.FocusRingWidth, c.InputGroup.InvalidOutlineWidth = 1, 0
	c.Checkbox.Size, c.Checkbox.ShadowOpacity = 16, 0
	c.Switch.SmallTrackWidth, c.Switch.SmallTrackHeight = 28, 16
	c.Switch.SmallThumbWidth, c.Switch.SmallThumbHeight = 14, 14
	c.Switch.MediumTrackWidth, c.Switch.MediumTrackHeight = 35, 20
	c.Switch.MediumThumbWidth, c.Switch.MediumThumbHeight = 18, 18
	c.Switch.LargeTrackWidth, c.Switch.LargeTrackHeight = 42, 24
	c.Switch.LargeThumbWidth, c.Switch.LargeThumbHeight = 22, 22
	c.Switch.ContentGap = 8
	c.SwitchGroup.VerticalGap, c.SwitchGroup.HorizontalGap = 10, 14
	c.RadioGroup.Size, c.RadioGroup.ContentGap, c.RadioGroup.VerticalGap = 16, 8, 8
	c.RadioGroup.PressedScale = 1
	c.ProgressBar.MediumHeight, c.ProgressBar.LargeHeight = 8, 12
	c.ProgressBar.MediumRadius, c.ProgressBar.LargeRadius = 4, 6
	c.Slider.TrackThickness, c.Slider.TrackRadius, c.Slider.EdgeInset = 8, 4, 8
	c.Slider.ThumbLength, c.Slider.ThumbExtra, c.Slider.DraggingScale = 16, 0, 1
	c.Slider.ThumbSize = 16

	c.ListBox.Gap, c.ListBox.ItemMinHeight, c.ListBox.ItemRadius, c.ListBox.ItemPaddingY = 0, 32, 4, 6
	c.ListBox.PressedScale = 1
	c.Tree.Gap, c.Tree.RowHeight, c.Tree.RowRadius, c.Tree.RowPaddingY = 2, 32, 4, 5
	c.Tree.SurfaceRadius = 8
	c.Sidebar.ItemHeight, c.Sidebar.ItemRadius, c.Sidebar.ItemGap = 32, 4, 2
	c.Scrollbar.TrackWidth, c.Scrollbar.ThumbWidth, c.Scrollbar.MinThumbLength = 9, 7, 20
	c.Scrollbar.Radius, c.Scrollbar.ThumbOpacity, c.Scrollbar.HoverOpacity = 3.5, .25, .4
	c.Toolbar.Radius, c.Toolbar.Gap = 4, 8
	c.Table.RootRadius, c.Table.HeaderRadius, c.Table.BodyRadius, c.Table.FocusRadius = 8, 0, 0, 4
	c.Table.RootPadding = 1
	c.Table.HeaderHeight, c.Table.RowMinHeight = 44, 44
	c.Table.CellPaddingX, c.Table.CellPaddingY = 12, 8
	c.Table.FooterPaddingX, c.Table.FooterPaddingY = 12, 8
	c.Pagination.Radius = 4
	c.Menu.Radius, c.Menu.BorderWidth, c.Menu.ItemMinHeight, c.Menu.ItemRadius = 8, 1, 32, 4
	c.Menu.ItemPaddingY, c.Menu.PressedScale, c.Menu.EnterScale, c.Menu.ExitScale = 5, 1, 1, 1
	c.Dropdown.TriggerFocusRadius, c.Dropdown.TriggerPressedScale = 4, 1
	c.Menubar.TriggerRadius = 4
	c.Tabs.ListRadius, c.Tabs.IndicatorRadius, c.Tabs.IndicatorLineWidth = 4, 3, 2
	c.Tabs.SmallTabHeight, c.Tabs.TabHeight, c.Tabs.LargeTabHeight = 24, 32, 40
	c.Tabs.TabPaddingX, c.Tabs.LargeTabPaddingX = 12, 16
	c.Select.Height, c.Select.Radius, c.Select.TriggerPaddingX, c.Select.TriggerPaddingY = 32, 4, 8, 5
	c.Select.PanelRadius, c.Select.PanelPadding, c.Select.PanelBorderWidth, c.Select.AnimationScale = 8, 4, 1, 1
	c.Popover.Radius, c.Popover.Padding, c.Popover.AnimationScale = 8, 16, 1
	c.Tooltip.Radius, c.Tooltip.AnimationScale, c.Tooltip.ExitScale = 4, 1, 1
	c.Toast.Radius, c.Toast.PaddingX, c.Toast.PaddingY, c.Toast.ScaleFactor = 8, 16, 12, 0
	c.Modal.Radius, c.Modal.Padding, c.Modal.TitleSize = 12, 24, 20
	c.Modal.AnimationScale, c.Modal.AnimationBounceScale, c.Modal.AnimationBounceSettleScale = 1, 1, 1
	c.Modal.AnimationZoomScale, c.Modal.AnimationPopStartScale = 1, 1
	c.Modal.AnimationPopOvershootScale, c.Modal.AnimationPopSettleScale = 1, 1
	c.ComboBox.Height, c.ComboBox.Radius, c.ComboBox.TriggerWidth = 32, 4, 28
	c.ComboBox.PanelRadius, c.ComboBox.PanelPadding, c.ComboBox.PanelBorderWidth = 8, 4, 1
	c.ComboBox.ItemRadius, c.ComboBox.ItemHeight = 4, 32
	c.ComboBox.ItemPaddingY, c.ComboBox.AnimationScale = 5, 1
	c.DatePicker.Height, c.DatePicker.Radius, c.DatePicker.TriggerWidth = 32, 4, 28
	c.DatePicker.RangeRadius, c.DatePicker.SegmentRadius, c.DatePicker.PopoverRadius = 4, 4, 8
	c.DatePicker.PopoverBorderWidth = 1
	c.DatePicker.HeaderTextSize = 15
	c.DatePicker.WeekdayTextSize = 11
	c.ColorPicker.PanelRadius, c.ColorPicker.PanelPadding, c.ColorPicker.EnterScale = 8, 16, 1
	c.ColorArea.Radius = 4
	c.ColorSwatchPicker.SquareItemRadiusSmall, c.ColorSwatchPicker.SquareItemRadius = 4, 6
	c.ColorSwatchPicker.SquareSwatchRadiusSmall, c.ColorSwatchPicker.SquareSwatchRadius = 4, 4
	c.ColorSwatchPicker.ShadowOpacity = 0
}
