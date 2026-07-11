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

func TestThemeControlsToggleButtonHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.ToggleButton.MediumHeight = 42
	dims := ToggleButton("pin", false, Text("Pin")).Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size.Y != 42 {
		t.Fatalf("toggle button height = %d, want 42", dims.Size.Y)
	}
}

func TestThemeControlsCloseButtonSize(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.CloseButton.Size = 30
	dims := CloseButton("close").Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size != image.Pt(30, 30) {
		t.Fatalf("close button size = %v, want (30,30)", dims.Size)
	}
}

func TestThemeControlsCardSpacing(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Card.Padding = 20
	dims := Card(Spacer(40, 10)).Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size != image.Pt(80, 50) {
		t.Fatalf("card size = %v, want (80,50)", dims.Size)
	}
}

func TestThemeControlsSliderThickness(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Slider.TrackThickness = 28
	dims := Slider("volume", 30).Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size.Y != 28 {
		t.Fatalf("slider height = %d, want 28", dims.Size.Y)
	}
}

func TestThemeControlsSpinnerSize(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Spinner.MediumSize = 30
	dims := Spinner().Layout(themeTestContext(&theme), testLayoutContext())
	if dims.Size != image.Pt(30, 30) {
		t.Fatalf("spinner size = %v, want (30,30)", dims.Size)
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
