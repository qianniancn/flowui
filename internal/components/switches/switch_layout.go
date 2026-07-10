package switches

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (s SwitchWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style switchStyle, size switchSizeStyle, checked bool) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return s.layoutRow(ctx, gtx, style, size, checked)
	}
	if s.description == "" {
		return row(gtx)
	}
	theme := frame.ActiveTheme(ctx).Components.Switch
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(theme.DescriptionGap),
	}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutDescription(ctx, gtx, size)
		}),
	)
}

func (s SwitchWidget) layoutRow(ctx *frame.Context, gtx layout.Context, style switchStyle, size switchSizeStyle, checked bool) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return s.layoutControl(ctx, gtx, style, size, checked)
	}
	if s.label == "" {
		return control(gtx)
	}
	label := func(gtx layout.Context) layout.Dimensions {
		return s.layoutLabel(ctx, gtx, style)
	}
	first := layout.Rigid(control)
	second := layout.Rigid(label)
	if s.labelBefore {
		first = layout.Rigid(label)
		second = layout.Rigid(control)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(frame.ActiveTheme(ctx).Components.Switch.ContentGap),
	}.Layout(gtx, first, second)
}

func (s SwitchWidget) layoutControl(ctx *frame.Context, gtx layout.Context, style switchStyle, size switchSizeStyle, checked bool) layout.Dimensions {
	result := drawSwitch(gtx, frame.ActiveTheme(ctx), style, size)
	if s.thumb == nil || result.thumbRect.Empty() {
		return result.dims
	}
	child := s.thumb(checked)
	if child == nil {
		return result.dims
	}
	child = s.styleThumbContent(ctx, style, child)
	childGtx := gtx
	childGtx.Constraints = layout.Exact(result.thumbRect.Size())
	stack := op.Offset(result.thumbRect.Min).Push(gtx.Ops)
	layout.Center.Layout(childGtx, func(gtx layout.Context) layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
	stack.Pop()
	return result.dims
}

func (s SwitchWidget) styleThumbContent(ctx *frame.Context, style switchStyle, child frame.Widget) frame.Widget {
	text, ok := child.(text.Widget)
	if !ok {
		return child
	}
	text = text.DefaultColor(switchThumbContentColor(style))
	text = text.DefaultSize(float32(frame.ActiveTheme(ctx).Typography.SmallSize))
	text = text.DefaultWeight(font.Medium)
	return text
}

func (s SwitchWidget) layoutLabel(ctx *frame.Context, gtx layout.Context, style switchStyle) layout.Dimensions {
	return text.New(s.label).
		Size(float32(frame.ActiveTheme(ctx).Components.Switch.TextSize)).
		Weight(font.Medium).
		Color(style.label).
		Layout(ctx, gtx)
}

func (s SwitchWidget) layoutDescription(ctx *frame.Context, gtx layout.Context, size switchSizeStyle) layout.Dimensions {
	left := unit.Dp(0)
	if s.label != "" && !s.labelBefore {
		theme := frame.ActiveTheme(ctx).Components.Switch
		left = size.trackWidth + theme.FocusSpace*2 + theme.ContentGap
	}
	return layout.Inset{Left: left}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return description.Description(s.description).
			For(s.key).
			Disabled(s.disabled).
			Layout(ctx, gtx)
	})
}
