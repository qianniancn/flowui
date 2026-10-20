package menu

import (
	"github.com/qianniancn/flowui/style"
	"image"

	"gioui.org/io/event"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type ContextMenuWidget struct {
	override          *theme.MenuOverride
	key               string
	trigger           core.Widget
	menu              Widget
	open              bool
	hasOpen           bool
	defaultOpen       bool
	hasDefaultOpen    bool
	onOpenChange      func(bool)
	focusTarget       event.Tag
	focusTargets      []event.Tag
	disabled          bool
	longPressDisabled bool
}

func ContextMenu(key string, trigger core.Widget, menu Widget) ContextMenuWidget {
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

// FocusTarget sets the focusable trigger restored after the menu closes.
func (c ContextMenuWidget) FocusTarget(target event.Tag) ContextMenuWidget {
	c.focusTarget = target
	return c
}

// FocusTargets adds focusable descendants that can open the menu with
// Shift+F10 and receive focus again after it closes.
func (c ContextMenuWidget) FocusTargets(targets ...event.Tag) ContextMenuWidget {
	c.focusTargets = append([]event.Tag(nil), targets...)
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

func (c ContextMenuWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layoutWithMenuOverride(ctx, gtx, c.override)
}

func (c ContextMenuWidget) layoutWithMenuOverride(ctx *core.Context, gtx layout.Context, menuOverride *theme.MenuOverride) layout.Dimensions {
	state := contextMenuStateFor(ctx, c.key)
	state.bind(c)
	open := state.isOpen(c)
	if c.disabled || !gtx.Enabled() {
		open = state.requestOpen(ctx, c, false)
	}

	dims := c.layoutTrigger(ctx, gtx, state, &open)
	if open && !state.wasOpen {
		state.focusTarget = c.focusedTarget(gtx, state)
	}
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
		core.AfterOverlays(ctx, func() {
			if !core.HasTopOverlay(ctx) {
				core.RequestFocusVisible(ctx, c.triggerFocusTarget(state), state.focusVisible)
			}
		})
	}

	visual := c.menu.visualWithCache(ctx, menuOverride, &state.visualCache)
	progress := state.progressFor(gtx, open && !c.disabled, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration, visual.Motion)
	if progress > 0 && state.hasAnchor {
		c.registerOverlay(
			ctx,
			state,
			open,
			progress,
			!gtx.Enabled(),
			visual,
		)
	}
	return dims
}

func (c ContextMenuWidget) triggerFocusTarget(state *contextMenuState) event.Tag {
	if state.focusTarget != nil {
		return state.focusTarget
	}
	return c.defaultFocusTarget(state)
}

func (c ContextMenuWidget) defaultFocusTarget(state *contextMenuState) event.Tag {
	if c.focusTarget != nil {
		return c.focusTarget
	}
	return &state.trigger
}

func (c ContextMenuWidget) focusedTarget(gtx layout.Context, state *contextMenuState) event.Tag {
	if target := c.defaultFocusTarget(state); gtx.Focused(target) {
		return target
	}
	for _, target := range c.focusTargets {
		if target != nil && gtx.Focused(target) {
			return target
		}
	}
	return c.defaultFocusTarget(state)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ContextMenuWidget) Style(values ...style.Style) ContextMenuWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.MenuOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
