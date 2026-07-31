package ui

import "github.com/qianniancn/flowui/internal/components/surface"

// SurfaceVariant selects the surface treatment.
type SurfaceVariant = surface.SurfaceVariant

type SurfaceWidget = surface.SurfaceWidget

const (
	SurfaceDefault     = surface.SurfaceDefault
	SurfaceSecondary   = surface.SurfaceSecondary
	SurfaceTertiary    = surface.SurfaceTertiary
	SurfaceTransparent = surface.SurfaceTransparent
)

// Surface creates a themed surface around child.
func Surface(child Widget) SurfaceWidget {
	return surface.Surface(child)
}
