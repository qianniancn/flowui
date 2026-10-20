package statusbar

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Widget presents compact application state at the bottom of a window.
type Widget struct {
	override    *theme.StatusBarOverride
	left        core.Widget
	right       core.Widget
	visualCache *statusBarVisualCache
}

func New(left, right core.Widget) Widget {
	return Widget{left: left, right: right, visualCache: &statusBarVisualCache{}}
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.StatusBarOverride) layout.Dimensions {
	visual := w.visualWithCache(ctx, override)
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(visual.Metrics.Height)))
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)
	children := make([]core.Widget, 0, 3)
	if w.left != nil {
		children = append(children, w.left)
	}
	children = append(children, layoutui.Expanded(layoutui.Spacer(0, 0)))
	if w.right != nil {
		children = append(children, w.right)
	}
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.TrackedDirection(ctx, gtx, layout.Center, func(gtx layout.Context) layout.Dimensions {
			return layoutui.TrackedFlex(ctx, gtx, layout.Horizontal, visual.Metrics.Gap, layout.Middle, children...)
		})
	})
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		visual.Override,
	)
	if len(resolved.Transitions) != 0 {
		resolved = computed.ApplyTransitions(ctx, rootGtx, core.ClaimKey(ctx, core.KindStyle, "status-bar"), resolved)
	}
	return layoutui.Resolved(ctx, rootGtx, resolved, content)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.StatusBarOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
