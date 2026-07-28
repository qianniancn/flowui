package spinner

import (
	"image"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

const spinnerPeriod = 750 * time.Millisecond

type SpinnerWidget struct {
	color       SpinnerColor
	size        SpinnerSize
	label       string
	customStyle flowstyle.Style
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

func (s SpinnerWidget) Style(value flowstyle.Style) SpinnerWidget {
	s.customStyle = value
	return s
}

func (s SpinnerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	resolved := s.resolveStyle(ctx, gtx)
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		diameter := max(min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 1)
		size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
		iconSize := min(size.X, size.Y)
		if iconSize > 0 {
			offset := image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)
			stack := op.Offset(offset).Push(gtx.Ops)
			period := theme.ResolveMotionDuration(activeTheme.Motion, spinnerPeriod)
			drawSpinner(gtx, iconSize, resolved.size.strokeRatio, resolved.size.insetRatio, resolved.visual.color, period)
			stack.Pop()
		}
		clipStack := clip.Rect{Max: size}.Push(gtx.Ops)
		label := s.label
		if label == "" {
			label = "Loading"
		}
		semantic.DescriptionOp(label).Add(gtx.Ops)
		clipStack.Pop()
		return layout.Dimensions{Size: size}
	}))
}
