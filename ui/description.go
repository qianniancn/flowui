package ui

import "github.com/qianniancn/flowui/internal/components/description"

type DescriptionWidget = description.DescriptionWidget

// Description creates secondary explanatory text.
func Description(text string) DescriptionWidget {
	return description.Description(text)
}
