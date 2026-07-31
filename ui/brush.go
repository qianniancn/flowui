package ui

import "github.com/qianniancn/flowui/internal/style"

// LinearGradient creates a linear gradient from its ordered stops.
func LinearGradient(stops ...GradientStop) Gradient {
	return style.LinearGradient(stops...)
}

// ColorStop creates a gradient stop at a normalized offset.
func ColorStop(offset float32, value ColorSource) GradientStop {
	return style.ColorStop(offset, value)
}
