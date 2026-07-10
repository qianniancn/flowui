package flowui

import (
	"image"
	"image/color"
	"testing"
	"time"
)

func TestThemeDisabledColorUsesOpacity(t *testing.T) {
	theme := DefaultTheme()
	theme.DisabledOpacity = 0.25

	got := theme.DisabledColor(color.NRGBA{R: 1, G: 2, B: 3, A: 200})
	want := color.NRGBA{R: 1, G: 2, B: 3, A: 50}
	if got != want {
		t.Fatalf("disabled color = %#v, want %#v", got, want)
	}
}

func TestThemeDisabledColorAllowsZeroOpacity(t *testing.T) {
	theme := DefaultTheme()
	theme.DisabledOpacity = 0

	got := theme.DisabledColor(color.NRGBA{R: 1, G: 2, B: 3, A: 200})
	if got.A != 0 {
		t.Fatalf("disabled alpha = %d, want 0", got.A)
	}
}

func TestThemeDisabledOpacityIsClamped(t *testing.T) {
	theme := DefaultTheme()
	theme.DisabledOpacity = 2
	if got := theme.DisabledColor(color.NRGBA{A: 200}).A; got != 200 {
		t.Fatalf("disabled alpha = %d, want 200", got)
	}

	theme.DisabledOpacity = -1
	if got := theme.DisabledColor(color.NRGBA{A: 200}).A; got != 0 {
		t.Fatalf("disabled alpha = %d, want 0", got)
	}
}

func TestDefaultThemeSyncsMaterialBridge(t *testing.T) {
	theme := DefaultTheme()

	if theme.Material == nil {
		t.Fatal("missing material bridge")
	}
	if theme.Material.Palette.ContrastBg != theme.Palette.Accent {
		t.Fatalf("material accent = %#v, want %#v", theme.Material.Palette.ContrastBg, theme.Palette.Accent)
	}
	if theme.Material.Palette.Fg != theme.Palette.Foreground {
		t.Fatalf("material foreground = %#v, want %#v", theme.Material.Palette.Fg, theme.Palette.Foreground)
	}
}

func TestDarkThemeDefinesThemedSurfaceAndShadow(t *testing.T) {
	theme := DarkTheme()
	if theme.Palette.Surface == DefaultTheme().Palette.Surface {
		t.Fatal("dark theme did not override surface")
	}
	if theme.Palette.Shadow == DefaultTheme().Palette.Shadow {
		t.Fatal("dark theme did not override shadow")
	}
}

func TestThemeControlsButtonHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Spacing.ControlHeight = 52
	ctx := newContextWithTheme(nil, &theme)

	dims := Button("save", Text("Save")).Layout(ctx, testLayoutContext())
	if dims.Size.Y != 52 {
		t.Fatalf("button height = %d, want 52", dims.Size.Y)
	}
}

func TestThemeControlsCheckboxSize(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Checkbox.Size = 22
	theme.Components.Checkbox.FocusSpace = 3
	ctx := newContextWithTheme(nil, &theme)

	dims := Checkbox("done", false, "").Layout(ctx, testLayoutContext())
	if dims.Size != image.Pt(28, 28) {
		t.Fatalf("checkbox size = %v, want (28,28)", dims.Size)
	}
}

func TestThemeControlsSwitchSize(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Switch.MediumTrackWidth = 52
	theme.Components.Switch.MediumTrackHeight = 24
	theme.Components.Switch.FocusSpace = 3
	ctx := newContextWithTheme(nil, &theme)

	dims := Switch("notifications", false, "").Layout(ctx, testLayoutContext())
	if dims.Size != image.Pt(58, 30) {
		t.Fatalf("switch size = %v, want (58,30)", dims.Size)
	}
}

func TestThemeControlsComboBoxHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.ComboBox.Height = 48
	ctx := newContextWithTheme(nil, &theme)

	dims := ComboBox("animal", "", comboBoxTestItems()).Layout(ctx, testLayoutContext())
	if dims.Size.Y != 48 {
		t.Fatalf("combobox height = %d, want 48", dims.Size.Y)
	}
}

func TestThemeControlsSelectHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Select.Height = 44
	ctx := newContextWithTheme(nil, &theme)

	dims := Select("language", "", selectTestItems()).Layout(ctx, testLayoutContext())

	if dims.Size.Y != 44 {
		t.Fatalf("select height = %d, want 44", dims.Size.Y)
	}
}

func TestThemeControlsDatePickerHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.DatePicker.Height = 44
	ctx := newContextWithTheme(nil, &theme)

	dims := DatePicker("date", time.Time{}).Layout(ctx, testLayoutContext())
	if dims.Size.Y != 44 {
		t.Fatalf("date picker height = %d, want 44", dims.Size.Y)
	}
}

func TestComboBoxItemStyleUsesThemePalette(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.SurfaceRaised = color.NRGBA{R: 7, G: 8, B: 9, A: 255}

	style := comboBoxItemStyleFor(&theme, true, false, false, false)
	if style.bg != theme.Palette.SurfaceRaised {
		t.Fatalf("combobox item background = %#v, want %#v", style.bg, theme.Palette.SurfaceRaised)
	}
}

func TestDatePickerCellStyleUsesThemePalette(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.Accent = color.NRGBA{R: 7, G: 8, B: 9, A: 255}
	theme.Palette.AccentForeground = color.NRGBA{R: 250, G: 251, B: 252, A: 255}

	style := datePickerCellStyleFor(&theme, false, false, true, false, false, false)
	if style.bg != theme.Palette.Accent {
		t.Fatalf("date picker selected background = %#v, want %#v", style.bg, theme.Palette.Accent)
	}
	if style.fg != theme.Palette.AccentForeground {
		t.Fatalf("date picker selected foreground = %#v, want %#v", style.fg, theme.Palette.AccentForeground)
	}
}
