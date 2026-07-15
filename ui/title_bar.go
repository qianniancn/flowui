package ui

import "github.com/qianniancn/FlowUI/internal/components/titlebar"

type WindowTitleBarWidget = titlebar.Widget

// WindowTitleBar creates client-side decorations for an undecorated window.
func WindowTitleBar(key, title string, menu Widget) WindowTitleBarWidget {
	return titlebar.New(key, title, menu)
}
