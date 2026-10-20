package portal

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
)

// PortalLayer selects the portal's root stacking group.
type Layer uint8

const (
	LayerPopup Layer = iota
	LayerModal
)

// Content creates non-nil root-level content from the resolved viewport anchor.
type Content func(anchor image.Rectangle, interactive bool) core.Widget

// Widget registers arbitrary content with the root Overlay Host.
type Widget struct {
	key      string
	visible  bool
	anchor   core.Widget
	content  Content
	layer    Layer
	passive  bool
	disabled bool
}

func New(key string, visible bool, anchor core.Widget, content Content) Widget {
	return Widget{key: key, visible: visible, anchor: anchor, content: content}
}

// Layer sets the root stacking group. It controls ordering, not modal behavior.
func (p Widget) Layer(layer Layer) Widget {
	p.layer = layer
	return p
}

// Passive excludes the portal from root input ownership.
func (p Widget) Passive(passive bool) Widget {
	p.passive = passive
	return p
}

// Disabled lays out portal content with disabled input.
func (p Widget) Disabled(disabled bool) Widget {
	p.disabled = disabled
	return p
}

func (p Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindPortal, p.key)
	anchorDims := layout.Dimensions{}
	if p.anchor != nil {
		anchorDims = p.anchor.Layout(ctx, gtx)
	}
	if !p.visible || p.content == nil {
		return anchorDims
	}

	request := core.OverlayRequest{
		Key:      key,
		Layer:    p.overlayLayer(),
		Disabled: p.disabled || core.OverlayNaturallyDisabled(gtx),
		Passive:  p.passive,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			content := p.content(anchor, interactive)
			if content == nil {
				panic("flowui: portal content returned nil")
			}
			return content.Layout(ctx, gtx)
		},
	}
	if p.anchor != nil {
		request.Anchor = image.Rectangle{Max: anchorDims.Size}
		request.HasAnchor = true
	}
	core.RegisterOverlay(ctx, request)
	return anchorDims
}

func (p Widget) overlayLayer() core.OverlayLayer {
	if p.layer == LayerModal {
		return core.OverlayLayerModal
	}
	return core.OverlayLayerPopup
}
