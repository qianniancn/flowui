package ui

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/app"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func TestMaterialThemeOption(t *testing.T) {
	want := color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}
	cfg := newRunOptions([]Option{
		MaterialTheme(func(th *material.Theme) {
			th.Palette.ContrastBg = want
		}),
	})

	theme := cfg.newTheme()
	got := MaterialOf(theme).Palette.ContrastBg
	if got != want {
		t.Fatalf("contrast background = %#v, want %#v", got, want)
	}
}

func TestWithThemeOption(t *testing.T) {
	want := color.NRGBA{R: 0x22, G: 0x88, B: 0xdd, A: 0xff}
	theme := DefaultTheme()
	sourceMaterial := MaterialOf(&theme)
	theme.Palette.Accent = want
	cfg := newRunOptions([]Option{
		WithTheme(theme),
	})

	first := cfg.newTheme()
	second := cfg.newTheme()
	if first.Palette.Accent != want {
		t.Fatalf("accent = %#v, want %#v", first.Palette.Accent, want)
	}
	if MaterialOf(first) == sourceMaterial || MaterialOf(second) == sourceMaterial || MaterialOf(first) == MaterialOf(second) {
		t.Fatal("WithTheme reused a material theme between source or window instances")
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
	if MaterialOf(theme).Palette.ContrastBg != want {
		t.Fatalf("material contrast background = %#v, want %#v", MaterialOf(theme).Palette.ContrastBg, want)
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
}

func TestWindowOptions(t *testing.T) {
	cfg := newRunOptions([]Option{
		Title("FlowUI"),
		Size(640, 480),
		MinSize(320, 240),
		MaxSize(1280, 960),
		TopMost(true),
		Decorated(false),
	})
	if len(cfg.window) != 6 {
		t.Fatalf("window options = %d, want 6", len(cfg.window))
	}
	config := app.Config{Decorated: true}
	metric := unit.Metric{PxPerDp: 1, PxPerSp: 1}
	for _, option := range cfg.window {
		option(metric, &config)
	}
	if config.Title != "FlowUI" || config.Size != image.Pt(640, 480) || config.MinSize != image.Pt(320, 240) || config.MaxSize != image.Pt(1280, 960) || !config.TopMost || config.Decorated {
		t.Fatalf("window config = %#v", config)
	}
}

func TestOnErrorOption(t *testing.T) {
	handler := func(error) {}
	cfg := newRunOptions([]Option{OnError(handler)})
	if cfg.errorHandler == nil {
		t.Fatal("error handler was not configured")
	}
}

func TestRetainModelOnCloseOption(t *testing.T) {
	cfg := newRunOptions([]Option{RetainModelOnClose()})
	if !cfg.retainModel {
		t.Fatal("model retention was not configured")
	}
}

func TestWindowCloseRequestOption(t *testing.T) {
	handler := func() WindowCloseDecision { return WindowCloseCancel }
	cfg := newRunOptions([]Option{OnWindowCloseRequest(handler)})
	if cfg.closeRequestHandler == nil || cfg.closeRequestHandler() != WindowCloseCancel {
		t.Fatal("window close-request handler was not configured")
	}
}
