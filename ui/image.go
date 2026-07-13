package ui

import (
	"gioui.org/op/paint"
	imageview "github.com/qianniancn/FlowUI/internal/components/image"
)

type ImageWidget = imageview.Widget
type ImageFit = imageview.Fit

const (
	ImageScaleDown = imageview.FitScaleDown
	ImageContain   = imageview.FitContain
	ImageCover     = imageview.FitCover
	ImageFill      = imageview.FitFill
	ImageUnscaled  = imageview.FitUnscaled
)

func Image(source paint.ImageOp) ImageWidget {
	return imageview.New(source)
}
