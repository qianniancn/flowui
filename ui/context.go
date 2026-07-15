package ui

import "github.com/qianniancn/FlowUI/internal/frame"

type Context = frame.Context
type WindowState = frame.WindowState
type WindowMode = frame.WindowMode

const (
	WindowModeWindowed   = frame.Windowed
	WindowModeFullscreen = frame.Fullscreen
	WindowModeMinimized  = frame.Minimized
	WindowModeMaximized  = frame.Maximized
)
