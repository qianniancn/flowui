package theme

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
)

// AvatarRequest describes the semantic axes and caller-owned declaration needed
// by an avatar visual provider. The widget owns image/fallback content and
// accessible semantics; the provider owns geometry, typography and colors.
type AvatarRequest struct {
	Color    AvatarColor
	Variant  AvatarVariant
	Size     AvatarSize
	Override *AvatarOverride
}

// AvatarOverride carries caller customization across the neutral provider
// boundary. It is deliberately provider-owned rather than stored on Widget.
type AvatarOverride struct {
	Declaration style.Style
}

// AvatarOverrideStyle extracts the provider-owned instance layer.
func AvatarOverrideStyle(request AvatarRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// AvatarIconRenderer paints the provider-owned default avatar glyph. The
// widget supplies only the resolved icon slot constraints and foreground
// color; the provider chooses the glyph and drawing strategy.
type AvatarIconRenderer func(layout.Context, color.NRGBA) layout.Dimensions

// AvatarVisual is the complete visual plan consumed by avatar behavior.
type AvatarVisual struct {
	Defaults     style.Style
	Variant      style.Style
	Size         style.Style
	Override     style.Style
	IconRenderer AvatarIconRenderer
}

// AvatarProvider supplies avatar declarations without taking ownership of
// image, fallback content or accessibility behavior.
type AvatarProvider interface {
	Avatar(active *Theme, request AvatarRequest) AvatarVisual
}

// ValidateAvatarVisual checks the geometry, paint and fallback-part contract
// before a widget stores a provider result in its visual cache.
func ValidateAvatarVisual(visual AvatarVisual) {
	if visual.IconRenderer == nil {
		invalidVisual("Avatar", "IconRenderer", "must be provided")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if root.Box == nil {
		invalidVisual("Avatar", "Root.Box", "must be provided")
	}
	if !hasPositiveHeightOrAspect(root.Box) {
		invalidVisual("Avatar", "Root.Box", "must have positive height or aspect ratio")
	}
	if root.Box.Width == nil {
		invalidVisual("Avatar", "Root.Box.Width", "must be provided")
	}
	if !validPositiveDp(*root.Box.Width) {
		invalidVisual("Avatar", "Root.Box.Width", "must be finite and greater than 0")
	}
	if !validOptionalPadding(root.Box.Padding) {
		invalidVisual("Avatar", "Root.Box.Padding", "padding sides must be finite and non-negative")
	}
	if root.Paint == nil {
		invalidVisual("Avatar", "Root.Paint", "must be provided")
	}
	if root.Paint.Radius == nil {
		invalidVisual("Avatar", "Root.Paint.Radius", "must be provided")
	}
	if !validNonNegativeDp(*root.Paint.Radius) {
		invalidVisual("Avatar", "Root.Paint.Radius", "must be finite and non-negative")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("Avatar", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}

	label := style.CascadePart(style.State{}, style.PartLabel, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if label.Text == nil {
		invalidVisual("Avatar", "Label.Text", "must be provided")
	}
	if label.Text.FontSize == nil {
		invalidVisual("Avatar", "Label.Text.FontSize", "must be provided")
	}
	if !validPositiveSp(*label.Text.FontSize) {
		invalidVisual("Avatar", "Label.Text.FontSize", "must be finite and greater than 0")
	}
	if !hasColor(label.Text.Color) {
		invalidVisual("Avatar", "Label.Text.Color", "must resolve to a solid color")
	}

	icon := style.CascadePart(style.State{}, style.PartIcon, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if icon.Box == nil {
		invalidVisual("Avatar", "Icon.Box", "must be provided")
	}
	if icon.Box.Width == nil {
		invalidVisual("Avatar", "Icon.Box.Width", "must be provided")
	}
	if icon.Box.Height == nil {
		invalidVisual("Avatar", "Icon.Box.Height", "must be provided")
	}
	if !validPositiveDp(*icon.Box.Width) {
		invalidVisual("Avatar", "Icon.Box.Width", "must be finite and greater than 0")
	}
	if !validPositiveDp(*icon.Box.Height) {
		invalidVisual("Avatar", "Icon.Box.Height", "must be finite and greater than 0")
	}
}
