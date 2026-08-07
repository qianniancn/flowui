package ui

import (
	"image"
	"testing"
	"time"

	"github.com/qianniancn/flowui/internal/frame"
)

func TestThemeControlsComponentGeometry(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*Theme)
		widget    func() Widget
		want      image.Point
	}{
		{
			name:      "button",
			configure: func(theme *Theme) { theme.Spacing.ControlHeight = 52 },
			widget:    func() Widget { return Button("save", Text("Save")) },
			want:      image.Pt(0, 52),
		},
		{
			name:      "toggle button",
			configure: func(theme *Theme) { theme.Components.ToggleButton.MediumHeight = 42 },
			widget:    func() Widget { return ToggleButton("pin", false, Text("Pin")) },
			want:      image.Pt(0, 42),
		},
		{
			name:      "input",
			configure: func(theme *Theme) { theme.Components.Input.Height = 42 },
			widget:    func() Widget { return Input("name", "").Placeholder("Name") },
			want:      image.Pt(0, 42),
		},
		{
			name:      "close button",
			configure: func(theme *Theme) { theme.Components.CloseButton.Size = 30 },
			widget:    func() Widget { return CloseButton("close") },
			want:      image.Pt(30, 30),
		},
		{
			name:      "card",
			configure: func(theme *Theme) { theme.Components.Card.Padding = 20 },
			widget:    func() Widget { return Card(Spacer(40, 10)) },
			want:      image.Pt(80, 50),
		},
		{
			name:      "slider",
			configure: func(theme *Theme) { theme.Components.Slider.TrackThickness = 28 },
			widget:    func() Widget { return Slider("volume", 30) },
			want:      image.Pt(0, 28),
		},
		{
			name:      "spinner",
			configure: func(theme *Theme) { theme.Components.Spinner.MediumSize = 30 },
			widget:    func() Widget { return Spinner() },
			want:      image.Pt(30, 30),
		},
		{
			name: "checkbox",
			configure: func(theme *Theme) {
				theme.Components.Checkbox.Size, theme.Components.Checkbox.FocusSpace = 22, 3
			},
			widget: func() Widget { return Checkbox("done", false, "") },
			want:   image.Pt(28, 28),
		},
		{
			name: "switch",
			configure: func(theme *Theme) {
				theme.Components.Switch.MediumTrackWidth = 52
				theme.Components.Switch.MediumTrackHeight = 24
				theme.Components.Switch.FocusSpace = 3
			},
			widget: func() Widget { return Switch("notifications", false, "") },
			want:   image.Pt(58, 30),
		},
		{
			name:      "combobox",
			configure: func(theme *Theme) { theme.Components.ComboBox.Height = 48 },
			widget: func() Widget {
				return ComboBox("animal", "", []ComboBoxItem{{Key: "cat", Label: "Cat"}})
			},
			want: image.Pt(0, 48),
		},
		{
			name:      "select",
			configure: func(theme *Theme) { theme.Components.Select.Height = 44 },
			widget: func() Widget {
				return Select("language", "", []SelectItem{{Key: "go", Label: "Go"}})
			},
			want: image.Pt(0, 44),
		},
		{
			name:      "date picker",
			configure: func(theme *Theme) { theme.Components.DatePicker.Height = 44 },
			widget:    func() Widget { return DatePicker("date", time.Time{}) },
			want:      image.Pt(0, 44),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			activeTheme := DefaultTheme()
			test.configure(&activeTheme)
			got := test.widget().Layout(themeTestContext(&activeTheme), testLayoutContext()).Size
			widthMatches := test.want.X == 0 || got.X == test.want.X
			heightMatches := test.want.Y == 0 || got.Y == test.want.Y
			if !widthMatches || !heightMatches {
				t.Fatalf("size = %v, want %v (zero axes are ignored)", got, test.want)
			}
		})
	}
}

func themeTestContext(theme *Theme) *Context {
	return frame.New(nil, theme, LanguageAuto)
}
