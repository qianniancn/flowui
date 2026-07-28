package ui

import "github.com/qianniancn/flowui/internal/components/description"

type DescriptionWidget = description.DescriptionWidget

func Description(text string) DescriptionWidget {
	return description.Description(text)
}
