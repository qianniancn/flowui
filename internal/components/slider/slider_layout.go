package slider

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
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
	return text.New(s.label).
		Size(float32(frame.ActiveTheme(ctx).Components.Slider.TextSize)).
		Weight(font.Medium).
		Color(style.label).
		Layout(ctx, gtx)
}

func (s SliderWidget) layoutOutput(ctx *frame.Context, gtx layout.Context, style sliderStyle, output string) layout.Dimensions {
	return text.New(output).
		Size(float32(frame.ActiveTheme(ctx).Components.Slider.TextSize)).
		Weight(font.Medium).
		Color(style.output).
		Layout(ctx, gtx)
}

func (s SliderWidget) layoutTrack(ctx *frame.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Slider
	thickness := max(gtx.Dp(tokens.TrackThickness), 1)
	edge := max(gtx.Dp(tokens.EdgeInset), 0)
	minimumLength := 2*edge + max(gtx.Dp(tokens.ThumbExtra), 1)
	var size image.Point
	if s.orientation == SliderVertical {
		height := max(gtx.Constraints.Max.Y, minimumLength)
		size = gtx.Constraints.Constrain(image.Pt(thickness, height))
	} else {
		width := max(gtx.Constraints.Max.X, minimumLength)
		size = gtx.Constraints.Constrain(image.Pt(width, thickness))
	}

	geometry := newSliderGeometry(size, s.axis(), edge, state.lower.Value, state.upper.Value, values.rangeMode, gtx.Dp(tokens.ThumbLength), gtx.Dp(tokens.ThumbExtra))
	drawSliderTrack(gtx, frame.ActiveTheme(ctx), style, geometry, values.rangeMode)
	s.layoutThumb(ctx, gtx, state, style, values, geometry, 0)
	if values.rangeMode {
		s.layoutThumb(ctx, gtx, state, style, values, geometry, 1)
	}
	// Float input ops are registered last so dragging owns pointer gestures even
	// when the press starts directly over a thumb's keyboard focus region.
	s.layoutFloatInputs(gtx, state, geometry, tokens.EdgeInset, values.rangeMode)
	return layout.Dimensions{Size: size}
}

func (s SliderWidget) layoutFloatInputs(gtx layout.Context, state *sliderState, geometry sliderGeometry, pointerMargin unit.Dp, rangeMode bool) {
	axis := geometry.axis
	cross := sizeCross(axis, geometry.size)
	innerSize := axis.Convert(image.Pt(geometry.inner, cross))
	offset := axis.Convert(image.Pt(geometry.edge, 0))
	layoutInput := func(value *widget.Float, inputClip image.Rectangle) {
		clipStack := clip.Rect(inputClip).Push(gtx.Ops)
		transform := op.Offset(offset).Push(gtx.Ops)
		inputGtx := gtx
		inputGtx.Constraints = layout.Exact(innerSize)
		value.Layout(inputGtx, axis, pointerMargin)
		transform.Pop()
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

func (s SliderWidget) layoutThumb(ctx *frame.Context, gtx layout.Context, state *sliderState, style sliderStyle, values sliderValues, geometry sliderGeometry, index int) {
	thumb := state.thumb(index)
	rect := geometry.thumbRects[index]
	focus := thumb.focusOpacity(ctx, gtx, gtx.Focused(&thumb.clickable))
	scale := thumb.draggingScale(ctx, gtx, state.dragging(index), frame.ActiveTheme(ctx).Components.Slider.DraggingScale)
	drawSliderThumb(gtx, frame.ActiveTheme(ctx), style, rect, s.axis(), focus, scale)

	stack := op.Offset(rect.Min).Push(gtx.Ops)
	thumbGtx := gtx
	thumbGtx.Constraints = layout.Exact(rect.Size())
	thumb.clickable.Layout(thumbGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(s.thumbSemanticLabel(values, index)).Add(gtx.Ops)
		return layout.Dimensions{Size: rect.Size()}
	})
	stack.Pop()
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
