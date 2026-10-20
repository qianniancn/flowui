package ui

import "github.com/qianniancn/flowui/internal/core"

// Context provides per-frame services to FlowUI widgets.
type Context = core.Context

// WindowState describes the latest native window configuration.
type WindowState = core.WindowState

// WindowMode identifies the current native window mode.
type WindowMode = core.WindowMode

const (
	// WindowModeWindowed is the normal resizable window mode.
	WindowModeWindowed = core.Windowed
	// WindowModeFullscreen indicates a borderless fullscreen window.
	WindowModeFullscreen = core.Fullscreen
	// WindowModeMinimized indicates that the native window is minimized.
	WindowModeMinimized = core.Minimized
	// WindowModeMaximized indicates that the native window is maximized.
	WindowModeMaximized = core.Maximized
)
