package alert

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/closebutton"
	"github.com/qianniancn/flowui/widget/icon"
	textui "github.com/qianniancn/flowui/widget/text"
)

func (a Widget) layoutWithVisual(ctx *core.Context, gtx layout.Context, visual theme.AlertVisual) layout.Dimensions {
	resolved := a.resolveStyle(ctx, gtx, visual)
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if a.title != "" {
			semantic.LabelOp(a.title).Add(gtx.Ops)
		}
		if a.description != "" && a.content == nil {
			semantic.DescriptionOp(a.description).Add(gtx.Ops)
		}
		return a.layoutRow(ctx, gtx, resolved, visual)
	})
	return layoutui.Resolved(ctx, gtx, resolved.root, content)
}

func (a Widget) layoutRow(ctx *core.Context, gtx layout.Context, resolved alertResolvedStyle, visual theme.AlertVisual) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return a.layoutIndicator(ctx, gtx, resolved.indicator, visual)
	}))
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return a.layoutContent(ctx, gtx, resolved)
	}))
	if a.action != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.action.Layout(ctx, gtx)
		}))
	}
	if a.closable {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return closebutton.New(a.closeKey()).
				OnClick(a.onClose).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Start,
		Gap:       gtx.Dp(visual.Metrics.Gap),
	}.Layout(gtx, children...)
}

func (a Widget) closeKey() string {
	if a.key != "" {
		return "alert-close-" + a.key
	}
	return "alert-close-" + a.title + "\x00" + a.description
}

func (a Widget) layoutIndicator(ctx *core.Context, gtx layout.Context, resolved style.ResolvedStyle, visual theme.AlertVisual) layout.Dimensions {
	return layout.Inset{
		Top:    visual.Metrics.IndicatorPadding,
		Right:  visual.Metrics.IndicatorPadding,
		Bottom: visual.Metrics.IndicatorPadding,
		Left:   visual.Metrics.IndicatorPadding,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.indicator != nil {
			return layoutui.Resolved(ctx, gtx, resolved, a.indicator)
		}
		size := gtx.Dp(visual.Metrics.IconSize)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(size, size))
		return layoutui.Resolved(ctx, iconGtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			if len(visual.Icon) == 0 {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}
			return icon.Layout(visual.Icon, gtx, ctx.ForegroundColor())
		}))
	})
}

func (a Widget) layoutContent(ctx *core.Context, gtx layout.Context, style alertResolvedStyle) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 2)
	if a.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutAlertLineBox(gtx, styleLineHeight(style.title), func(gtx layout.Context) layout.Dimensions {
				return layoutui.Resolved(ctx, gtx, style.title, textui.New(a.title))
			})
		}))
	}
	if a.content != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, style.content, a.content)
		}))
	} else if a.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutAlertLineBox(gtx, styleLineHeight(style.description), func(gtx layout.Context) layout.Dimensions {
				return layoutui.Resolved(ctx, gtx, style.description, textui.New(a.description))
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx, children...)
}

func styleLineHeight(value style.ResolvedStyle) unit.Sp {
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
