package input

import (
	"gioui.org/io/event"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotInputActionVisual = "input-action-visual"

type inputActionVisualCache struct {
	cache computed.ProviderCache[inputActionVisualKey, theme.InputActionVisual]
}

type inputActionVisualKey struct {
	overrideHash uint64
	overrideSet  bool
}

// GroupActionWidget is a compact clickable control for an input prefix
// or suffix. Its size and interaction style are stable across field heights.
type GroupActionWidget struct {
	override    *theme.InputActionOverride
	key         string
	child       core.Widget
	label       string
	onClick     func()
	disabled    bool
	focusTarget event.Tag
}

// GroupAction creates a compact text action for an input prefix or suffix.
// The text is also its accessible label. Use Content to display an icon or
// composite widget; the active Provider supplies the action's chrome.
func GroupAction(key, label string) GroupActionWidget {
	return GroupActionWidget{
		key:   key,
		label: label,
		child: text.New(label),
	}
}

// Content replaces the visible content, preserving the accessible label.
func (a GroupActionWidget) Content(value core.Widget) GroupActionWidget {
	a.child = value
	return a
}

// OnClick sets the action callback.
func (a GroupActionWidget) OnClick(fn func()) GroupActionWidget {
	a.onClick = fn
	return a
}

// Disabled controls whether the action can be activated.
func (a GroupActionWidget) Disabled(disabled bool) GroupActionWidget {
	a.disabled = disabled
	return a
}

// Label changes the accessible action label.
func (a GroupActionWidget) Label(label string) GroupActionWidget {
	a.label = label
	return a
}

func (a GroupActionWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	visual := inputGroupActionVisualCached(ctx, a.key, theme.InputActionRequest{Override: a.override})
	return a.layoutWithVisual(ctx, gtx, visual)
}

func (a GroupActionWidget) layoutWithVisual(ctx *core.Context, gtx layout.Context, visual theme.InputActionVisual) layout.Dimensions {
	click := core.Begin(ctx, gtx, a.key, nil, !a.disabled, true, func() {
		if a.focusTarget != nil {
			core.RequestFocus(ctx, a.focusTarget)
		}
		if a.onClick != nil {
			a.onClick()
		}
	})
	resolved := computed.Resolve(
		ctx,
		gtx,
		click.Key,
		click.State,
		style.Style{},
		style.Style{},
		style.Style{},
		visual.Declaration,
	)
	child := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if a.child == nil {
				return layout.Dimensions{}
			}
			return a.child.Layout(ctx, gtx)
		})
	})
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layoutui.Resolved(ctx, gtx, resolved, child)
	}, a.label)
}

func inputGroupActionVisual(activeTheme *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	return theme.ResolveVisual(activeTheme, request,
		theme.ComponentVisualProvider.InputAction, theme.ValidateInputActionVisual)
}

func inputGroupActionVisualCached(ctx *core.Context, key string, request theme.InputActionRequest) theme.InputActionVisual {
	activeTheme := core.ActiveTheme(ctx)
	overrideHash, overrideSet := inputActionOverrideKey(request.Override)
	cache := core.UseState[inputActionVisualCache](ctx, core.FullKey(ctx, key), stateSlotInputActionVisual)
	core.RequireVisualProvider(ctx)
	cacheKey := inputActionVisualKey{overrideHash: overrideHash, overrideSet: overrideSet}
	return cache.cache.Resolve(activeTheme, cacheKey, func() theme.InputActionVisual {
		return inputGroupActionVisual(activeTheme, request)
	})
}

func inputActionOverrideKey(override *theme.InputActionOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	hash := override.Declaration.Hash64()
	if hash == (style.Style{}).Hash64() {
		return 0, false
	}
	return hash, true
}

func (a GroupActionWidget) withFocusTarget(target event.Tag) GroupActionWidget {
	a.focusTarget = target
	return a
}

func inputGroupActionWithFocus(widget core.Widget, target event.Tag) core.Widget {
	switch action := widget.(type) {
	case GroupActionWidget:
		return action.withFocusTarget(target)
	default:
		return widget
	}
}

// Style merges instance declarations into the group action.
func (a GroupActionWidget) Style(values ...style.Style) GroupActionWidget {
	declaration := style.Join(values...)
	if a.override == nil {
		a.override = &theme.InputActionOverride{}
	}
	a.override.Declaration = style.Join(a.override.Declaration, declaration)
	return a
}
