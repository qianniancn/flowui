package spinner

import (
	"image"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const spinnerPeriod = 750 * time.Millisecond

type SpinnerWidget struct {
	color SpinnerColor
	size  SpinnerSize
	label string
}

type SpinnerColor uint8

const (
	SpinnerAccent SpinnerColor = iota
	SpinnerCurrent
	SpinnerSuccess
	SpinnerWarning
	SpinnerDanger
)

type SpinnerSize uint8

const (
	SpinnerMedium SpinnerSize = iota
	SpinnerSmall
	SpinnerLarge
	SpinnerExtraLarge
)

func Spinner() SpinnerWidget {
	return SpinnerWidget{}
}

func (s SpinnerWidget) Color(color SpinnerColor) SpinnerWidget {
	s.color = color
	return s
}

func (s SpinnerWidget) Size(size SpinnerSize) SpinnerWidget {
	s.size = size
	return s
}

func (s SpinnerWidget) Label(label string) SpinnerWidget {
	s.label = label
	return s
}

func (s SpinnerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	style := spinnerStyleFor(activeTheme, s.color)
	sizeStyle := spinnerSizeStyleFor(activeTheme, s.size)
	diameter := max(gtx.Dp(sizeStyle.diameter), 1)
	size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
	iconSize := min(diameter, size.X, size.Y)

	macro := op.Record(gtx.Ops)
	if iconSize > 0 {
		offset := image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)
		stack := op.Offset(offset).Push(gtx.Ops)
		period := theme.ResolveMotionDuration(activeTheme.Motion, spinnerPeriod)
		drawSpinner(gtx, iconSize, sizeStyle.strokeRatio, sizeStyle.insetRatio, style.color, period)
		stack.Pop()
	}
	call := macro.Stop()

	clipStack := clip.Rect{Max: size}.Push(gtx.Ops)
	label := s.label
	if label == "" {
		label = "Loading"
	}
	semantic.DescriptionOp(label).Add(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return layout.Dimensions{Size: size}
}
