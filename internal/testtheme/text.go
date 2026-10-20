package testtheme

import (
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/textstyle"
)

func (behaviorProvider) Text(active *theme.Theme, request theme.TextRequest) theme.TextVisual {
	return textstyle.BuildText(active, request)
}
