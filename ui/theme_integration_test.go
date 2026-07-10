package ui

import (
	"image"
	"testing"
	"time"

	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestThemeControlsButtonHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Spacing.ControlHeight = 52
	dims := Button("save", Text("Save")).Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size.Y != 52 {
		t.Fatalf("button height = %d, want 52", dims.Size.Y)
	}
}

func TestThemeControlsCheckboxSize(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Checkbox.Size = 22
	theme.Components.Checkbox.FocusSpace = 3
	dims := Checkbox("done", false, "").Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size != image.Pt(28, 28) {
		t.Fatalf("checkbox size = %v, want (28,28)", dims.Size)
	}
}

func TestThemeControlsSwitchSize(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Switch.MediumTrackWidth = 52
	theme.Components.Switch.MediumTrackHeight = 24
	theme.Components.Switch.FocusSpace = 3
	dims := Switch("notifications", false, "").Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size != image.Pt(58, 30) {
		t.Fatalf("switch size = %v, want (58,30)", dims.Size)
	}
}

func TestThemeControlsComboBoxHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.ComboBox.Height = 48
	dims := ComboBox("animal", "", []ComboBoxItem{{Key: "cat", Label: "Cat"}}).
		Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size.Y != 48 {
		t.Fatalf("combobox height = %d, want 48", dims.Size.Y)
	}
}

func TestThemeControlsSelectHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Select.Height = 44
	dims := Select("language", "", []SelectItem{{Key: "go", Label: "Go"}}).
		Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size.Y != 44 {
		t.Fatalf("select height = %d, want 44", dims.Size.Y)
	}
}

func TestThemeControlsDatePickerHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.DatePicker.Height = 44
	dims := DatePicker("date", time.Time{}).Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size.Y != 44 {
		t.Fatalf("date picker height = %d, want 44", dims.Size.Y)
	}
}

func themeTestContext(theme *Theme) *Context {
	return frame.New(nil, theme, LanguageAuto)
}
