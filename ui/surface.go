package ui

import "github.com/qianniancn/flowui/internal/components/surface"

type SurfaceVariant = surface.SurfaceVariant
type SurfaceWidget = surface.SurfaceWidget

const (
	SurfaceDefault     = surface.SurfaceDefault
	SurfaceSecondary   = surface.SurfaceSecondary
	SurfaceTertiary    = surface.SurfaceTertiary
	SurfaceTransparent = surface.SurfaceTransparent
)

func Surface(child Widget) SurfaceWidget {
	return surface.Surface(child)
}
