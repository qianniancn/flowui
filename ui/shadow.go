// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
)

type ShadowShapeKind = render.ShadowShapeKind

const (
	ShadowRoundedRect = render.ShadowRoundedRect
	ShadowEllipse     = render.ShadowEllipse
)

type ShadowCornerRadii = render.ShadowCornerRadii
type ShadowShape = render.ShadowShape
type ShadowLayer = render.ShadowLayer
type BoxShadow = render.BoxShadow

func RoundedShadowCorners(nw, ne, se, sw unit.Dp) ShadowShape {
	return render.RoundedShadowCorners(nw, ne, se, sw)
}

func EllipseShadow() ShadowShape {
	return render.EllipseShadow()
}

func PopupShadow(col color.NRGBA) BoxShadow {
	return render.PopupShadow(col)
}

func SurfaceShadow(col color.NRGBA) BoxShadow {
	return render.SurfaceShadow(col)
}

func DrawShadow(gtx layout.Context, bounds image.Rectangle, shape ShadowShape, box BoxShadow) {
	render.DrawShadow(gtx, bounds, shape, box)
}
