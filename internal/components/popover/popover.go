package popover

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

type PopoverWidget struct {
	key                     string
	open                    bool
	trigger                 frame.Widget
	content                 frame.Widget
	heading                 string
	onOpenChange            func(bool)
	placement               overlay.PopoverPlacement
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

const (
	popoverEnterDuration = 150 * time.Millisecond
	popoverExitDuration  = 100 * time.Millisecond
)

func Popover(key string, open bool, trigger frame.Widget, content frame.Widget) PopoverWidget {
	return PopoverWidget{
		key:     key,
		open:    open,
		trigger: trigger,
		content: content,
	}
}

func (p PopoverWidget) OnOpenChange(fn func(bool)) PopoverWidget {
	p.onOpenChange = fn
	return p
}

func (p PopoverWidget) Heading(heading string) PopoverWidget {
	p.heading = heading
	return p
}

func (p PopoverWidget) Placement(placement overlay.PopoverPlacement) PopoverWidget {
	p.placement = placement
	return p
}

func (p PopoverWidget) Offset(dp int) PopoverWidget {
	p.offset = dp
	p.hasOffset = true
	return p
}

func (p PopoverWidget) ShouldFlip(shouldFlip bool) PopoverWidget {
	p.shouldFlip = shouldFlip
	p.hasShouldFlip = true
	return p
}

func (p PopoverWidget) AvoidOverflow(avoidOverflow bool) PopoverWidget {
	p.avoidOverflow = avoidOverflow
	p.hasAvoidOverflow = true
	return p
}

func (p PopoverWidget) Arrow(show bool) PopoverWidget {
	p.arrow = show
	p.hasArrow = true
	return p
}

func (p PopoverWidget) Dismissable(dismissable bool) PopoverWidget {
	p.dismissable = dismissable
	p.hasDismissable = true
	return p
}

func (p PopoverWidget) KeyboardDismissDisabled(disabled bool) PopoverWidget {
	p.keyboardDismissDisabled = disabled
	return p
}

func (p PopoverWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	fullKey := frame.FullKey(ctx, p.key)
	naturallyDisabled := frame.OverlayNaturallyDisabled(gtx)
	triggerDims := p.layoutTrigger(ctx, gtx)
	if !p.open && !hasVisiblePopover(ctx, fullKey) {
		return triggerDims
	}

	state := popoverStateFor(ctx, p.key)
	progress := state.progress(gtx, p.open)
	if !p.open && progress <= 0 {
		deletePopoverState(ctx, fullKey)
		return triggerDims
	}

	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       fullKey,
		Layer:     frame.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: triggerDims.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			if p.open && interactive && gtx.Enabled() {
				p.handleCloseEvents(ctx, gtx, state)
			}
			return p.layoutOverlay(ctx, gtx, state, anchor, progress, p.open && gtx.Enabled())
		},
	})

	return triggerDims
}

func (p PopoverWidget) layoutTrigger(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if p.trigger == nil {
		return layout.Dimensions{}
	}
	return p.trigger.Layout(ctx, gtx)
}

func (p PopoverWidget) handleCloseEvents(ctx *frame.Context, gtx layout.Context, popoverStateValue *popoverState) {
	for popoverStateValue.dialog.Clicked(gtx) {
	}
	if popoverStateValue.dialog.TakePressed() {
		frame.PreserveFocus(ctx)
	}
	for popoverStateValue.arrow.Clicked(gtx) {
	}
	if popoverStateValue.arrow.TakePressed() {
		frame.PreserveFocus(ctx)
	}
	if !p.keyboardDismissDisabled && popoverStateValue.escapePressed(gtx) {
		p.requestClose()
	}
	for i := range popoverStateValue.dismiss {
		for popoverStateValue.dismiss[i].Clicked(gtx) {
			if p.isDismissable() {
				p.requestClose()
			}
		}
		if popoverStateValue.dismiss[i].TakePressed() {
			frame.PreserveFocus(ctx)
		}
	}
}

func (p PopoverWidget) requestClose() {
	if p.onOpenChange != nil {
		p.onOpenChange(false)
	}
}

func (p PopoverWidget) isDismissable() bool {
	if !p.hasDismissable {
		return true
	}
	return p.dismissable
}

func (p PopoverWidget) showArrow() bool {
	return p.hasArrow && p.arrow
}

func (p PopoverWidget) flipEnabled() bool {
	if !p.hasShouldFlip {
		return true
	}
	return p.shouldFlip
}

func (p PopoverWidget) overflowAvoidanceEnabled() bool {
	if !p.hasAvoidOverflow {
		return true
	}
	return p.avoidOverflow
}
