package popover

import (
	"github.com/qianniancn/flowui/style"
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	override                *theme.PopoverOverride
	key                     string
	open                    bool
	hasOpen                 bool
	defaultOpen             bool
	hasDefaultOpen          bool
	trigger                 core.Widget
	content                 core.Widget
	heading                 string
	onOpenChange            func(bool)
	placement               core.PopoverPlacement
	offset                  int
	hasOffset               bool
	shouldFlip              bool
	hasShouldFlip           bool
	avoidOverflow           bool
	hasAvoidOverflow        bool
	arrow                   bool
	hasArrow                bool
	dismissable             bool
	hasDismissable          bool
	keyboardDismissDisabled bool
}

// New creates a popover. It starts uncontrolled and closed; Open sets the
// controlled value and DefaultOpen the uncontrolled starting value.
func New(trigger core.Widget, content core.Widget) Widget {
	return Widget{
		key:     core.AutoKey(),
		trigger: trigger,
		content: content,
	}
}

// Key sets an explicit identity for the popover state.
func (p Widget) Key(key string) Widget {
	p.key = key
	return p
}

// Uncontrolled makes the constructor's open value an internal initial state.
func (p Widget) Uncontrolled() Widget {
	p.defaultOpen = p.open
	p.hasDefaultOpen = true
	p.hasOpen = false
	return p
}

func (p Widget) OnOpenChange(fn func(bool)) Widget {
	p.onOpenChange = fn
	return p
}

func (p Widget) Open(open bool) Widget {
	p.open = open
	p.hasOpen = true
	return p
}

func (p Widget) DefaultOpen(open bool) Widget {
	p.defaultOpen = open
	p.hasDefaultOpen = true
	return p
}

func (p Widget) Heading(heading string) Widget {
	p.heading = heading
	return p
}

func (p Widget) Placement(placement core.PopoverPlacement) Widget {
	p.placement = placement
	return p
}

func (p Widget) Offset(dp int) Widget {
	p.offset = dp
	p.hasOffset = true
	return p
}

func (p Widget) ShouldFlip(shouldFlip bool) Widget {
	p.shouldFlip = shouldFlip
	p.hasShouldFlip = true
	return p
}

func (p Widget) AvoidOverflow(avoidOverflow bool) Widget {
	p.avoidOverflow = avoidOverflow
	p.hasAvoidOverflow = true
	return p
}

func (p Widget) Arrow(show bool) Widget {
	p.arrow = show
	p.hasArrow = true
	return p
}

func (p Widget) Dismissable(dismissable bool) Widget {
	p.dismissable = dismissable
	p.hasDismissable = true
	return p
}

func (p Widget) KeyboardDismissDisabled(disabled bool) Widget {
	p.keyboardDismissDisabled = disabled
	return p
}

func (p Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithOverride(ctx, gtx, p.override)
}

func (p Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.PopoverOverride) layout.Dimensions {
	fullKey := core.FullKey(ctx, p.key)
	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)
	triggerDims := p.layoutTrigger(ctx, gtx)

	// Check if we need to evaluate open state (peek first to avoid claiming state).
	// A default-open uncontrolled binding is seeded once; after it is closed,
	// keep its state record so the default value is not applied again.
	stateValue, stateExists := core.PeekState[popoverState](ctx, fullKey, stateSlotPopover)
	needsState := stateExists && (stateValue.open || stateValue.visible() ||
		(!p.hasOpen && p.hasDefaultOpen && p.defaultOpen))
	if !needsState {
		// For uncontrolled mode, check if we have a default that would open it
		if !p.hasOpen && p.hasDefaultOpen && p.defaultOpen && !stateExists {
			needsState = true
		} else if p.hasOpen && p.open {
			needsState = true
		}
	}

	if !needsState {
		return triggerDims
	}

	// Now we know we need state, claim it
	state := popoverStateFor(ctx, p.key)
	state.bind(p)
	open := state.isOpen(p)

	if !open && !state.visible() {
		if p.hasOpen || !p.hasDefaultOpen || !p.defaultOpen {
			deletePopoverState(ctx, fullKey)
		}
		return triggerDims
	}

	visual := p.visualWithCache(ctx, override, &state.visualCache)
	progress := state.progressFor(gtx, open, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration, visual.Motion)
	if !open && progress <= 0 {
		deletePopoverState(ctx, fullKey)
		return triggerDims
	}

	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       fullKey,
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: triggerDims.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			if open && interactive && gtx.Enabled() {
				p.handleCloseEvents(ctx, gtx, state)
			}
			return p.layoutOverlay(ctx, gtx, state, anchor, progress, open && gtx.Enabled(), visual)
		},
	})

	return triggerDims
}

func (p Widget) layoutTrigger(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if p.trigger == nil {
		return layout.Dimensions{}
	}
	return p.trigger.Layout(ctx, gtx)
}

func (p Widget) handleCloseEvents(ctx *core.Context, gtx layout.Context, popoverStateValue *popoverState) {
	for popoverStateValue.dialog.Clicked(gtx) {
	}
	if popoverStateValue.dialog.TakePressed() {
		core.PreserveFocus(ctx)
	}
	for popoverStateValue.arrow.Clicked(gtx) {
	}
	if popoverStateValue.arrow.TakePressed() {
		core.PreserveFocus(ctx)
	}
	if !p.keyboardDismissDisabled && popoverStateValue.escapePressed(gtx) {
		p.requestClose(popoverStateValue)
	}
	for i := range popoverStateValue.dismiss {
		for popoverStateValue.dismiss[i].Clicked(gtx) {
			if p.isDismissable() {
				p.requestClose(popoverStateValue)
			}
		}
		if popoverStateValue.dismiss[i].TakePressed() {
			core.PreserveFocus(ctx)
		}
	}
}

func (p Widget) requestClose(state *popoverState) {
	state.requestOpen(p, false)
}

func (p Widget) isDismissable() bool {
	if !p.hasDismissable {
		return true
	}
	return p.dismissable
}

func (p Widget) showArrow() bool {
	return p.hasArrow && p.arrow
}

func (p Widget) flipEnabled() bool {
	if !p.hasShouldFlip {
		return true
	}
	return p.shouldFlip
}

func (p Widget) overflowAvoidanceEnabled() bool {
	if !p.hasAvoidOverflow {
		return true
	}
	return p.avoidOverflow
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.PopoverOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
