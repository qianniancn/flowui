package ui

import "github.com/qianniancn/FlowUI/internal/components/avatar"

type AvatarWidget = avatar.Widget
type AvatarColor = avatar.Color
type AvatarVariant = avatar.Variant
type AvatarSize = avatar.Size

const (
	AvatarDefault = avatar.ColorDefault
	AvatarAccent  = avatar.ColorAccent
	AvatarSuccess = avatar.ColorSuccess
	AvatarWarning = avatar.ColorWarning
	AvatarDanger  = avatar.ColorDanger

	AvatarVariantDefault = avatar.VariantDefault
	AvatarSoft           = avatar.VariantSoft

	AvatarMedium = avatar.SizeMedium
	AvatarSmall  = avatar.SizeSmall
	AvatarLarge  = avatar.SizeLarge
)

func Avatar(fallbackText string) AvatarWidget {
	return avatar.New(fallbackText)
}
