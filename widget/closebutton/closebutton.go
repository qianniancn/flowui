package closebutton

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	override    *theme.CloseButtonOverride
	key         string
	onClick     func()
	disabled    bool
	icon        core.Widget
	label       string
	visualCache *closeButtonVisualCache
}

func New(key string) Widget {
	return Widget{key: key, visualCache: &closeButtonVisualCache{}}
}

func (b Widget) OnClick(fn func()) Widget {
	b.onClick = fn
	return b
}

func (b Widget) Disabled(disabled bool) Widget {
	b.disabled = disabled
	return b
}

func (b Widget) Icon(icon core.Widget) Widget {
	b.icon = icon
	return b
}

func (b Widget) Label(label string) Widget {
	b.label = label
	return b
}

func (b Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, nil, b.override, true)
}

// LayoutWithClickableNoEvents renders a close button with caller-owned state and events.
func LayoutWithClickableNoEvents(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, nil, false)
}

func layoutWithClickable(
	b Widget,
	ctx *core.Context,
	gtx layout.Context,
	clickable *widget.Clickable,
	override *theme.CloseButtonOverride,
	handleEvents bool,
) layout.Dimensions {
	click := core.Begin(ctx, gtx, b.key, clickable, !b.disabled, handleEvents, b.onClick)
	visual := b.visualWithCache(ctx, override)
	resolved := b.resolveStyle(ctx, gtx, click.Key, click.State, visual)
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			child := b.icon
			if child == nil {
				child = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					return visual.IconRenderer(gtx, ctx.ForegroundColor())
				})
			}
			return layoutui.Resolved(ctx, gtx, resolved.icon, child)
		})
	})

	return layoutui.InteractiveResolved(ctx, gtx, resolved.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return click.Layout(gtx, visual, b.semanticLabel(ctx))
	})
}

func (b Widget) semanticLabel(ctx *core.Context) string {
	if b.label != "" {
		return b.label
	}
	return ctx.Text(i18n.ActionClose)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.

// Style merges instance declarations into the widget.
func (b Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if b.override == nil {
		b.override = &theme.CloseButtonOverride{}
	}
	b.override.Declaration = style.Join(b.override.Declaration, declaration)
	return b
}
