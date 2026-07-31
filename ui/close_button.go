package ui

import "github.com/qianniancn/flowui/internal/components/closebutton"

type CloseButtonWidget = closebutton.CloseButtonWidget

// CloseButton creates a keyed button that requests the current window to close.
func CloseButton(key string) CloseButtonWidget {
	return closebutton.CloseButton(key)
}
