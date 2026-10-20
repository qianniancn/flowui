package window

import (
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

// Option configures a FlowUI window.
type Option interface {
	apply(*Config)
}

// Config is the resolved window configuration consumed by the runtime.
type Config struct {
	Native         []app.Option
	InitialActions system.Action
	ThemeOps       []func(*theme.Theme)
	// InitialTheme is the explicitly selected visual provider used before the
	// queued theme operations are applied. ThemeSet distinguishes an explicit
	// zero-value theme from a missing theme; windows must not silently fall back
	// to a framework default appearance.
	InitialTheme        theme.Theme
	ThemeSet            bool
	Language            Language
	Translations        i18n.Catalog
	ErrorHandler        func(error)
	CloseRequestHandler func() CloseDecision
	RetainModel         bool
}

// Language identifies the language used by localized widgets.
type Language = locale.Language

const (
	LanguageAuto    = locale.LanguageAuto
	LanguageEnglish = locale.LanguageEnglish
	LanguageChinese = locale.LanguageChinese
)

// Action is a native action requested for a window.
type Action = system.Action

const (
	ActionMinimize   = system.ActionMinimize
	ActionMaximize   = system.ActionMaximize
	ActionRestore    = system.ActionUnmaximize
	ActionFullscreen = system.ActionFullscreen
	ActionRaise      = system.ActionRaise
	ActionCenter     = system.ActionCenter
)

// CloseDecision controls a close request from the native window or FlowUI.
type CloseDecision uint8

const (
	CloseProceed CloseDecision = iota
	CloseCancel
	CloseKeepAlive
)

type optionFunc func(*Config)

func (fn optionFunc) apply(cfg *Config) { fn(cfg) }

type nativeOption struct{ value app.Option }

func (option nativeOption) apply(cfg *Config) {
	cfg.Native = append(cfg.Native, option.value)
}

// Resolve applies options in declaration order and returns the resulting
// configuration. A nil option is ignored.
func Resolve(options ...Option) Config {
	var cfg Config
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

// Title sets the native window title.
func Title(value string) Option { return nativeOption{value: app.Title(value)} }

// Size sets the initial window size in dp.
func Size(width, height int) Option {
	return nativeOption{value: app.Size(unit.Dp(width), unit.Dp(height))}
}

// MinSize sets the minimum window size in dp.
func MinSize(width, height int) Option {
	return nativeOption{value: app.MinSize(unit.Dp(width), unit.Dp(height))}
}

// MaxSize sets the maximum window size in dp.
func MaxSize(width, height int) Option {
	return nativeOption{value: app.MaxSize(unit.Dp(width), unit.Dp(height))}
}

// TopMost controls whether the window stays above non-top-most windows.
func TopMost(enabled bool) Option { return nativeOption{value: app.TopMost(enabled)} }

// Decorated controls native or Gio-provided window decorations.
func Decorated(enabled bool) Option {
	return nativeOption{value: app.Decorated(enabled)}
}

// CenterOnStart centers the native window once after it is created.
func CenterOnStart() Option {
	return optionFunc(func(cfg *Config) { cfg.InitialActions |= system.ActionCenter })
}

// Theme replaces the FlowUI theme used by widgets.
func Theme(value theme.Theme) Option {
	return optionFunc(func(cfg *Config) {
		cfg.InitialTheme = value
		cfg.ThemeSet = true
		cfg.ThemeOps = append(cfg.ThemeOps, func(current *theme.Theme) {
			*current = value
			theme.DetachMaterial(current)
			theme.SyncMaterialTheme(current)
		})
	})
}

// CustomizeTheme mutates the FlowUI theme used by widgets.
func CustomizeTheme(fn func(*theme.Theme)) Option {
	return optionFunc(func(cfg *Config) {
		if fn == nil {
			return
		}
		cfg.ThemeOps = append(cfg.ThemeOps, func(current *theme.Theme) {
			fn(current)
			theme.SyncMaterialTheme(current)
		})
	})
}

// Locale sets the application language used by localized widgets.
func Locale(language Language) Option {
	return optionFunc(func(cfg *Config) { cfg.Language = language })
}

// Translations adds application-owned message overrides for this window.
func Translations(catalog i18n.Catalog) Option {
	return optionFunc(func(cfg *Config) { cfg.Translations = catalog.Clone() })
}

// OnError handles command, subscription, runtime, and window lifecycle errors.
func OnError(handler func(error)) Option {
	return optionFunc(func(cfg *Config) { cfg.ErrorHandler = handler })
}

// OnWindowCloseRequest handles native and FlowUI close requests for this
// window. The handler runs synchronously on the window event loop and must
// return promptly; it cannot dispatch MVU messages. Use
// ui.Program.CloseRequestMessage when a native close request must open an
// in-app confirmation instead. Returning CloseProceed lets the close happen,
// CloseCancel aborts the native request, and CloseKeepAlive closes the native
// window while keeping the process alive. Unknown decisions fail closed.
func OnWindowCloseRequest(handler func() CloseDecision) Option {
	return optionFunc(func(cfg *Config) { cfg.CloseRequestHandler = handler })
}

// RetainModelOnClose keeps the latest MVU model in the WindowSpec after its
// native window is destroyed.
func RetainModelOnClose() Option {
	return optionFunc(func(cfg *Config) { cfg.RetainModel = true })
}
