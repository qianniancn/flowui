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
	window              []app.Option
	themeOps            []func(*Theme)
	language            Language
	errorHandler        func(error)
	closeRequestHandler func() WindowCloseDecision
	retainModel         bool
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
	if MaterialOf(&theme) == nil {
		SyncMaterialTheme(&theme)
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
			if MaterialOf(current) != nil {
				DetachMaterial(current)
			}
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
				if MaterialOf(theme) == nil {
					SyncMaterialTheme(theme)
				}
				fn(MaterialOf(theme))
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

// OnError handles command, subscription, synchronous runtime, and window
// lifecycle errors. Effect errors are delivered on the window event thread;
// synchronous callback panics are returned as *RuntimePanicError values.
func OnError(handler func(error)) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.errorHandler = handler
	})
}

// OnWindowCloseRequest handles FlowUI close requests for this window. The
// handler is used by Application.RequestClose and by WindowTitleBar's default
// close button. Application.Close and Application.Quit remain force-close
// operations and do not call it.
//
// Gio does not expose a cancellable native close-request event. Native title
// bar close buttons, Alt+F4, and window-manager close commands may therefore
// destroy the window without calling this handler.
func OnWindowCloseRequest(handler func() WindowCloseDecision) Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.closeRequestHandler = handler
	})
}

// RetainModelOnClose keeps the latest MVU model in the WindowSpec after its
// native window is destroyed. Reopening the same WindowSpec resumes from that
// model without rerunning its initializer or initial command. Subscriptions
// are recreated for each native window instance.
//
// The model is retained by assignment; FlowUI does not deep-copy it.
func RetainModelOnClose() Option {
	return optionFunc(func(cfg *runOptions) {
		cfg.retainModel = true
	})
}
