package button

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

type Widget struct {
	key         string
	child       core.Widget
	label       string
	onClick     func()
	variant     Variant
	variantSet  bool
	size        Size
	sizeSet     bool
	disabled    bool
	disabledSet bool
	loading     bool
	fullWidth   bool
	iconOnly    bool
	group       buttonGroupItemStyle
	override    *theme.ButtonOverride
}

type Variant = theme.ButtonVariant

const (
	Primary    = theme.ButtonPrimary
	Secondary  = theme.ButtonSecondary
	Tertiary   = theme.ButtonTertiary
	Ghost      = theme.ButtonGhost
	Outline    = theme.ButtonOutline
	Danger     = theme.ButtonDanger
	DangerSoft = theme.ButtonDangerSoft
	// DangerOutline is the bordered danger button.
	DangerOutline = theme.ButtonDangerOutline
)

type Size = theme.ButtonSize

const (
	Medium = theme.ButtonMedium
	Small  = theme.ButtonSmall
	Large  = theme.ButtonLarge
)

// New creates a button with visible text and a matching accessible label.
// Use Content to replace the text with an icon or composite widget.
// New creates a button. Its identity is derived from the enclosing scope; call
// Key when the same position may hold different buttons.
func New(label string) Widget {
	return Widget{key: core.AutoKey(), child: text.New(label), label: label}
}

// Key sets an explicit identity for the button state.
func (b Widget) Key(key string) Widget {
	b.key = key
	return b
}

// Content replaces the visible content, preserving the accessible label.
func (b Widget) Content(value core.Widget) Widget {
	b.child = value
	return b
}

// Styled applies an instance visual declaration at the provider boundary.
//
// New code should prefer this wrapper when a style is supplied separately from
// button behavior. The declaration lives on the wrapper, while the wrapped
// Widget continues to contain only behavior and semantic configuration.
// Style merges instance declarations into the button. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (b Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if b.override == nil {
		b.override = &theme.ButtonOverride{}
	}
	b.override.Declaration = style.Join(b.override.Declaration, declaration)
	return b
}

func (b Widget) OnClick(fn func()) Widget {
	b.onClick = fn
	return b
}

// Label overrides the accessible label without changing the visible content.
func (b Widget) Label(value string) Widget {
	b.label = value
	return b
}

func (b Widget) Disabled(disabled bool) Widget {
	b.disabled = disabled
	b.disabledSet = true
	return b
}

func (b Widget) Loading(loading bool) Widget {
	b.loading = loading
	return b
}

func (b Widget) Variant(variant Variant) Widget {
	b.variant = variant
	b.variantSet = true
	return b
}

func (b Widget) Size(size Size) Widget {
	b.size = size
	b.sizeSet = true
	return b
}

func (b Widget) FullWidth() Widget {
	b.fullWidth = true
	return b
}

func (b Widget) IconOnly() Widget {
	b.iconOnly = true
	return b
}

// DisabledState reports the resolved instance disabled flag to composite
// widgets inside the internal component tree.
func (b Widget) DisabledState() bool {
	return b.disabled
}

func (b Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, nil, true, b.override)
}

// LayoutWithClickable renders a button with caller-owned interaction state.
// It is intended for internal component composition.
func LayoutWithClickable(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, true, b.override)
}

// LayoutWithClickableNoEvents renders a button with caller-owned event handling.
func LayoutWithClickableNoEvents(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, false, b.override)
}

func layoutWithClickable(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable, handleEvents bool, override *theme.ButtonOverride) layout.Dimensions {
	dims, _ := layoutWithClickableResolved(b, ctx, gtx, clickable, handleEvents, override, nil)
	return dims
}

func layoutWithClickableResolved(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable, handleEvents bool, override *theme.ButtonOverride, prepared *buttonPreparedContent) (layout.Dimensions, buttonStyle) {
	core.RequireVisualProvider(ctx)
	animGtx := gtx
	click := core.Begin(ctx, gtx, b.key, clickable, !b.disabled && !b.loading, handleEvents, b.onClick)
	styleState := click.State
	styleState.Disabled = b.disabled || !gtx.Enabled()
	styleState.Loading = b.loading
	var resolvedStyle buttonStyle
	if prepared != nil {
		resolvedStyle = b.resolvePreparedStyleWithDeclarations(ctx, animGtx, click.Key, styleState, prepared.declarations)
	} else {
		resolvedStyle = b.resolveStyle(ctx, animGtx, click.Key, styleState, override)
	}

	if b.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if b.loading && !b.iconOnly && !b.fullWidth {
		resolvedStyle = buttonLoadingStyle(gtx, resolvedStyle)
	}

	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		motion := resolvedStyle.visual.Motion
		if b.loading && theme.MotionEnabled(motion) {
			animGtx.Execute(op.InvalidateCmd{})
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if prepared != nil {
				prepared.call.Add(gtx.Ops)
				return prepared.dims
			}
			return b.layoutContent(ctx, gtx, resolvedStyle)
		})
	})

	// ripple: keyed state absorbs presses, the overlay paints
	// them on top of the button box while one is animating. The wave anchors
	// to the button box (the overlay's dims), not the inner content; the
	// box-shadow spread hugs the border, padding included.
	ripples, spreadDuration, fadeDuration := b.rippleOverlay(ctx, click, resolvedStyle.visual, gtx)
	var waveOverlay func(layout.Context, layout.Dimensions)
	if ripples != nil {
		waveCol, waveOK := buttonRippleColor(resolvedStyle.visual, resolvedStyle.root)
		waveOverlay = func(gtx layout.Context, dims layout.Dimensions) {
			resolvedStyle.visual.RippleRenderer(gtx, theme.ButtonRippleRequest{
				Ripples:        ripples.active,
				SpreadDuration: spreadDuration,
				FadeDuration:   fadeDuration,
				Size:           dims.Size,
				Radius:         buttonRadius(resolvedStyle.root),
				Color:          waveCol,
				ColorValid:     waveOK,
			})
		}
	}

	dims := flowlayout.InteractiveResolvedWithDrawAndOverlay(ctx, gtx, resolvedStyle.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		// Keep the semantic and pointer region bounded by the button size while
		// letting paint that overflows it -- the focus ring, the drop shadow --
		// escape the hit target's clip.
		return click.LayoutVisualOverflow(gtx, visual, b.label)
	}, nil, waveOverlay)
	return dims, resolvedStyle
}

// rippleOverlay returns the keyed ripple state while ButtonTheme.Ripple is on
// and the button is interactive; nil disables the overlay entirely.
func (b Widget) rippleOverlay(ctx *core.Context, click core.Click, visual theme.ButtonVisual, gtx layout.Context) (*rippleState, time.Duration, time.Duration) {
	if !visual.Ripple || b.disabled || b.loading || !gtx.Enabled() {
		return nil, 0, 0
	}
	state := core.UseStateWith(ctx, click.Key, stateSlotRipple, func() *rippleState { return &rippleState{} })
	spread := rippleDurationFor(visual)
	fade := rippleFadeDurationFor(visual)
	updateRipples(state, click.Clickable.History(), fade, gtx.Now)
	return state, spread, fade
}
