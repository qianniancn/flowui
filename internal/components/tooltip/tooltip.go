package tooltip

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

type TooltipTrigger uint8

const (
	TooltipHover TooltipTrigger = iota
	TooltipFocus
)

const (
	tooltipEnterDuration = 150 * time.Millisecond
	tooltipExitDuration  = 100 * time.Millisecond
)

type TooltipWidget struct {
	key              string
	trigger          frame.Widget
	content          frame.Widget
	placement        overlay.PopoverPlacement
	triggerMode      TooltipTrigger
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
	hasArrow         bool
	disabled         bool
}

func Tooltip(key string, trigger frame.Widget, content frame.Widget) TooltipWidget {
	return TooltipWidget{
		key:       key,
		trigger:   trigger,
		content:   content,
		placement: overlay.PopoverTop,
	}
}

func (t TooltipWidget) Placement(placement overlay.PopoverPlacement) TooltipWidget {
	t.placement = placement
	return t
}

func (t TooltipWidget) Trigger(trigger TooltipTrigger) TooltipWidget {
	t.triggerMode = trigger
	return t
}

func (t TooltipWidget) Offset(dp int) TooltipWidget {
	t.offset = dp
	t.hasOffset = true
	return t
}

func (t TooltipWidget) Delay(delay time.Duration) TooltipWidget {
	t.delay = max(delay, 0)
	t.hasDelay = true
	return t
}

func (t TooltipWidget) CloseDelay(delay time.Duration) TooltipWidget {
	t.closeDelay = max(delay, 0)
	t.hasCloseDelay = true
	return t
}

func (t TooltipWidget) ShouldFlip(shouldFlip bool) TooltipWidget {
	t.shouldFlip = shouldFlip
	t.hasShouldFlip = true
	return t
}

func (t TooltipWidget) AvoidOverflow(avoidOverflow bool) TooltipWidget {
	t.avoidOverflow = avoidOverflow
	t.hasAvoidOverflow = true
	return t
}

func (t TooltipWidget) Arrow(show bool) TooltipWidget {
	t.arrow = show
	t.hasArrow = true
	return t
}

func (t TooltipWidget) Disabled(disabled bool) TooltipWidget {
	t.disabled = disabled
	return t
}

func (t TooltipWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	coordinator := tooltipCoordinatorFor(ctx)
	coordinator.update(gtx)
	fullKey, state := tooltipStateFor(ctx, t.key)
	coordinator.register(ctx, gtx, fullKey)
	disabled := t.disabled || !gtx.Enabled()
	wasActive := state.active
	wasOpen := state.open
	delay := t.showDelay(ctx)
	if coordinator.warmed {
		delay = 0
	}
	closeDelay := t.hideDelay(ctx)
	state.update(gtx, t.triggerMode, disabled, delay, closeDelay)
	if wasActive && !state.active {
		coordinator.beginCooldown(gtx, fullKey, closeDelay)
	}
	if state.open && !wasOpen {
		coordinator.open(fullKey, closeDelay)
		frame.ActivateExclusive(ctx, tooltipExclusiveGroup, fullKey)
	} else if wasOpen && !state.open {
		frame.ReleaseExclusive(ctx, tooltipExclusiveGroup, fullKey)
	}

	triggerDims := t.layoutTrigger(ctx, gtx)
	state.addInput(gtx, triggerDims.Size, disabled)

	progress := state.progress(gtx)
	if progress <= 0 && !state.open {
		return triggerDims
	}

	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       fullKey,
		Layer:     frame.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: triggerDims.Size},
		HasAnchor: true,
		Disabled:  true,
		Passive:   true,
		Layout: func(gtx layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			return t.layoutOverlay(ctx, gtx, state, anchor, progress)
		},
	})
	return triggerDims
}

func (t TooltipWidget) layoutTrigger(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if t.trigger == nil {
		return layout.Dimensions{}
	}
	return t.trigger.Layout(ctx, gtx)
}

func (t TooltipWidget) showArrow() bool {
	return t.hasArrow && t.arrow
}

func (t TooltipWidget) flipEnabled() bool {
	return !t.hasShouldFlip || t.shouldFlip
}

func (t TooltipWidget) overflowAvoidanceEnabled() bool {
	return !t.hasAvoidOverflow || t.avoidOverflow
}

func (t TooltipWidget) showDelay(ctx *frame.Context) time.Duration {
	if t.hasDelay {
		return t.delay
	}
	return max(frame.ActiveTheme(ctx).Components.Tooltip.Delay, 0)
}

func (t TooltipWidget) hideDelay(ctx *frame.Context) time.Duration {
	if t.hasCloseDelay {
		return t.closeDelay
	}
	return max(frame.ActiveTheme(ctx).Components.Tooltip.CloseDelay, 0)
}
