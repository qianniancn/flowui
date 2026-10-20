package slider

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

func (s Widget) layout(ctx *core.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	if s.orientation == Vertical {
		return s.layoutVertical(ctx, gtx, state, style, values)
	}
	return s.layoutHorizontal(ctx, gtx, state, style, values)
}

func (s Widget) layoutHorizontal(ctx *core.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	output := s.outputText(values)
	if s.label == "" && output == "" {
		return s.layoutTrack(ctx, gtx, state, style, values)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(style.visual.Metrics.HeaderGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutHeader(ctx, gtx, style, output)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutTrack(ctx, gtx, state, style, values)
		}),
	)
}

func (s Widget) layoutVertical(ctx *core.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	output := s.outputText(values)
	children := make([]layout.FlexChild, 0, 3)
	if output != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return s.layoutOutput(ctx, gtx, style, output)
			})
		}))
	}
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return s.layoutTrack(ctx, gtx, state, style, values)
		})
	}))
	if s.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return s.layoutLabel(ctx, gtx, style)
			})
		}))
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(style.visual.Metrics.VerticalGap),
	}.Layout(gtx, children...)
}

func (s Widget) layoutHeader(ctx *core.Context, gtx layout.Context, style sliderStyle, output string) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if s.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutLabel(ctx, gtx, style)
		}))
	}
	if output != "" {
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
		}))
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutOutput(ctx, gtx, style, output)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (s Widget) layoutLabel(ctx *core.Context, gtx layout.Context, style sliderStyle) layout.Dimensions {
	semantic.LabelOp(s.label).Add(gtx.Ops)
	return layoutui.Resolved(ctx, gtx, style.label, text.New(s.label))
}

func (s Widget) layoutOutput(ctx *core.Context, gtx layout.Context, style sliderStyle, output string) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, style.label, text.New(output))
}

func (s Widget) layoutTrack(ctx *core.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	tokens := style.visual.Metrics
	edge := max(gtx.Dp(tokens.EdgeInset), 0)
	minimumLength := 2*edge + max(gtx.Dp(tokens.ThumbExtra), 1)
	return layoutui.Resolved(ctx, gtx, style.track, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Constrain(gtx.Constraints.Min)
		axisSize := s.axis().Convert(size)
		axisSize.X = max(axisSize.X, minimumLength)
		size = gtx.Constraints.Constrain(s.axis().Convert(axisSize))
		thumbSize := sliderThumbSize(gtx, style.thumb)
		geometry := newSliderGeometry(size, s.axis(), edge, state.lower.Value, state.upper.Value, values.rangeMode, thumbSize)
		s.layoutFill(ctx, gtx, style.fill, geometry, values.rangeMode, tokens.TrackRadius)
		s.layoutThumb(ctx, gtx, state, style, values, geometry, 0)
		if values.rangeMode {
			s.layoutThumb(ctx, gtx, state, style, values, geometry, 1)
		}
		// Float input ops are registered last so dragging owns pointer gestures even
		// when the press starts directly over a thumb's keyboard focus region.
		var cursor *pointer.Cursor
		if style.track.Box != nil {
			cursor = style.track.Box.Cursor
		}
		s.layoutFloatInputs(gtx, state, geometry, tokens.EdgeInset, cursor, values.rangeMode)
		return layout.Dimensions{Size: size}
	}))
}

func sliderThumbSize(gtx layout.Context, resolved style.ResolvedStyle) image.Point {
	if resolved.Box == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	width := resolved.Box.Width
	if width == nil {
		width = resolved.Box.MinWidth
	}
	height := resolved.Box.Height
	if height == nil {
		height = resolved.Box.MinHeight
	}
	if width == nil || height == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return image.Pt(max(gtx.Dp(*width), 1), max(gtx.Dp(*height), 1))
}

func (s Widget) layoutFill(ctx *core.Context, gtx layout.Context, resolved style.ResolvedStyle, geometry sliderGeometry, rangeMode bool, trackRadius unit.Dp) {
	rect := sliderFillRect(geometry, rangeMode)
	if rect.Empty() {
		return
	}
	fillGtx := gtx
	fillGtx.Constraints = layout.Exact(rect.Size())
	radius := min(max(gtx.Dp(trackRadius), 0), min(geometry.size.X, geometry.size.Y)/2)
	clipped := clip.UniformRRect(image.Rectangle{Max: geometry.size}, radius).Push(gtx.Ops)
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	layoutui.Resolved(ctx, fillGtx, resolved, nil)
	stack.Pop()
	clipped.Pop()
}

func (s Widget) layoutFloatInputs(gtx layout.Context, state *sliderState, geometry sliderGeometry, pointerMargin unit.Dp, cursor *pointer.Cursor, rangeMode bool) {
	axis := geometry.axis
	cross := sizeCross(axis, geometry.size)
	hitCross := max(cross, gtx.Dp(unit.Dp(24)))
	crossInset := (hitCross - cross) / 2
	innerSize := axis.Convert(image.Pt(geometry.inner, hitCross))
	offset := axis.Convert(image.Pt(geometry.edge, -crossInset))
	expandCross := func(rect image.Rectangle) image.Rectangle {
		if axis == layout.Vertical {
			rect.Min.X -= crossInset
			rect.Max.X += hitCross - cross - crossInset
		} else {
			rect.Min.Y -= crossInset
			rect.Max.Y += hitCross - cross - crossInset
		}
		return rect
	}
	layoutInput := func(value *widget.Float, inputClip image.Rectangle) {
		clipStack := clip.Rect(expandCross(inputClip)).Push(gtx.Ops)
		transform := op.Offset(offset).Push(gtx.Ops)
		inputGtx := gtx
		inputGtx.Constraints = layout.Exact(innerSize)
		value.Layout(inputGtx, axis, pointerMargin)
		transform.Pop()
		if cursor != nil {
			cursor.Add(gtx.Ops)
		}
		clipStack.Pop()
	}
	if !rangeMode {
		layoutInput(&state.lower, image.Rectangle{Max: geometry.size})
		return
	}
	if axis == layout.Vertical {
		split := (geometry.centers[0].Y + geometry.centers[1].Y) / 2
		layoutInput(&state.lower, image.Rect(0, split, geometry.size.X, geometry.size.Y))
		layoutInput(&state.upper, image.Rect(0, 0, geometry.size.X, split))
		return
	}
	split := (geometry.centers[0].X + geometry.centers[1].X) / 2
	layoutInput(&state.lower, image.Rect(0, 0, split, geometry.size.Y))
	layoutInput(&state.upper, image.Rect(split, 0, geometry.size.X, geometry.size.Y))
}

func (s Widget) layoutThumb(ctx *core.Context, gtx layout.Context, state *sliderState, component sliderStyle, values sliderValues, geometry sliderGeometry, index int) {
	thumb := state.thumb(index)
	rect := geometry.thumbRects[index]
	focus := thumb.focusOpacity(ctx, gtx, gtx.Focused(&thumb.clickable), component.visual.FocusDuration, component.visual.Motion)
	scale := thumb.draggingScale(ctx, gtx, state.dragging(index), component.visual.Metrics.DraggingScale, component.visual.DraggingDuration, component.visual.Motion)

	stack := op.Offset(rect.Min).Push(gtx.Ops)
	thumbGtx := gtx
	thumbGtx.Constraints = layout.Exact(rect.Size())
	s.layoutThumbVisual(ctx, thumbGtx, computed.ApplyOutlineOpacity(component.thumb, focus), scale)
	thumb.clickable.Layout(thumbGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(s.thumbSemanticLabel(ctx, values, index)).Add(gtx.Ops)
		return layout.Dimensions{Size: rect.Size()}
	})
	stack.Pop()
}

func (s Widget) layoutThumbVisual(ctx *core.Context, gtx layout.Context, resolved style.ResolvedStyle, scale float32) {
	outer, inner, inset, layered := sliderThumbLayers(resolved)
	if !layered {
		center := f32.Pt(float32(gtx.Constraints.Min.X)/2, float32(gtx.Constraints.Min.Y)/2)
		transform := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
		layoutui.Resolved(ctx, gtx, resolved, nil)
		transform.Pop()
		return
	}
	layoutui.Resolved(ctx, gtx, outer, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		pixels := min(max(gtx.Dp(inset), 0), min(size.X, size.Y)/2)
		rect := image.Rectangle{Min: image.Pt(pixels, pixels), Max: size.Sub(image.Pt(pixels, pixels))}
		if !rect.Empty() {
			center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
			transform := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
			stack := op.Offset(rect.Min).Push(gtx.Ops)
			innerGtx := gtx
			innerGtx.Constraints = layout.Exact(rect.Size())
			layoutui.Resolved(ctx, innerGtx, inner, nil)
			stack.Pop()
			transform.Pop()
		}
		return layout.Dimensions{Size: size}
	}))
}

func sliderThumbLayers(resolved style.ResolvedStyle) (outer, inner style.ResolvedStyle, inset unit.Dp, ok bool) {
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width <= 0 || resolved.Paint.Border.Color == nil {
		return resolved, style.ResolvedStyle{}, 0, false
	}
	inset = *resolved.Paint.Border.Width
	outer = resolved
	outerPaint := *resolved.Paint
	outerPaint.Background = outerPaint.Border.Color
	outerPaint.Border = nil
	outerPaint.Shadows = nil
	outer.Paint = &outerPaint

	innerPaint := style.PaintStyle{Background: resolved.Paint.Background, Shadows: resolved.Paint.Shadows}
	if resolved.Paint.Radius != nil {
		radius := max(*resolved.Paint.Radius-inset, 0)
		innerPaint.Radius = &radius
	}
	if resolved.Paint.Radii != nil {
		radii := *resolved.Paint.Radii
		radii.TopLeft = max(radii.TopLeft-inset, 0)
		radii.TopRight = max(radii.TopRight-inset, 0)
		radii.BottomRight = max(radii.BottomRight-inset, 0)
		radii.BottomLeft = max(radii.BottomLeft-inset, 0)
		innerPaint.Radii = &radii
	}
	inner.Paint = &innerPaint
	return outer, inner, inset, true
}

func (s Widget) thumbSemanticLabel(ctx *core.Context, values sliderValues, index int) string {
	label := s.label
	if label == "" {
		label = ctx.Text(i18n.SliderDefault)
	}
	if !values.rangeMode {
		return ctx.Textf(i18n.SliderValue, label, s.format(values.lower))
	}
	if index == 0 {
		return ctx.Textf(i18n.SliderMinimum, label, s.format(values.lower))
	}
	return ctx.Textf(i18n.SliderMaximum, label, s.format(values.upper))
}
