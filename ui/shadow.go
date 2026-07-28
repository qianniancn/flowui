// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/render"
)

type ShadowShapeKind = render.ShadowShapeKind

const (
	ShadowRoundedRect = render.ShadowRoundedRect
	ShadowEllipse     = render.ShadowEllipse
)

type ShadowCornerRadii = render.ShadowCornerRadii
type ShadowShape = render.ShadowShape
type ShadowLayer = render.ShadowLayer
type RenderBoxShadow = render.BoxShadow

func RoundedShadowCorners(nw, ne, se, sw unit.Dp) ShadowShape {
	return render.RoundedShadowCorners(nw, ne, se, sw)
}

func EllipseShadow() ShadowShape {
	return render.EllipseShadow()
}

func PopupShadow(col color.NRGBA) RenderBoxShadow {
	return render.ThemeShadow(DefaultShadows().Overlay, col, 1)
}

func SurfaceShadow(col color.NRGBA) RenderBoxShadow {
	return render.ThemeShadow(DefaultShadows().Surface, col, 1)
}

func ThemeShadow(style ShadowTheme, col color.NRGBA, opacity float32) RenderBoxShadow {
	return render.ThemeShadow(style, col, opacity)
}

func DrawShadow(gtx layout.Context, bounds image.Rectangle, shape ShadowShape, box RenderBoxShadow) {
	render.DrawShadow(gtx, bounds, shape, box)
}
