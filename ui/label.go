package ui

import "github.com/qianniancn/flowui/internal/components/label"

type LabelWidget = label.LabelWidget

func Label(text string) LabelWidget {
	return label.Label(text)
}
