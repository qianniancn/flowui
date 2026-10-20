package theme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
)

// TitleBarControlGlyph is the neutral glyph vocabulary for client-side
// window controls. The provider owns how each glyph is drawn.
type TitleBarControlGlyph uint8

const (
	TitleBarGlyphNone TitleBarControlGlyph = iota
	TitleBarGlyphMinimize
	TitleBarGlyphMaximize
	TitleBarGlyphRestore
	TitleBarGlyphClose
)

// TitleBarRequest describes title-bar behavior facts needed by a visual
// provider. Window state, action visibility, and callbacks remain widget-owned.
type TitleBarRequest struct {
	ClientDecorations bool
	Override          *TitleBarOverride
}

// TitleBarOverride carries caller customization across the neutral provider
// boundary. The provider owns how this declaration is merged.
type TitleBarOverride struct {
	Declaration style.Style
}

// TitleBarOverrideStyle extracts the provider-owned instance layer.
func TitleBarOverrideStyle(request TitleBarRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// TitleBarControlRenderRequest supplies neutral layout and interaction facts
// to a provider-owned window-control renderer.
type TitleBarControlRenderRequest struct {
	Size         image.Point
	Glyph        TitleBarControlGlyph
	Foreground   color.NRGBA
	Focus        color.NRGBA
	FocusVisible bool
}

// TitleBarControlRenderer draws one client-side window control.
type TitleBarControlRenderer func(layout.Context, TitleBarControlRenderRequest)

// TitleBarForegroundColorResolver extracts the concrete control foreground
// needed by a provider-owned renderer.
type TitleBarForegroundColorResolver func(style.ResolvedStyle) (color.NRGBA, bool)

// TitleBarVisual is the complete visual plan consumed by title-bar behavior.
// The widget arranges semantic slots and window actions; the provider owns
// every declaration and the control drawing strategy.
type TitleBarVisual struct {
	Defaults          style.Style
	Override          style.Style
	Focus             color.NRGBA
	Control           TitleBarControlRenderer
	ControlStyle      style.Style
	CloseControlStyle style.Style
	ForegroundColor   TitleBarForegroundColorResolver
}

// TitleBarProvider supplies title-bar declarations and control rendering
// without taking ownership of window lifecycle or slot ordering.
type TitleBarProvider interface {
	TitleBar(active *Theme, request TitleBarRequest) TitleBarVisual
}

// ValidateTitleBarVisual checks the provider-owned title-bar declaration and
// window-control renderer before behavior stores a visual plan in its cache.
func ValidateTitleBarVisual(visual TitleBarVisual) {
	if visual.Control == nil {
		invalidVisual("TitleBar", "Control", "must be provided")
	}
	if visual.ForegroundColor == nil {
		invalidVisual("TitleBar", "ForegroundColor", "must be provided")
	}

	for _, declaration := range []style.Style{visual.ControlStyle, visual.CloseControlStyle} {
		for _, state := range []style.State{{}, {Hovered: true}, {Pressed: true}} {
			control := style.CascadePart(state, style.PartItem, visual.Defaults, declaration, visual.Override)
			if control.Box == nil {
				invalidVisual("TitleBar", "Control.Box", "must be provided")
			}
			if control.Box.Width == nil {
				invalidVisual("TitleBar", "Control.Box.Width", "must be provided")
			}
			if *control.Box.Width <= 0 {
				invalidVisual("TitleBar", "Control.Box.Width", "must be greater than 0")
			}

			if _, ok := visual.ForegroundColor(control); !ok {
				invalidVisual("TitleBar", "ForegroundColor", "must resolve the declaration color")
			}
		}
	}
	resolved := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if resolved.Box == nil {
		invalidVisual("TitleBar", "Resolved.Box", "must be provided")
	}
	if resolved.Box.Height == nil {
		invalidVisual("TitleBar", "Resolved.Box.Height", "must be provided")
	}
	if *resolved.Box.Height <= 0 {
		invalidVisual("TitleBar", "Resolved.Box.Height", "must be greater than 0")
	}
	if resolved.Box.Padding == nil {
		invalidVisual("TitleBar", "Resolved.Box.Padding", "must be provided")
	}
	if resolved.Paint == nil {
		invalidVisual("TitleBar", "Resolved.Paint", "must be provided")
	}
	if resolved.Paint.Background == nil {
		invalidVisual("TitleBar", "Resolved.Paint.Background", "must be provided")
	}
	if resolved.Text == nil {
		invalidVisual("TitleBar", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("TitleBar", "Resolved.Text.Color", "must resolve to a solid color")
	}

	for _, part := range []style.Part{style.PartPrefix, style.PartLabel, style.PartContent, style.PartSuffix, style.PartIndicator} {
		partStyle := style.CascadePart(style.State{}, part, visual.Defaults, visual.Override)
		if part == style.PartIndicator {
			if partStyle.Box == nil {
				invalidVisual("TitleBar", "PartStyle.Box", "must be provided")
			}
			if partStyle.Box.Height == nil {
				invalidVisual("TitleBar", "PartStyle.Box.Height", "must be provided")
			}
			if *partStyle.Box.Height <= 0 {
				invalidVisual("TitleBar", "PartStyle.Box.Height", "must be greater than 0")
			}
			if partStyle.Paint == nil {
				invalidVisual("TitleBar", "PartStyle.Paint", "must be provided")
			}
			if !hasPaint(partStyle.Paint.Background) {
				invalidVisual("TitleBar", "PartStyle.Paint.Background", "must provide a solid color or nonempty gradient")
			}

			continue
		}
		if partStyle.Text == nil {
			invalidVisual("TitleBar", "PartStyle.Text", "must be provided")
		}
		if !hasColor(partStyle.Text.Color) {
			invalidVisual("TitleBar", "PartStyle.Text.Color", "must resolve to a solid color")
		}
	}
}
