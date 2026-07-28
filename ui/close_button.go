package ui

import "github.com/qianniancn/flowui/internal/components/closebutton"

type CloseButtonWidget = closebutton.CloseButtonWidget

func CloseButton(key string) CloseButtonWidget {
	return closebutton.CloseButton(key)
}
