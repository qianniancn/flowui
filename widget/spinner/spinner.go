package spinner

import (
	"github.com/qianniancn/flowui/style"
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	override    *theme.SpinnerOverride
	color       Color
	size        Size
	label       string
	visualCache *spinnerVisualCache
}

type Color = theme.SpinnerColor

const (
	Accent  = theme.SpinnerAccent
	Current = theme.SpinnerCurrent
	Success = theme.SpinnerSuccess
	Warning = theme.SpinnerWarning
	Danger  = theme.SpinnerDanger
)

type Size = theme.SpinnerSize

const (
	Medium     = theme.SpinnerMedium
	Small      = theme.SpinnerSmall
	Large      = theme.SpinnerLarge
	ExtraLarge = theme.SpinnerExtraLarge
)

func New() Widget {
	return Widget{visualCache: &spinnerVisualCache{}}
}

func (s Widget) Color(color Color) Widget {
	s.color = color
	return s
}

func (s Widget) Size(size Size) Widget {
	s.size = size
	return s
}

func (s Widget) Label(label string) Widget {
	s.label = label
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	resolved := s.resolveStyle(ctx, gtx)
	return s.layoutResolved(ctx, gtx, resolved)
}

func (s Widget) layoutResolved(ctx *core.Context, gtx layout.Context, resolved spinnerResolvedStyle) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		diameter := max(min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 1)
		size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
		iconSize := min(size.X, size.Y)
		if iconSize > 0 {
			offset := image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)
			stack := op.Offset(offset).Push(gtx.Ops)
			period := theme.ResolveMotionDuration(resolved.motion, resolved.metrics.SpinDuration)
			resolved.renderer(gtx, theme.SpinnerRenderRequest{
				Size:        gtx.Metric.PxToDp(iconSize),
				StrokeRatio: resolved.metrics.StrokeRatio,
				InsetRatio:  resolved.metrics.InsetRatio,
				Color:       resolved.color,
				Period:      period,
			})
			if period > 0 {
				gtx.Execute(op.InvalidateCmd{})
			}
			stack.Pop()
		}
		clipStack := clip.Rect{Max: size}.Push(gtx.Ops)
		label := s.label
		if label == "" {
			label = ctx.Text(i18n.SpinnerLoading)
		}
		semantic.DescriptionOp(label).Add(gtx.Ops)
		clipStack.Pop()
		return layout.Dimensions{Size: size}
	}))
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.SpinnerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
