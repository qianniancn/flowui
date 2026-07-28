package ui

import "github.com/qianniancn/flowui/internal/components/badge"

type BadgeWidget = badge.Widget
type BadgeColor = badge.Color
type BadgeVariant = badge.Variant
type BadgeSize = badge.Size
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

func Badge(anchor Widget, label string) BadgeWidget {
	return badge.New(anchor, label)
}
