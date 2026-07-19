package ui

import "github.com/qianniancn/FlowUI/internal/components/statusbar"

type StatusBarWidget = statusbar.Widget

func StatusBar(left, right Widget) StatusBarWidget {
	return statusbar.New(left, right)
}
