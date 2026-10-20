package theme

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
)

// CloseButtonRequest carries the optional instance declaration supplied by a
// close-button behavior. Disabled, focus and click state remain widget-owned.
type CloseButtonRequest struct {
	Override *CloseButtonOverride
}

// CloseButtonOverride carries caller customization across the neutral
// provider boundary. The behavior widget does not retain this declaration.
type CloseButtonOverride struct {
	Declaration style.Style
}

// CloseButtonOverrideStyle extracts the provider-owned instance layer.
func CloseButtonOverrideStyle(request CloseButtonRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// CloseButtonIconRenderer paints the provider-owned default close glyph.
// The widget supplies only the resolved foreground color and the icon slot's
// constraints; the theme chooses the actual glyph and drawing strategy.
type CloseButtonIconRenderer func(layout.Context, color.NRGBA) layout.Dimensions

// CloseButtonVisual is the complete visual plan consumed by close-button
// behavior. The provider owns root and icon declarations; the widget only
// resolves them and supplies interaction state.
type CloseButtonVisual struct {
	Defaults     style.Style
	Icon         style.Style
	Override     style.Style
	Motion       MotionTheme
	IconRenderer CloseButtonIconRenderer
}

// CloseButtonProvider supplies close-button declarations and geometry without
// taking ownership of click, focus, disabled, or accessibility behavior.
type CloseButtonProvider interface {
	CloseButton(active *Theme, request CloseButtonRequest) CloseButtonVisual
}

// ValidateCloseButtonVisual checks the root and icon declaration contract
// before a widget stores a provider result in its visual cache.
func ValidateCloseButtonVisual(visual CloseButtonVisual) {
	if visual.IconRenderer == nil {
		invalidVisual("CloseButton", "IconRenderer", "must be provided")
	}

	for _, state := range []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{FocusVisible: true},
		{Disabled: true},
	} {
		root := style.Cascade(state, visual.Defaults, visual.Override)
		if root.Box == nil {
			invalidVisual("CloseButton", "Root.Box", "must be provided")
		}
		if root.Box.Width == nil {
			invalidVisual("CloseButton", "Root.Box.Width", "must be provided")
		}
		if root.Box.Height == nil {
			invalidVisual("CloseButton", "Root.Box.Height", "must be provided")
		}
		if !validPositiveDp(*root.Box.Width) {
			invalidVisual("CloseButton", "Root.Box.Width", "must be finite and greater than 0")
		}
		if !validPositiveDp(*root.Box.Height) {
			invalidVisual("CloseButton", "Root.Box.Height", "must be finite and greater than 0")
		}
		if !validOptionalPadding(root.Box.Padding) {
			invalidVisual("CloseButton", "Root.Box.Padding", "padding sides must be finite and non-negative")
		}
		if root.Paint == nil {
			invalidVisual("CloseButton", "Root.Paint", "must be provided")
		}
		if !hasPaint(root.Paint.Background) {
			invalidVisual("CloseButton", "Root.Paint.Background", "must provide a solid color or nonempty gradient")
		}
		if root.Paint.Radius == nil {
			invalidVisual("CloseButton", "Root.Paint.Radius", "must be provided")
		}
		if !validNonNegativeDp(*root.Paint.Radius) {
			invalidVisual("CloseButton", "Root.Paint.Radius", "must be finite and non-negative")
		}
		if !hasValidOpacity(root.Paint.Opacity) {
			invalidVisual("CloseButton", "Root.Paint.Opacity", "must be finite and between 0 and 1")
		}
		if root.Text == nil {
			invalidVisual("CloseButton", "Root.Text", "must be provided")
		}
		if !hasColor(root.Text.Color) {
			invalidVisual("CloseButton", "Root.Text.Color", "must resolve to a solid color")
		}

		icon := style.CascadePart(state, style.PartIcon, visual.Icon, visual.Override)
		if icon.Box == nil {
			invalidVisual("CloseButton", "Icon.Box", "must be provided")
		}
		if icon.Box.Width == nil {
			invalidVisual("CloseButton", "Icon.Box.Width", "must be provided")
		}
		if icon.Box.Height == nil {
			invalidVisual("CloseButton", "Icon.Box.Height", "must be provided")
		}
		if !validPositiveDp(*icon.Box.Width) {
			invalidVisual("CloseButton", "Icon.Box.Width", "must be finite and greater than 0")
		}
		if !validPositiveDp(*icon.Box.Height) {
			invalidVisual("CloseButton", "Icon.Box.Height", "must be finite and greater than 0")
		}
	}
}
