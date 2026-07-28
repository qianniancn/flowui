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
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

func (s SliderWidget) layout(ctx *frame.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	if s.orientation == SliderVertical {
		return s.layoutVertical(ctx, gtx, state, style, values)
	}
	return s.layoutHorizontal(ctx, gtx, state, style, values)
}

func (s SliderWidget) layoutHorizontal(ctx *frame.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	output := s.outputText(values)
	if s.label == "" && output == "" {
		return s.layoutTrack(ctx, gtx, state, style, values)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.Slider.HeaderGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutHeader(ctx, gtx, style, output)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutTrack(ctx, gtx, state, style, values)
		}),
	)
}

func (s SliderWidget) layoutVertical(ctx *frame.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
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
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.Slider.VerticalGap),
	}.Layout(gtx, children...)
}

func (s SliderWidget) layoutHeader(ctx *frame.Context, gtx layout.Context, style sliderStyle, output string) layout.Dimensions {
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

func (s SliderWidget) layoutLabel(ctx *frame.Context, gtx layout.Context, style sliderStyle) layout.Dimensions {
	semantic.LabelOp(s.label).Add(gtx.Ops)
	return layoutui.LayoutResolved(ctx, gtx, style.label, text.New(s.label))
}

func (s SliderWidget) layoutOutput(ctx *frame.Context, gtx layout.Context, style sliderStyle, output string) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, style.label, text.New(output))
}

func (s SliderWidget) layoutTrack(ctx *frame.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Slider
	edge := max(gtx.Dp(tokens.EdgeInset), 0)
	minimumLength := 2*edge + max(gtx.Dp(tokens.ThumbExtra), 1)
	return layoutui.LayoutResolved(ctx, gtx, style.track, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Constrain(gtx.Constraints.Min)
		axisSize := s.axis().Convert(size)
		axisSize.X = max(axisSize.X, minimumLength)
		size = gtx.Constraints.Constrain(s.axis().Convert(axisSize))
		thumbSize := sliderThumbSize(gtx, style.thumb, image.Pt(gtx.Dp(tokens.ThumbLength+tokens.ThumbExtra), gtx.Dp(tokens.TrackThickness)))
		if s.orientation == SliderVertical {
			thumbSize = sliderThumbSize(gtx, style.thumb, image.Pt(gtx.Dp(tokens.TrackThickness), gtx.Dp(tokens.ThumbLength+tokens.ThumbExtra)))
		}
		geometry := newSliderGeometry(size, s.axis(), edge, state.lower.Value, state.upper.Value, values.rangeMode, thumbSize)
		s.layoutFill(ctx, gtx, style.fill, geometry, values.rangeMode)
		s.layoutThumb(ctx, gtx, state, style.thumb, values, geometry, 0)
		if values.rangeMode {
			s.layoutThumb(ctx, gtx, state, style.thumb, values, geometry, 1)
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

func sliderThumbSize(gtx layout.Context, style flowstyle.ResolvedStyle, fallback image.Point) image.Point {
	if style.Box == nil {
		return fallback
	}
	if style.Box.Width != nil {
		fallback.X = max(gtx.Dp(*style.Box.Width), 1)
	}
	if style.Box.Height != nil {
		fallback.Y = max(gtx.Dp(*style.Box.Height), 1)
	}
	return fallback
}

func (s SliderWidget) layoutFill(ctx *frame.Context, gtx layout.Context, style flowstyle.ResolvedStyle, geometry sliderGeometry, rangeMode bool) {
	rect := sliderFillRect(geometry, rangeMode)
	if rect.Empty() {
		return
	}
	fillGtx := gtx
	fillGtx.Constraints = layout.Exact(rect.Size())
	tokens := frame.ActiveTheme(ctx).Components.Slider
	radius := min(max(gtx.Dp(tokens.TrackRadius), 0), min(geometry.size.X, geometry.size.Y)/2)
	clipped := clip.UniformRRect(image.Rectangle{Max: geometry.size}, radius).Push(gtx.Ops)
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	layoutui.LayoutResolved(ctx, fillGtx, style, nil)
	stack.Pop()
	clipped.Pop()
}

func (s SliderWidget) layoutFloatInputs(gtx layout.Context, state *sliderState, geometry sliderGeometry, pointerMargin unit.Dp, cursor *pointer.Cursor, rangeMode bool) {
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

func (s SliderWidget) layoutThumb(ctx *frame.Context, gtx layout.Context, state *sliderState, style flowstyle.ResolvedStyle, values sliderValues, geometry sliderGeometry, index int) {
	thumb := state.thumb(index)
	rect := geometry.thumbRects[index]
	focus := thumb.focusOpacity(ctx, gtx, gtx.Focused(&thumb.clickable))
	scale := thumb.draggingScale(ctx, gtx, state.dragging(index), frame.ActiveTheme(ctx).Components.Slider.DraggingScale)

	stack := op.Offset(rect.Min).Push(gtx.Ops)
	thumbGtx := gtx
	thumbGtx.Constraints = layout.Exact(rect.Size())
	s.layoutThumbVisual(ctx, thumbGtx, styleruntime.ApplyOutlineOpacity(style, focus), scale)
	thumb.clickable.Layout(thumbGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(s.thumbSemanticLabel(values, index)).Add(gtx.Ops)
		return layout.Dimensions{Size: rect.Size()}
	})
	stack.Pop()
}

func (s SliderWidget) layoutThumbVisual(ctx *frame.Context, gtx layout.Context, style flowstyle.ResolvedStyle, scale float32) {
	outer, inner, inset, layered := sliderThumbLayers(style)
	if !layered {
		center := f32.Pt(float32(gtx.Constraints.Min.X)/2, float32(gtx.Constraints.Min.Y)/2)
		transform := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
		layoutui.LayoutResolved(ctx, gtx, style, nil)
		transform.Pop()
		return
	}
	layoutui.LayoutResolved(ctx, gtx, outer, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		pixels := min(max(gtx.Dp(inset), 0), min(size.X, size.Y)/2)
		rect := image.Rectangle{Min: image.Pt(pixels, pixels), Max: size.Sub(image.Pt(pixels, pixels))}
		if !rect.Empty() {
			center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
			transform := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
			stack := op.Offset(rect.Min).Push(gtx.Ops)
			innerGtx := gtx
			innerGtx.Constraints = layout.Exact(rect.Size())
			layoutui.LayoutResolved(ctx, innerGtx, inner, nil)
			stack.Pop()
			transform.Pop()
		}
		return layout.Dimensions{Size: size}
	}))
}

func sliderThumbLayers(style flowstyle.ResolvedStyle) (outer, inner flowstyle.ResolvedStyle, inset unit.Dp, ok bool) {
	if style.Paint == nil || style.Paint.Border == nil || style.Paint.Border.Width == nil || *style.Paint.Border.Width <= 0 || style.Paint.Border.Color == nil {
		return style, flowstyle.ResolvedStyle{}, 0, false
	}
	inset = *style.Paint.Border.Width
	outer = style
	outerPaint := *style.Paint
	outerPaint.Background = outerPaint.Border.Color
	outerPaint.Border = nil
	outerPaint.Shadows = nil
	outer.Paint = &outerPaint

	innerPaint := flowstyle.PaintStyle{Background: style.Paint.Background, Shadows: style.Paint.Shadows}
	if style.Paint.Radius != nil {
		radius := max(*style.Paint.Radius-inset, 0)
		innerPaint.Radius = &radius
	}
	if style.Paint.Radii != nil {
		radii := *style.Paint.Radii
		radii.TopLeft = max(radii.TopLeft-inset, 0)
		radii.TopRight = max(radii.TopRight-inset, 0)
		radii.BottomRight = max(radii.BottomRight-inset, 0)
		radii.BottomLeft = max(radii.BottomLeft-inset, 0)
		innerPaint.Radii = &radii
	}
	inner.Paint = &innerPaint
	return outer, inner, inset, true
}

func (s SliderWidget) thumbSemanticLabel(values sliderValues, index int) string {
	label := s.label
	if label == "" {
		label = "Slider"
	}
	if !values.rangeMode {
		return label + " " + s.format(values.lower)
	}
	if index == 0 {
		return label + " minimum " + s.format(values.lower)
	}
	return label + " maximum " + s.format(values.upper)
}
