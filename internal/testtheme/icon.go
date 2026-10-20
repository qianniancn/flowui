package testtheme

import (
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/iconstyle"
)

func (behaviorProvider) Icon(active *theme.Theme, request theme.IconRequest) theme.IconVisual {
	return iconstyle.BuildIcon(active, request)
}
