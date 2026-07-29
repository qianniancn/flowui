package ui

import "github.com/qianniancn/flowui/internal/components/titlebar"

type WindowTitleBarWidget = titlebar.Widget

// WindowTitleBar creates Windows client-side decorations for an undecorated
// window. On macOS and Linux it becomes an in-content application header that
// keeps Leading, Menu, Center, and Trailing content but omits window controls
// and draggable regions. Use WindowTitleBarSupported before disabling native
// decorations. Instance styles use PartPrefix for Leading, PartLabel for the
// default title, PartContent for Center, PartSuffix for Trailing, PartItem for
// system controls, and PartIndicator for the separator.
func WindowTitleBar(key, title string, menu Widget) WindowTitleBarWidget {
	return titlebar.NewPlatform(key, title, menu)
}

// WindowTitleBarSupported reports whether WindowTitleBar replaces native
// decorations on the current platform. When it returns false, keep native
// decorations enabled; WindowTitleBar remains usable as an application header.
func WindowTitleBarSupported() bool {
	return titlebar.Supported()
}
