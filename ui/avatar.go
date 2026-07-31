package ui

import "github.com/qianniancn/flowui/internal/components/avatar"

type AvatarWidget = avatar.Widget

// AvatarColor selects the avatar's semantic color.
type AvatarColor = avatar.Color

// AvatarVariant selects the avatar's visual treatment.
type AvatarVariant = avatar.Variant

// AvatarSize selects the avatar's diameter.
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

// Avatar creates an avatar that displays fallbackText when no image is supplied.
func Avatar(fallbackText string) AvatarWidget {
	return avatar.New(fallbackText)
}
