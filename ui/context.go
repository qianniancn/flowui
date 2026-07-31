package ui

import "github.com/qianniancn/flowui/internal/frame"

// Context provides per-frame services to FlowUI widgets.
type Context = frame.Context

// WindowState describes the latest native window configuration.
type WindowState = frame.WindowState

// WindowMode identifies the current native window mode.
type WindowMode = frame.WindowMode

const (
	// WindowModeWindowed is the normal resizable window mode.
	WindowModeWindowed = frame.Windowed
	// WindowModeFullscreen indicates a borderless fullscreen window.
	WindowModeFullscreen = frame.Fullscreen
	// WindowModeMinimized indicates that the native window is minimized.
	WindowModeMinimized = frame.Minimized
	// WindowModeMaximized indicates that the native window is maximized.
	WindowModeMaximized = frame.Maximized
)
