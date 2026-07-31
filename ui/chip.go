package ui

import "github.com/qianniancn/flowui/internal/components/chip"

type ChipWidget = chip.Widget

// ChipColor selects the chip's semantic color.
type ChipColor = chip.Color

// ChipVariant selects the chip's visual treatment.
type ChipVariant = chip.Variant

// ChipSize selects the chip's padding and text size.
type ChipSize = chip.Size

const (
	ChipDefault = chip.ColorDefault
	ChipAccent  = chip.ColorAccent
	ChipSuccess = chip.ColorSuccess
	ChipWarning = chip.ColorWarning
	ChipDanger  = chip.ColorDanger

	ChipSecondary = chip.VariantSecondary
	ChipPrimary   = chip.VariantPrimary
	ChipTertiary  = chip.VariantTertiary
	ChipSoft      = chip.VariantSoft

	ChipMedium = chip.SizeMedium
	ChipSmall  = chip.SizeSmall
	ChipLarge  = chip.SizeLarge
)

// Chip creates a compact labeled chip.
func Chip(label string) ChipWidget {
	return chip.New(label)
}
