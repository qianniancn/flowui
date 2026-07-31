package ui

import "github.com/qianniancn/flowui/internal/components/statusbar"

type StatusBarWidget = statusbar.Widget

// StatusBar creates a bar with left and right aligned content.
func StatusBar(left, right Widget) StatusBarWidget {
	return statusbar.New(left, right)
}
