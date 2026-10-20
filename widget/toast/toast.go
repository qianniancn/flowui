package toast

import (
	"github.com/qianniancn/flowui/style"
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// Variant identifies a toast's semantic visual treatment.
type Variant = theme.ToastVariant

const (
	Default = theme.ToastDefault
	Accent  = theme.ToastAccent
	Success = theme.ToastSuccess
	Warning = theme.ToastWarning
	Danger  = theme.ToastDanger
)

type Placement uint8

const (
	Bottom Placement = iota
	BottomStart
	BottomEnd
	Top
	TopStart
	TopEnd
)

type Item struct {
	key           string
	title         string
	description   string
	variant       Variant
	indicator     core.Widget
	hasIndicator  bool
	loading       bool
	compact       bool
	actionLabel   string
	actionVariant theme.ButtonVariant
	timeout       time.Duration
	hasTimeout    bool
}

func NewItem(key, title string) Item {
	return Item{key: key, title: title, actionVariant: theme.ButtonGhost}
}

// MessageItem creates a compact, transient message: a single-line title,
// auto-width, and no description, action, or close button. It is the
// lightweight top-center feedback pattern; place its host with Placement(Top).
func MessageItem(key, title string) Item {
	return NewItem(key, title).Compact(true).Timeout(3 * time.Second)
}

func (t Item) Key() string {
	return t.key
}

func (t Item) Description(description string) Item {
	t.description = description
	return t
}

func (t Item) Variant(variant Variant) Item {
	t.variant = variant
	return t
}

// Indicator replaces the default variant icon. Passing nil hides the indicator.
func (t Item) Indicator(indicator core.Widget) Item {
	t.indicator = indicator
	t.hasIndicator = true
	return t
}

func (t Item) Loading(loading bool) Item {
	t.loading = loading
	return t
}

// Compact renders only the indicator and title, with content-driven width. It
// omits the description, action, and close button.
func (t Item) Compact(compact bool) Item {
	t.compact = compact
	return t
}

func (t Item) Action(label string) Item {
	t.actionLabel = label
	return t
}

func (t Item) ActionVariant(variant theme.ButtonVariant) Item {
	t.actionVariant = variant
	return t
}

// Timeout sets the auto-close delay. A zero duration keeps the toast open.
func (t Item) Timeout(timeout time.Duration) Item {
	t.timeout = max(timeout, 0)
	t.hasTimeout = true
	return t
}

func (t Item) showIndicator() bool {
	return !t.hasIndicator || t.indicator != nil
}

type Widget struct {
	override      *theme.ToastOverride
	key           string
	items         []Item
	onClose       func(string)
	onAction      func(string)
	placement     Placement
	offset        unit.Dp
	hasOffset     bool
	gap           unit.Dp
	hasGap        bool
	maxVisible    int
	hasMaxVisible bool
	scaleFactor   float32
	hasScale      bool
	width         unit.Dp
	hasWidth      bool
	paused        bool
	visualCache   *toastVisualCache
}

func New(key string, items []Item) Widget {
	return Widget{key: key, items: items, visualCache: &toastVisualCache{}}
}

func (p Widget) OnClose(onClose func(string)) Widget {
	p.onClose = onClose
	return p
}

func (p Widget) OnAction(onAction func(string)) Widget {
	p.onAction = onAction
	return p
}

func (p Widget) Placement(placement Placement) Widget {
	p.placement = placement
	return p
}

// Offset sets the distance from the top or bottom viewport edge.
func (p Widget) Offset(dp int) Widget {
	p.offset = unit.Dp(max(dp, 0))
	p.hasOffset = true
	return p
}

func (p Widget) Gap(dp int) Widget {
	p.gap = unit.Dp(max(dp, 0))
	p.hasGap = true
	return p
}

func (p Widget) MaxVisible(maxVisible int) Widget {
	p.maxVisible = max(maxVisible, 1)
	p.hasMaxVisible = true
	return p
}

func (p Widget) ScaleFactor(scale float32) Widget {
	p.scaleFactor = min(max(scale, 0), 1)
	p.hasScale = true
	return p
}

func (p Widget) Width(dp int) Widget {
	p.width = unit.Dp(max(dp, 0))
	p.hasWidth = true
	return p
}

func (p Widget) Paused(paused bool) Widget {
	p.paused = paused
	return p
}

func (p Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithOverride(ctx, gtx, p.override)
}

func (p Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ToastOverride) layout.Dimensions {
	fullKey, providerState := toastStateFor(ctx, p.key)
	baseVisual := p.baseVisual(ctx, override)
	providerState.sync(gtx, p.items, baseVisual.Metrics.DefaultTimeout)
	providerState.cleanup()
	if !providerState.visible() {
		providerState.resetRegion()
		return layout.Dimensions{}
	}

	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:      fullKey,
		Layer:    core.OverlayLayerPopup,
		Disabled: core.OverlayNaturallyDisabled(gtx),
		Passive:  true,
		Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			return p.layoutOverlay(ctx, gtx, providerState, baseVisual, override)
		},
	})
	core.AfterOverlays(ctx, providerState.cleanup)
	return layout.Dimensions{}
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ToastOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
