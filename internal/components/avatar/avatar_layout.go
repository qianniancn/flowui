package avatar

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/components/icon"
	imageview "github.com/qianniancn/flowui/internal/components/image"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	textui "github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui-icons-lucide"
)

func (a Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	resolved := a.resolveStyle(ctx, gtx)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if label := a.semanticLabel(); label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		if a.image.Size() != (image.Point{}) {
			return a.layoutImage(ctx, gtx)
		}
		return a.layoutFallback(ctx, gtx, resolved)
	})
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, content)
}

func (a Widget) layoutImage(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return imageview.New(a.image).
		Fit(imageview.FitCover).
		Alt(a.semanticLabel()).
		Layout(ctx, gtx)
}

func (a Widget) layoutFallback(ctx *frame.Context, gtx layout.Context, resolved avatarResolvedStyle) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.fallback != nil {
			return layoutui.LayoutResolved(ctx, gtx, resolved.label, a.fallback)
		}
		if a.fallbackText != "" {
			return layoutui.LayoutResolved(ctx, gtx, resolved.label, textui.New(a.fallbackText))
		}
		return layoutui.LayoutResolved(ctx, gtx, resolved.icon, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			size := resolvedPartSize(resolved.icon)
			return icon.New(lucide.UserRound).Size(float32(size)).Color(ctx.ForegroundColor()).Layout(ctx, gtx)
		}))
	})
}

func resolvedPartSize(value flowstyle.ResolvedStyle) float32 {
	if value.Box != nil {
		if value.Box.Width != nil {
			return float32(*value.Box.Width)
		}
		if value.Box.Height != nil {
			return float32(*value.Box.Height)
		}
	}
	return 0
}

func (a Widget) semanticLabel() string {
	if a.alt != "" {
		return a.alt
	}
	return a.fallbackText
}
