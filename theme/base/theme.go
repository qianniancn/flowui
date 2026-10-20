// Package base provides FlowUI's default theme and complete shared visual Provider.
package base

import (
	"runtime"
	"time"

	"github.com/qianniancn/flowui/theme"
)

// Light returns the default light theme, ready for window.Theme.
func Light() theme.Theme { return New(defaultSeed()) }

// Dark returns the default dark theme, ready for window.Theme.
func Dark() theme.Theme { return NewDark(defaultSeed()) }

// New creates a light theme with the shared component recipes installed.
// Replace VisualProvider with a type embedding Provider to customize recipes.
func New(seed theme.Seed) theme.Theme {
	active := theme.New(seed)
	configure(&active, seed, false)
	return active
}

// NewDark creates a dark theme with the shared component recipes installed.
func NewDark(seed theme.Seed) theme.Theme {
	active := theme.NewDark(seed)
	configure(&active, seed, true)
	return active
}

// configureTypography names real installed families. Gio does not map CSS
// generic names such as "sans-serif" or "monospace", so a generic name on its
// own falls back to an arbitrary face, and a monospace block can lose its CJK
// glyphs entirely. Each list puts the UI face first and a CJK face after it so
// fallback happens per glyph.
func configureTypography(typography *theme.Typography) {
	switch runtime.GOOS {
	case "windows":
		// Generic sans-serif can choose a CJK face with excessive line gap on
		// Windows, especially at medium weight. Prefer matching UI families for
		// Latin and Chinese so controls center a consistent font line box.
		typography.Typeface = "Segoe UI, Microsoft YaHei UI"
		typography.MonoTypeface = "Cascadia Mono, Consolas, Microsoft YaHei UI"
	case "darwin":
		typography.Typeface = "Helvetica Neue, PingFang SC"
		typography.MonoTypeface = "Menlo, PingFang SC"
	default:
		typography.Typeface = "Noto Sans, DejaVu Sans, Noto Sans CJK SC"
		typography.MonoTypeface = "DejaVu Sans Mono, Noto Sans Mono, Noto Sans CJK SC"
	}
}

func defaultSeed() theme.Seed {
	return theme.Seed{
		Primary: rgb(0x3e63dd), Danger: rgb(0xe5484d),
		Success: rgb(0x30a46c), Warning: rgb(0xffc53d), Info: rgb(0x3e63dd),
		Motion: theme.MotionSeed{Enabled: true, Duration: 140 * time.Millisecond},
	}
}

func configure(active *theme.Theme, seed theme.Seed, dark bool) {
	active.UseSemantic(semanticColors(active.Semantic(), seed, dark))
	configureTypography(&active.Typography)
	configureComponents(&active.Components)
	if dark {
		active.Components.DatePicker.CurrentTextColor = rgb(0xffffff)
		active.Components.DatePicker.OutsideTextColor = rgb(0x696e77)
	} else {
		active.Components.DatePicker.CurrentTextColor = rgb(0x000000)
		active.Components.DatePicker.OutsideTextColor = rgb(0x8b8d98)
	}
	active.Shadows = theme.ShadowsTheme{
		Overlay: theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
			{OffsetY: 4, Blur: 8, Spread: -2, Opacity: .4},
			{OffsetY: 12, Blur: 32, Spread: -4, Opacity: 1},
		}},
		Menu: theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
			{Blur: 0, Spread: 1, Opacity: .60},
			{OffsetY: 6, Blur: 16, Opacity: .70},
			{OffsetY: 16, Blur: 40, Opacity: .55},
		}},
	}
	active.Components.Menu.ShadowColor = active.Semantic().OverlayShadow
	if dark {
		active.Components.Menu.ShadowColor.A = 140
	}
	active.Components.Menu.BorderColor = active.Semantic().Border
	active.VisualProvider = Provider{}
	theme.SyncMaterialTheme(active)
}
