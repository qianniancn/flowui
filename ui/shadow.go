// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/render"
)

// ShadowShapeKind identifies the geometry used by a shadow.
type ShadowShapeKind = render.ShadowShapeKind

const (
	ShadowRoundedRect = render.ShadowRoundedRect
	ShadowEllipse     = render.ShadowEllipse
)

// ShadowCornerRadii stores the four corner radii of a shadow shape.
type ShadowCornerRadii = render.ShadowCornerRadii

// ShadowShape describes the geometry onto which a shadow is projected.
type ShadowShape = render.ShadowShape

// ShadowLayer describes one rendered shadow pass.
type ShadowLayer = render.ShadowLayer

// RenderBoxShadow is the resolved box-shadow configuration used by DrawShadow.
type RenderBoxShadow = render.BoxShadow

// RoundedShadowCorners creates a rounded-rectangle shadow shape.
func RoundedShadowCorners(nw, ne, se, sw unit.Dp) ShadowShape {
	return render.RoundedShadowCorners(nw, ne, se, sw)
}

// EllipseShadow creates an elliptical shadow shape.
func EllipseShadow() ShadowShape {
	return render.EllipseShadow()
}

// PopupShadow returns the default overlay shadow profile with col as its color.
func PopupShadow(col color.NRGBA) RenderBoxShadow {
	return render.ThemeShadow(DefaultShadows().Overlay, col, 1)
}

// SurfaceShadow returns the default surface shadow profile with col as its color.
func SurfaceShadow(col color.NRGBA) RenderBoxShadow {
	return render.ThemeShadow(DefaultShadows().Surface, col, 1)
}

// ThemeShadow converts a theme shadow profile into a renderable shadow.
func ThemeShadow(style ShadowTheme, col color.NRGBA, opacity float32) RenderBoxShadow {
	return render.ThemeShadow(style, col, opacity)
}

// DrawShadow paints box using shape within bounds.
func DrawShadow(gtx layout.Context, bounds image.Rectangle, shape ShadowShape, box RenderBoxShadow) {
	render.DrawShadow(gtx, bounds, shape, box)
}
