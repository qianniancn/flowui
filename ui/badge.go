package ui

import "github.com/qianniancn/flowui/internal/components/badge"

type BadgeWidget = badge.Widget

// BadgeColor selects the badge's semantic color.
type BadgeColor = badge.Color

// BadgeVariant selects the badge's visual treatment.
type BadgeVariant = badge.Variant

// BadgeSize selects the badge's padding and text size.
type BadgeSize = badge.Size

// BadgePlacement selects the badge's position relative to its anchor.
type BadgePlacement = badge.Placement

const (
	BadgeDefault = badge.ColorDefault
	BadgeAccent  = badge.ColorAccent
	BadgeSuccess = badge.ColorSuccess
	BadgeWarning = badge.ColorWarning
	BadgeDanger  = badge.ColorDanger

	BadgePrimary   = badge.VariantPrimary
	BadgeSecondary = badge.VariantSecondary
	BadgeSoft      = badge.VariantSoft

	BadgeMedium = badge.SizeMedium
	BadgeSmall  = badge.SizeSmall
	BadgeLarge  = badge.SizeLarge

	BadgeTopRight    = badge.PlacementTopRight
	BadgeTopLeft     = badge.PlacementTopLeft
	BadgeBottomRight = badge.PlacementBottomRight
	BadgeBottomLeft  = badge.PlacementBottomLeft
)

// Badge attaches a labeled badge to anchor.
func Badge(anchor Widget, label string) BadgeWidget {
	return badge.New(anchor, label)
}
