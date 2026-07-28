package alert

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/components/icon"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	textui "github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui-icons-lucide"
)

func (a Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	resolved := a.resolveStyle(ctx, gtx)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if a.title != "" {
			semantic.LabelOp(a.title).Add(gtx.Ops)
		}
		if a.description != "" && a.content == nil {
			semantic.DescriptionOp(a.description).Add(gtx.Ops)
		}
		return a.layoutRow(ctx, gtx, resolved)
	})
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, content)
}

func (a Widget) layoutRow(ctx *frame.Context, gtx layout.Context, style alertResolvedStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Alert
	children := make([]layout.FlexChild, 0, 3)
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return a.layoutIndicator(ctx, gtx, style.indicator)
	}))
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return a.layoutContent(ctx, gtx, style)
	}))
	if a.action != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.action.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Start,
		Gap:       gtx.Dp(tokens.Gap),
	}.Layout(gtx, children...)
}

func (a Widget) layoutIndicator(ctx *frame.Context, gtx layout.Context, style flowstyle.ResolvedStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Alert
	return layout.Inset{
		Top:    tokens.IndicatorPadding,
		Right:  tokens.IndicatorPadding,
		Bottom: tokens.IndicatorPadding,
		Left:   tokens.IndicatorPadding,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.indicator != nil {
			return layoutui.LayoutResolved(ctx, gtx, style, a.indicator)
		}
		size := gtx.Dp(tokens.IconSize)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(size, size))
		return layoutui.LayoutResolved(ctx, iconGtx, style, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return icon.Layout(alertIcon(a.status), gtx, ctx.ForegroundColor())
		}))
	})
}

func (a Widget) layoutContent(ctx *frame.Context, gtx layout.Context, style alertResolvedStyle) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 2)
	if a.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutAlertLineBox(gtx, styleLineHeight(style.title), func(gtx layout.Context) layout.Dimensions {
				return layoutui.LayoutResolved(ctx, gtx, style.title, textui.New(a.title))
			})
		}))
	}
	if a.content != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, style.content, a.content)
		}))
	} else if a.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutAlertLineBox(gtx, styleLineHeight(style.description), func(gtx layout.Context) layout.Dimensions {
				return layoutui.LayoutResolved(ctx, gtx, style.description, textui.New(a.description))
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx, children...)
}

func styleLineHeight(value flowstyle.ResolvedStyle) unit.Sp {
	if value.Text != nil && value.Text.LineHeight != nil {
		return *value.Text.LineHeight
	}
	return 0
}

func layoutAlertLineBox(gtx layout.Context, lineHeight unit.Sp, child layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	childGtx := gtx
	childGtx.Constraints.Min.Y = 0
	dims := child(childGtx)
	call := macro.Stop()
	height := min(max(dims.Size.Y, gtx.Sp(lineHeight)), gtx.Constraints.Max.Y)
	offset := op.Offset(image.Pt(0, max((height-dims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	dims.Size.Y = height
	return dims
}

func alertIcon(status Status) []byte {
	switch status {
	case StatusSuccess:
		return lucide.CircleCheck
	case StatusWarning:
		return lucide.TriangleAlert
	case StatusDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
