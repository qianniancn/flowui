package menu

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ContextMenuWidget struct {
	key               string
	trigger           frame.Widget
	menu              Widget
	open              bool
	hasOpen           bool
	defaultOpen       bool
	hasDefaultOpen    bool
	onOpenChange      func(bool)
	disabled          bool
	longPressDisabled bool
}

func ContextMenu(key string, trigger frame.Widget, menu Widget) ContextMenuWidget {
	return ContextMenuWidget{key: key, trigger: trigger, menu: menu}
}

func (c ContextMenuWidget) Open(open bool) ContextMenuWidget {
	c.open = open
	c.hasOpen = true
	return c
}

func (c ContextMenuWidget) DefaultOpen(open bool) ContextMenuWidget {
	c.defaultOpen = open
	c.hasDefaultOpen = true
	return c
}

func (c ContextMenuWidget) OnOpenChange(fn func(bool)) ContextMenuWidget {
	c.onOpenChange = fn
	return c
}

func (c ContextMenuWidget) Disabled(disabled bool) ContextMenuWidget {
	c.disabled = disabled
	return c
}

func (c ContextMenuWidget) LongPressDisabled(disabled bool) ContextMenuWidget {
	c.longPressDisabled = disabled
	return c
}

func (c ContextMenuWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := contextMenuStateFor(ctx, c.key)
	state.bind(c)
	open := state.isOpen(c)
	if c.disabled || !gtx.Enabled() {
		open = state.requestOpen(ctx, c, false)
	}

	dims := c.layoutTrigger(ctx, gtx, state, &open)
	if open && !state.hasAnchor {
		state.anchor = image.Rect(state.triggerSize.X/2, state.triggerSize.Y/2, state.triggerSize.X/2+1, state.triggerSize.Y/2+1)
		state.hasAnchor = true
		state.focusVisible = true
	}
	if open && !state.wasOpen {
		activateContextMenu(ctx, state)
	} else if !open {
		releaseContextMenu(ctx, state)
	}
	restoreFocus := !open && state.wasOpen && !c.disabled && gtx.Enabled() && !state.skipRestore
	state.observeOpen(open)
	if restoreFocus {
		frame.AfterOverlays(ctx, func() {
			if !frame.HasTopOverlay(ctx) {
				frame.RequestFocus(ctx, &state.trigger)
			}
		})
	}

	progress := state.progress(gtx, open && !c.disabled)
	if progress > 0 && state.hasAnchor {
		c.registerOverlay(ctx, state, open, progress, !gtx.Enabled())
	}
	return dims
}
