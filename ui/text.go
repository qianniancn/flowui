package ui

import "github.com/qianniancn/FlowUI/internal/components/text"

type TextWidget = text.Widget

func Text(value string) TextWidget {
	return text.New(value)
}
