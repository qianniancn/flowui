package portal

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

// Layer selects the Portal's root stacking group.
type Layer uint8

const (
	LayerPopup Layer = iota
	LayerModal
)

// Content creates non-nil root-level content from the resolved viewport anchor.
type Content func(anchor image.Rectangle, interactive bool) frame.Widget

// Widget registers arbitrary content with the root Overlay Host.
type Widget struct {
	key      string
	visible  bool
	anchor   frame.Widget
	content  Content
	layer    Layer
	passive  bool
	disabled bool
}

func New(key string, visible bool, anchor frame.Widget, content Content) Widget {
	return Widget{key: key, visible: visible, anchor: anchor, content: content}
}

// Layer sets the root stacking group. It controls ordering, not modal behavior.
func (p Widget) Layer(layer Layer) Widget {
	p.layer = layer
	return p
}

// Passive excludes the Portal from root input ownership.
func (p Widget) Passive(passive bool) Widget {
	p.passive = passive
	return p
}

// Disabled lays out Portal content with disabled input.
func (p Widget) Disabled(disabled bool) Widget {
	p.disabled = disabled
	return p
}

func (p Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, state.KindPortal, p.key)
	anchorDims := layout.Dimensions{}
	if p.anchor != nil {
		anchorDims = p.anchor.Layout(ctx, gtx)
	}
	if !p.visible || p.content == nil {
		return anchorDims
	}

	request := frame.OverlayRequest{
		Key:      key,
		Layer:    p.overlayLayer(),
		Disabled: p.disabled || frame.OverlayNaturallyDisabled(gtx),
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
	frame.RegisterOverlay(ctx, request)
	return anchorDims
}

func (p Widget) overlayLayer() frame.OverlayLayer {
	if p.layer == LayerModal {
		return frame.OverlayLayerModal
	}
	return frame.OverlayLayerPopup
}
