package closebutton

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/components/icon"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/interact"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui-icons-lucide"
)

const (
	closeButtonColorDuration = 100 * time.Millisecond
	closeButtonScaleDuration = 250 * time.Millisecond
)

type CloseButtonWidget struct {
	key         string
	onClick     func()
	disabled    bool
	icon        frame.Widget
	label       string
	customStyle flowstyle.Style
}

func CloseButton(key string) CloseButtonWidget {
	return CloseButtonWidget{key: key}
}

func (b CloseButtonWidget) OnClick(fn func()) CloseButtonWidget {
	b.onClick = fn
	return b
}

func (b CloseButtonWidget) Disabled(disabled bool) CloseButtonWidget {
	b.disabled = disabled
	return b
}

func (b CloseButtonWidget) Icon(icon frame.Widget) CloseButtonWidget {
	b.icon = icon
	return b
}

func (b CloseButtonWidget) Label(label string) CloseButtonWidget {
	b.label = label
	return b
}

func (b CloseButtonWidget) Style(value flowstyle.Style) CloseButtonWidget {
	b.customStyle = value
	return b
}

func (b CloseButtonWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, nil, true)
}

// LayoutWithClickableNoEvents renders a close button with caller-owned state and events.
func LayoutWithClickableNoEvents(b CloseButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, false)
}

func layoutWithClickable(b CloseButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable, handleEvents bool) layout.Dimensions {
	click := interact.Begin(ctx, gtx, b.key, clickable, !b.disabled, handleEvents, b.onClick)
	style := b.resolveStyle(ctx, gtx, click.Key, click.StyleState)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			child := b.icon
			if child == nil {
				child = frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
					return icon.Layout(lucide.X, gtx, ctx.ForegroundColor())
				})
			}
			return layoutui.LayoutResolved(ctx, gtx, style.icon, child)
		})
	})

	return layoutui.LayoutInteractiveResolved(ctx, gtx, style.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return click.Layout(gtx, visual, b.semanticLabel(ctx))
	})
}

func (b CloseButtonWidget) semanticLabel(ctx *frame.Context) string {
	if b.label != "" {
		return b.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "关闭"
	}
	return "Close"
}
