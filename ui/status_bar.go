package ui

import "github.com/qianniancn/FlowUI/internal/components/statusbar"

type StatusBarWidget = statusbar.Widget
type StatusBarVariant = statusbar.Variant

const (
	StatusBarDefault     = statusbar.Default
	StatusBarAccent      = statusbar.Accent
	StatusBarTransparent = statusbar.Transparent
)

func StatusBar(left, right Widget) StatusBarWidget {
	return statusbar.New(left, right)
}
