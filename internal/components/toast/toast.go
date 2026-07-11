package toast

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ToastVariant uint8

const (
	ToastDefault ToastVariant = iota
	ToastAccent
	ToastSuccess
	ToastWarning
	ToastDanger
)

type ToastPlacement uint8

const (
	ToastBottom ToastPlacement = iota
	ToastBottomStart
	ToastBottomEnd
	ToastTop
	ToastTopStart
	ToastTopEnd
)

type ToastItem struct {
	key           string
	title         string
	description   string
	variant       ToastVariant
	indicator     frame.Widget
	hasIndicator  bool
	loading       bool
	actionLabel   string
	actionVariant button.ButtonVariant
	timeout       time.Duration
	hasTimeout    bool
}

func Toast(key, title string) ToastItem {
	return ToastItem{key: key, title: title, actionVariant: button.ButtonPrimary}
}

func (t ToastItem) Key() string {
	return t.key
}

func (t ToastItem) Description(description string) ToastItem {
	t.description = description
	return t
}

func (t ToastItem) Variant(variant ToastVariant) ToastItem {
	t.variant = variant
	return t
}

// Indicator replaces the default variant icon. Passing nil hides the indicator.
func (t ToastItem) Indicator(indicator frame.Widget) ToastItem {
	t.indicator = indicator
	t.hasIndicator = true
	return t
}

func (t ToastItem) Loading(loading bool) ToastItem {
	t.loading = loading
	return t
}

func (t ToastItem) Action(label string) ToastItem {
	t.actionLabel = label
	return t
}

func (t ToastItem) ActionVariant(variant button.ButtonVariant) ToastItem {
	t.actionVariant = variant
	return t
}

// Timeout sets the auto-close delay. A zero duration keeps the toast open.
func (t ToastItem) Timeout(timeout time.Duration) ToastItem {
	t.timeout = max(timeout, 0)
	t.hasTimeout = true
	return t
}

func (t ToastItem) showIndicator() bool {
	return !t.hasIndicator || t.indicator != nil
}

type ToastProviderWidget struct {
	key           string
	items         []ToastItem
	onClose       func(string)
	onAction      func(string)
	placement     ToastPlacement
	gap           unit.Dp
	hasGap        bool
	maxVisible    int
	hasMaxVisible bool
	scaleFactor   float32
	hasScale      bool
	width         unit.Dp
	hasWidth      bool
	paused        bool
}

func ToastProvider(key string, items []ToastItem) ToastProviderWidget {
	return ToastProviderWidget{key: key, items: items}
}

func (p ToastProviderWidget) OnClose(onClose func(string)) ToastProviderWidget {
	p.onClose = onClose
	return p
}

func (p ToastProviderWidget) OnAction(onAction func(string)) ToastProviderWidget {
	p.onAction = onAction
	return p
}

func (p ToastProviderWidget) Placement(placement ToastPlacement) ToastProviderWidget {
	p.placement = placement
	return p
}

func (p ToastProviderWidget) Gap(dp int) ToastProviderWidget {
	p.gap = unit.Dp(max(dp, 0))
	p.hasGap = true
	return p
}

func (p ToastProviderWidget) MaxVisible(maxVisible int) ToastProviderWidget {
	p.maxVisible = max(maxVisible, 1)
	p.hasMaxVisible = true
	return p
}

func (p ToastProviderWidget) ScaleFactor(scale float32) ToastProviderWidget {
	p.scaleFactor = min(max(scale, 0), 1)
	p.hasScale = true
	return p
}

func (p ToastProviderWidget) Width(dp int) ToastProviderWidget {
	p.width = unit.Dp(max(dp, 0))
	p.hasWidth = true
	return p
}

func (p ToastProviderWidget) Paused(paused bool) ToastProviderWidget {
	p.paused = paused
	return p
}

func (p ToastProviderWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	fullKey, providerState := toastStateFor(ctx, p.key)
	providerState.sync(gtx, p.items, p.defaultTimeout(ctx))
	providerState.cleanup()
	if !providerState.visible() {
		return layout.Dimensions{}
	}

	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:      fullKey,
		Layer:    frame.OverlayLayerPopup,
		Disabled: frame.OverlayNaturallyDisabled(gtx),
		Passive:  true,
		Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			return p.layoutOverlay(ctx, gtx, providerState)
		},
	})
	frame.AfterOverlays(ctx, providerState.cleanup)
	return layout.Dimensions{}
}

func (p ToastProviderWidget) defaultTimeout(ctx *frame.Context) time.Duration {
	return max(frame.ActiveTheme(ctx).Components.Toast.DefaultTimeout, 0)
}
