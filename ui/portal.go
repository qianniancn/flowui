package ui

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/components/portal"
	"github.com/qianniancn/flowui/internal/frame"
)

type PortalWidget = portal.Widget
type PortalLayer = portal.Layer
type OverlayPlacement = frame.OverlayPlacement

// PortalContent creates root-level content from its resolved viewport anchor.
// Interactive reports whether this Portal owned root input in the preceding frame.
// Returning nil is invalid; set visible to false when no content should be registered.
type PortalContent func(anchor image.Rectangle, interactive bool) Widget

const (
	PortalLayerPopup = portal.LayerPopup
	PortalLayerModal = portal.LayerModal
)

// Portal registers arbitrary content with the root Overlay Host.
// Prefer Popover or Modal when their positioning and interaction policies fit.
func Portal(key string, visible bool, anchor Widget, content PortalContent) PortalWidget {
	var internal portal.Content
	if content != nil {
		internal = func(anchor image.Rectangle, interactive bool) frame.Widget {
			return content(anchor, interactive)
		}
	}
	return portal.New(key, visible, anchor, internal)
}

// TrackOverlayPlacement tracks transforms applied by a custom Gio layout.
// Call PlaceOffset or PlaceTransform after child returns, including for an
// identity placement; unplaced children do not resolve their overlays.
// The returned placement is valid only for the current frame.
func TrackOverlayPlacement(ctx *Context, child func() layout.Dimensions) (layout.Dimensions, OverlayPlacement) {
	return frame.TrackOverlayPlacement(ctx, child)
}
