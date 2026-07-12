package ui

import "github.com/qianniancn/FlowUI/internal/components/chip"

type ChipWidget = chip.Widget
type ChipColor = chip.Color
type ChipVariant = chip.Variant
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

func Chip(label string) ChipWidget {
	return chip.New(label)
}
