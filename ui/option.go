package ui

import (
	"gioui.org/app"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Option configures a FlowUI window.
type Option interface {
	apply(*runOptions)
}

// WindowOption configures native window properties at creation or runtime.
type WindowOption interface {
	Option
	appOption() app.Option
}

type optionFunc func(*runOptions)

func (fn optionFunc) apply(cfg *runOptions) {
	fn(cfg)
}

type windowOption struct {
	value app.Option
}

func (option windowOption) apply(cfg *runOptions) {
	cfg.window = append(cfg.window, option.value)
}

func (option windowOption) appOption() app.Option {
	return option.value
}

type runOptions struct {
	window       []app.Option
	themeOps     []func(*Theme)
	language     Language
	errorHandler func(error)
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

// Title sets the window title.
func Title(title string) WindowOption {
	return windowOption{value: app.Title(title)}
}

// Size sets the window size in dp.
func Size(width, height int) WindowOption {
	return windowOption{value: app.Size(unit.Dp(width), unit.Dp(height))}
}

// MinSize sets the minimum window size in dp.
func MinSize(width, height int) WindowOption {
	return windowOption{value: app.MinSize(unit.Dp(width), unit.Dp(height))}
}

// MaxSize sets the maximum window size in dp.
func MaxSize(width, height int) WindowOption {
	return windowOption{value: app.MaxSize(unit.Dp(width), unit.Dp(height))}
}

// TopMost controls whether the window stays above non-top-most windows. Gio
// supports top-most windows on macOS and Windows.
func TopMost(enabled bool) WindowOption {
	return windowOption{value: app.TopMost(enabled)}
}

// Decorated controls native or Gio-provided window decorations.
func Decorated(enabled bool) WindowOption {
	return windowOption{value: app.Decorated(enabled)}
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

// OnError handles command, subscription, and window lifecycle errors on the
// window event thread. Panics are recovered and reported as *EffectError values
// too.
func OnError(handler func(error)) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.errorHandler = handler
	})
}
