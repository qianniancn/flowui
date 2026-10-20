package switchcontrol

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/text"
)

func (s Widget) layoutContent(ctx *core.Context, gtx layout.Context, style switchStyle, checked bool) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return s.layoutRow(ctx, gtx, style, checked)
	}
	if s.description == "" {
		return row(gtx)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(style.visual.Metrics.DescriptionGap),
	}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutDescription(ctx, gtx, style)
		}),
	)
}

func (s Widget) layoutRow(ctx *core.Context, gtx layout.Context, style switchStyle, checked bool) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return s.layoutControl(ctx, gtx, style, checked)
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
		Gap:       gtx.Dp(style.visual.Metrics.ContentGap),
	}.Layout(gtx, first, second)
}

func (s Widget) layoutControl(ctx *core.Context, gtx layout.Context, style switchStyle, checked bool) layout.Dimensions {
	result := switchRects(gtx, style.visual.Metrics.FocusSpace, style)
	if result.trackRect.Empty() {
		return result.dims
	}
	layoutSwitchLayer(ctx, gtx, result.trackRect, style.trackOff, 1-style.selected, style.focus, nil)
	layoutSwitchLayer(ctx, gtx, result.trackRect, style.trackOn, style.selected, style.focus, nil)

	var child core.Widget
	if s.thumb != nil && !result.thumbRect.Empty() {
		child = s.thumb(checked)
	}
	activeThumb := style.thumbOff
	if checked {
		activeThumb = style.thumbOn
	}
	child = s.styleThumbContent(activeThumb, child)
	var offChild, onChild core.Widget
	if checked {
		onChild = child
	} else {
		offChild = child
	}
	layoutSwitchLayer(ctx, gtx, result.thumbRect, style.thumbOff, 1-style.selected, 1, centeredSwitchChild(offChild))
	layoutSwitchLayer(ctx, gtx, result.thumbRect, style.thumbOn, style.selected, 1, centeredSwitchChild(onChild))
	return result.dims
}

func layoutSwitchLayer(ctx *core.Context, gtx layout.Context, rect image.Rectangle, resolved style.ResolvedStyle, opacity, outlineOpacity float32, child core.Widget) {
	if rect.Empty() {
		return
	}
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.Resolved(ctx, layerGtx, computed.ApplyOutlineOpacity(resolved, outlineOpacity), child)
	fade.Pop()
	stack.Pop()
}

func centeredSwitchChild(child core.Widget) core.Widget {
	if child == nil {
		return nil
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		})
	})
}

func (s Widget) styleThumbContent(resolved style.ResolvedStyle, child core.Widget) core.Widget {
	if child == nil {
		return nil
	}
	value, ok := child.(text.Widget)
	if !ok {
		return child
	}
	return text.WithDefaults(value, style.TextDeclaration(resolved.Text))
}

func (s Widget) layoutLabel(ctx *core.Context, gtx layout.Context, style switchStyle) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, style.label, text.New(s.label))
}

func (s Widget) layoutDescription(ctx *core.Context, gtx layout.Context, resolved switchStyle) layout.Dimensions {
	left := unit.Dp(0)
	if s.label != "" && !s.labelBefore {
		trackWidth, _ := switchTrackDpSize(resolved)
		left = trackWidth + resolved.visual.Metrics.FocusSpace*2 + resolved.visual.Metrics.ContentGap
	}
	return layout.Inset{Left: left}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layoutui.Resolved(ctx, gtx, resolved.description,
			(description.New(s.description).
				For(s.key).
				Disabled(s.disabled)).Style(style.TextDeclaration(resolved.description.Text)),
		)
	})
}
