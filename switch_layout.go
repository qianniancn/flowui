package flowui

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func (s SwitchWidget) layoutContent(ctx *Context, gtx layout.Context, style switchStyle, size switchSizeStyle, checked bool) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return s.layoutRow(ctx, gtx, style, size, checked)
	}
	if s.description == "" {
		return row(gtx)
	}
	theme := ctx.Theme.Components.Switch
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(theme.DescriptionGap),
	}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutDescription(ctx, gtx, style, size)
		}),
	)
}

func (s SwitchWidget) layoutRow(ctx *Context, gtx layout.Context, style switchStyle, size switchSizeStyle, checked bool) layout.Dimensions {
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
		Gap:       gtx.Dp(ctx.Theme.Components.Switch.ContentGap),
	}.Layout(gtx, first, second)
}

func (s SwitchWidget) layoutControl(ctx *Context, gtx layout.Context, style switchStyle, size switchSizeStyle, checked bool) layout.Dimensions {
	result := drawSwitch(gtx, ctx.Theme, style, size)
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

func (s SwitchWidget) styleThumbContent(ctx *Context, style switchStyle, child Widget) Widget {
	text, ok := child.(TextWidget)
	if !ok {
		return child
	}
	if !text.hasColor {
		text = text.Color(switchThumbContentColor(style))
	}
	if text.size == 0 {
		text = text.Size(float32(ctx.Theme.Typography.SmallSize))
	}
	if text.weight == 0 {
		text = text.Weight(font.Medium)
	}
	return text
}

func (s SwitchWidget) layoutLabel(ctx *Context, gtx layout.Context, style switchStyle) layout.Dimensions {
	return Text(s.label).
		Size(float32(ctx.Theme.Components.Switch.TextSize)).
		Weight(font.Medium).
		Color(style.label).
		Layout(ctx, gtx)
}

func (s SwitchWidget) layoutDescription(ctx *Context, gtx layout.Context, style switchStyle, size switchSizeStyle) layout.Dimensions {
	left := unit.Dp(0)
	if s.label != "" && !s.labelBefore {
		theme := ctx.Theme.Components.Switch
		left = size.trackWidth + theme.FocusSpace*2 + theme.ContentGap
	}
	return layout.Inset{Left: left}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return Text(s.description).
			Size(float32(ctx.Theme.Components.Switch.DescriptionSize)).
			Color(style.description).
			Layout(ctx, gtx)
	})
}
