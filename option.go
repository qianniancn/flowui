package flowui

import (
	"gioui.org/app"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Option configures Run and RunCmd.
type Option interface {
	apply(*runOptions)
}

type optionFunc func(*runOptions)

func (fn optionFunc) apply(cfg *runOptions) {
	fn(cfg)
}

type runOptions struct {
	window   []app.Option
	themeOps []func(*Theme)
	language Language
}

func newRunOptions(opts []Option) runOptions {
	var cfg runOptions
	for _, opt := range opts {
		if opt != nil {
			opt.apply(&cfg)
		}
	}
	return cfg
}

func (cfg runOptions) newTheme() *Theme {
	theme := DefaultTheme()
	for _, fn := range cfg.themeOps {
		fn(&theme)
	}
	if theme.Material == nil {
		theme.Material = material.NewTheme()
	}
	return &theme
}

func (cfg runOptions) datePickerLocale() DatePickerLocale {
	return datePickerLocaleForLanguage(cfg.language)
}

// Title sets the window title.
func Title(title string) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.window = append(cfg.window, app.Title(title))
	})
}

// Size sets the initial window size in dp.
func Size(width, height int) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.window = append(cfg.window, app.Size(unit.Dp(width), unit.Dp(height)))
	})
}

// WithTheme replaces the FlowUI theme used by widgets.
func WithTheme(theme Theme) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.themeOps = append(cfg.themeOps, func(current *Theme) {
			*current = theme
			syncMaterialTheme(current)
		})
	})
}

// CustomizeTheme mutates the FlowUI theme used by widgets.
func CustomizeTheme(fn func(*Theme)) Option {
	return optionFunc(func(cfg *runOptions) {
		if fn != nil {
			cfg.themeOps = append(cfg.themeOps, func(theme *Theme) {
				fn(theme)
				syncMaterialTheme(theme)
			})
		}
	})
}

// MaterialTheme mutates the Gio material theme bridge used by editor/text internals.
func MaterialTheme(fn func(*material.Theme)) Option {
	return optionFunc(func(cfg *runOptions) {
		if fn != nil {
			cfg.themeOps = append(cfg.themeOps, func(theme *Theme) {
				if theme.Material == nil {
					theme.Material = material.NewTheme()
				}
				fn(theme.Material)
			})
		}
	})
}

// Locale sets the application language used by localized FlowUI widgets.
func Locale(language Language) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.language = language
	})
}
