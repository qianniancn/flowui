package ui

import (
	"gioui.org/op/paint"
	imageview "github.com/qianniancn/flowui/internal/components/image"
)

type ImageWidget = imageview.Widget

// ImageFit controls how an image is fitted inside its bounds.
type ImageFit = imageview.Fit

const (
	ImageScaleDown = imageview.FitScaleDown
	ImageContain   = imageview.FitContain
	ImageCover     = imageview.FitCover
	ImageFill      = imageview.FitFill
	ImageUnscaled  = imageview.FitUnscaled
)

// Image creates an image widget from a Gio image operation.
func Image(source paint.ImageOp) ImageWidget {
	return imageview.New(source)
}
