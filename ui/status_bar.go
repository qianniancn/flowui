package ui

import "github.com/qianniancn/flowui/internal/components/statusbar"

type StatusBarWidget = statusbar.Widget

func StatusBar(left, right Widget) StatusBarWidget {
	return statusbar.New(left, right)
}
