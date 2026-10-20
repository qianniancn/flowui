package label

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type labelVisualKey struct {
	Foreground color.NRGBA
}

type labelVisualCache struct {
	cache computed.ProviderCache[labelVisualKey, theme.LabelVisual]
}

func (l Widget) layoutWithOverride(
	ctx *core.Context,
	gtx layout.Context,
	override *theme.LabelOverride,
) layout.Dimensions {
	disabled := l.disabled || !gtx.Enabled()
	state := style.State{Disabled: disabled, Invalid: l.invalid}
	visual := l.visualWithCache(ctx, override)
	if l.forKey == "" {
		resolved := l.resolveLayoutStyle(ctx, gtx, state, visual)
		return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return l.layoutContent(ctx, gtx, resolved, visual.RequiredMarkGap)
		}))
	}

	fieldKey := l.registerFieldAssociation(ctx)
	key, clickable := core.DerivedClickableWithKey(ctx, l.forKey, "label")
	presses := core.SnapshotPresses(clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			core.RequestFieldFocusVisible(ctx, fieldKey, presses.ClickFocusVisible(clickable.History()))
		}
	}
	state.Hovered = clickable.Hovered()
	state.Pressed = clickable.Pressed()
	state.Focused = gtx.Focused(clickable)
	state.FocusVisible = core.FocusVisible(ctx, clickable, state.Focused)
	resolved := l.resolveStyle(ctx, gtx, key, state, visual)
	child := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return l.layoutContent(ctx, gtx, resolved, visual.RequiredMarkGap)
	})
	return layoutui.InteractiveResolved(ctx, gtx, resolved.root, child, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return clickable.Layout(gtx, visual)
	})
}
