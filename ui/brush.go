package ui

import "github.com/qianniancn/FlowUI/internal/style"

func LinearGradient(stops ...GradientStop) Gradient {
	return style.LinearGradient(stops...)
}

func ColorStop(offset float32, value ColorSource) GradientStop {
	return style.ColorStop(offset, value)
}
