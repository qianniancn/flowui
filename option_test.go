package flowui

import (
	"image/color"
	"testing"

	"gioui.org/widget/material"
)

func TestMaterialThemeOption(t *testing.T) {
	want := color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}
	cfg := newRunOptions([]Option{
		MaterialTheme(func(th *material.Theme) {
			th.Palette.ContrastBg = want
		}),
	})

	got := cfg.newTheme().Material.Palette.ContrastBg
	if got != want {
		t.Fatalf("contrast background = %#v, want %#v", got, want)
	}
}

func TestWithThemeOption(t *testing.T) {
	want := color.NRGBA{R: 0x22, G: 0x88, B: 0xdd, A: 0xff}
	theme := DefaultTheme()
	theme.Palette.Accent = want
	cfg := newRunOptions([]Option{
		WithTheme(theme),
	})

	got := cfg.newTheme().Palette.Accent
	if got != want {
		t.Fatalf("accent = %#v, want %#v", got, want)
	}
}

func TestCustomizeThemeOption(t *testing.T) {
	want := color.NRGBA{R: 0x22, G: 0x88, B: 0xdd, A: 0xff}
	cfg := newRunOptions([]Option{
		CustomizeTheme(func(theme *Theme) {
			theme.Palette.Accent = want
		}),
	})

	theme := cfg.newTheme()
	if theme.Palette.Accent != want {
		t.Fatalf("accent = %#v, want %#v", theme.Palette.Accent, want)
	}
	if theme.Material.Palette.ContrastBg != want {
		t.Fatalf("material contrast background = %#v, want %#v", theme.Material.Palette.ContrastBg, want)
	}
}

func TestNilOptionsAreIgnored(t *testing.T) {
	cfg := newRunOptions([]Option{
		nil,
		CustomizeTheme(nil),
		MaterialTheme(nil),
	})
	if len(cfg.window) != 0 {
		t.Fatalf("window options = %d, want 0", len(cfg.window))
	}
	if len(cfg.themeOps) != 0 {
		t.Fatalf("theme options = %d, want 0", len(cfg.themeOps))
	}
}

func TestLocaleOption(t *testing.T) {
	cfg := newRunOptions([]Option{
		Locale(LanguageChinese),
	})

	if cfg.language != LanguageChinese {
		t.Fatalf("language = %q, want Chinese", cfg.language)
	}
	if got := cfg.datePickerLocale().WeekStart; got != 1 {
		t.Fatalf("date picker week start = %v, want Monday", got)
	}
}

func TestWindowOptions(t *testing.T) {
	cfg := newRunOptions([]Option{
		Title("FlowUI"),
		Size(640, 480),
	})
	if len(cfg.window) != 2 {
		t.Fatalf("window options = %d, want 2", len(cfg.window))
	}
}
