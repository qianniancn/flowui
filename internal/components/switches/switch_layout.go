package switches

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/description"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
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
			return s.layoutDescription(ctx, gtx, style, size)
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
	result := switchRects(gtx, frame.ActiveTheme(ctx), style, size)
	if result.trackRect.Empty() {
		return result.dims
	}
	layoutSwitchLayer(ctx, gtx, result.trackRect, style.trackOff, 1-style.selected, style.focus, nil)
	layoutSwitchLayer(ctx, gtx, result.trackRect, style.trackOn, style.selected, style.focus, nil)

	var child frame.Widget
	if s.thumb != nil && !result.thumbRect.Empty() {
		child = s.thumb(checked)
	}
	activeThumb := style.thumbOff
	if checked {
		activeThumb = style.thumbOn
	}
	child = s.styleThumbContent(activeThumb, child)
	var offChild, onChild frame.Widget
	if checked {
		onChild = child
	} else {
		offChild = child
	}
	layoutSwitchLayer(ctx, gtx, result.thumbRect, style.thumbOff, 1-style.selected, 1, centeredSwitchChild(offChild))
	layoutSwitchLayer(ctx, gtx, result.thumbRect, style.thumbOn, style.selected, 1, centeredSwitchChild(onChild))
	return result.dims
}

func layoutSwitchLayer(ctx *frame.Context, gtx layout.Context, rect image.Rectangle, style flowstyle.ResolvedStyle, opacity, outlineOpacity float32, child frame.Widget) {
	if rect.Empty() {
		return
	}
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.LayoutResolved(ctx, layerGtx, styleruntime.ApplyOutlineOpacity(style, outlineOpacity), child)
	fade.Pop()
	stack.Pop()
}

func centeredSwitchChild(child frame.Widget) frame.Widget {
	if child == nil {
		return nil
	}
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		})
	})
}

func (s SwitchWidget) styleThumbContent(style flowstyle.ResolvedStyle, child frame.Widget) frame.Widget {
	if child == nil {
		return nil
	}
	value, ok := child.(text.Widget)
	if !ok {
		return child
	}
	return text.WithDefaults(value, flowstyle.TextDeclaration(style.Text))
}

func (s SwitchWidget) layoutLabel(ctx *frame.Context, gtx layout.Context, style switchStyle) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, style.label, text.New(s.label))
}

func (s SwitchWidget) layoutDescription(ctx *frame.Context, gtx layout.Context, style switchStyle, size switchSizeStyle) layout.Dimensions {
	left := unit.Dp(0)
	if s.label != "" && !s.labelBefore {
		theme := frame.ActiveTheme(ctx).Components.Switch
		trackWidth, _ := switchTrackDpSize(style, size)
		left = trackWidth + theme.FocusSpace*2 + theme.ContentGap
	}
	return layout.Inset{Left: left}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layoutui.LayoutResolved(ctx, gtx, style.description,
			description.Description(s.description).
				For(s.key).
				Disabled(s.disabled).
				Style(flowstyle.TextDeclaration(style.description.Text)),
		)
	})
}
