package tooltip

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

type Trigger uint8

const (
	Hover Trigger = iota
	Focus
)

type Widget struct {
	override         *theme.TooltipOverride
	key              string
	trigger          core.Widget
	content          core.Widget
	placement        core.PopoverPlacement
	triggerMode      Trigger
	offset           int
	hasOffset        bool
	delay            time.Duration
	hasDelay         bool
	closeDelay       time.Duration
	hasCloseDelay    bool
	shouldFlip       bool
	hasShouldFlip    bool
	avoidOverflow    bool
	hasAvoidOverflow bool
	arrow            bool
	disabled         bool
}

// New attaches a text tooltip to trigger. Use Content for rich tooltip content.
func New(key string, trigger core.Widget, label string) Widget {
	return Widget{
		key:       key,
		trigger:   trigger,
		content:   text.New(label),
		placement: core.PopoverTop,
	}
}

// Content replaces the tooltip text with custom content.
func (t Widget) Content(value core.Widget) Widget {
	t.content = value
	return t
}

func (t Widget) Placement(placement core.PopoverPlacement) Widget {
	t.placement = placement
	return t
}

func (t Widget) Trigger(trigger Trigger) Widget {
	t.triggerMode = trigger
	return t
}

func (t Widget) Offset(dp int) Widget {
	t.offset = dp
	t.hasOffset = true
	return t
}

func (t Widget) Delay(delay time.Duration) Widget {
	t.delay = max(delay, 0)
	t.hasDelay = true
	return t
}

func (t Widget) CloseDelay(delay time.Duration) Widget {
	t.closeDelay = max(delay, 0)
	t.hasCloseDelay = true
	return t
}

func (t Widget) ShouldFlip(shouldFlip bool) Widget {
	t.shouldFlip = shouldFlip
	t.hasShouldFlip = true
	return t
}

func (t Widget) AvoidOverflow(avoidOverflow bool) Widget {
	t.avoidOverflow = avoidOverflow
	t.hasAvoidOverflow = true
	return t
}

func (t Widget) Arrow(show bool) Widget {
	t.arrow = show
	return t
}

func (t Widget) Disabled(disabled bool) Widget {
	t.disabled = disabled
	return t
}

func (t Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return t.layoutWithOverride(ctx, gtx, t.override)
}

// Styled applies an instance declaration at the Tooltip Provider boundary.
// Tooltip behavior itself does not retain visual state.
func (t Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.TooltipOverride) layout.Dimensions {
	coordinator := tooltipCoordinatorFor(ctx)
	coordinator.update(gtx)
	fullKey, state := tooltipStateFor(ctx, t.key)
	coordinator.register(ctx, gtx, fullKey)
	visual := tooltipVisualWithCache(ctx, override, &state.visualCache)
	disabled := t.disabled || !gtx.Enabled()
	wasActive := state.active
	wasOpen := state.open
	delay := t.showDelay(visual)
	if coordinator.warmed {
		delay = 0
	}
	closeDelay := t.hideDelay(visual)
	state.update(gtx, t.triggerMode, disabled, delay, closeDelay)
	if wasActive && !state.active {
		coordinator.beginCooldown(gtx, fullKey, closeDelay)
	}
	if state.open && !wasOpen {
		coordinator.open(fullKey, closeDelay)
		core.ActivateExclusive(ctx, tooltipExclusiveGroup, fullKey)
	} else if wasOpen && !state.open {
		core.ReleaseExclusive(ctx, tooltipExclusiveGroup, fullKey)
	}

	triggerDims := t.layoutTrigger(ctx, gtx)
	state.addInput(gtx, triggerDims.Size, disabled)

	progress := state.progress(gtx, visual.Motion, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration)
	if progress <= 0 && !state.open {
		return triggerDims
	}

	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       fullKey,
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: triggerDims.Size},
		HasAnchor: true,
		Disabled:  true,
		Passive:   true,
		Layout: func(gtx layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			popup := t.popup(progress, state.exiting())
			popup.styleKey = fullKey
			popup.styleState = style.State{Disabled: disabled, Open: state.open}
			return popup.layoutWithVisual(ctx, gtx, anchor, visual)
		},
	})
	return triggerDims
}

func (t Widget) layoutTrigger(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if t.trigger == nil {
		return layout.Dimensions{}
	}
	return t.trigger.Layout(ctx, gtx)
}

func (t Widget) showDelay(visual theme.TooltipVisual) time.Duration {
	if t.hasDelay {
		return t.delay
	}
	return max(visual.Metrics.Delay, 0)
}

func (t Widget) hideDelay(visual theme.TooltipVisual) time.Duration {
	if t.hasCloseDelay {
		return t.closeDelay
	}
	return max(visual.Metrics.CloseDelay, 0)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
