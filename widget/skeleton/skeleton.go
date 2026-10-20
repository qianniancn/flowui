// Package skeleton provides passive placeholders while content is loading.
package skeleton

import (
	"github.com/qianniancn/flowui/style"
	"image"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	width, height    unit.Dp
	radius           unit.Dp
	circle, animated bool
	hasHeight        bool
	hasRadius        bool
	override         *theme.SkeletonOverride
}

// New creates a full-width placeholder with a gentle pulse. Its depth and
// corner radius come from the provider.
func New() Widget                           { return Widget{animated: true} }
func (s Widget) Width(value unit.Dp) Widget { s.width = max(0, value); return s }

// Height sets the placeholder depth; without it the provider's default applies.
func (s Widget) Height(value unit.Dp) Widget { s.height = max(0, value); s.hasHeight = true; return s }

// Radius sets the corner radius; without it the provider's default applies.
func (s Widget) Radius(value unit.Dp) Widget { s.radius = max(0, value); s.hasRadius = true; return s }
func (s Widget) Circle() Widget              { s.circle = true; return s }
func (s Widget) Animated(value bool) Widget  { s.animated = value; return s }

// Style merges instance declarations into the placeholder.
func (s Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if s.override == nil {
		s.override = &theme.SkeletonOverride{}
	}
	s.override.Declaration = style.Join(s.override.Declaration, declaration)
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	visual := theme.ResolveVisual(active, theme.SkeletonRequest{Circle: s.circle, Override: s.override},
		theme.ComponentVisualProvider.Skeleton, theme.ValidateSkeletonVisual)

	// A placeholder keeps its requested shape even in a stretching parent.
	gtx.Constraints.Min = image.Point{}
	width := gtx.Constraints.Max.X
	if s.width > 0 {
		width = gtx.Dp(s.width)
	}
	height := s.height
	if !s.hasHeight {
		height = visual.Metrics.Height
	}
	heightPx := gtx.Dp(height)
	if s.circle {
		heightPx = min(heightPx, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
		width = heightPx
	}
	size := gtx.Constraints.Constrain(image.Pt(width, heightPx))
	radius := s.radius
	if !s.hasRadius {
		radius = visual.Metrics.Radius
	}
	radiusPx := min(gtx.Dp(radius), min(size.X, size.Y)/2)
	if s.circle {
		radiusPx = min(size.X, size.Y) / 2
	}
	fill := visual.Colors.Base
	period := theme.ResolveMotionDuration(visual.Motion, visual.Metrics.Period)
	if s.animated && period > 0 {
		phase := float64(gtx.Now.UnixNano()%int64(period)) / float64(period)
		fill = animation.LerpColor(visual.Colors.Base, visual.Colors.Highlight, float32((1-math.Cos(phase*2*math.Pi))/2))
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(time.Second / 60)})
	}
	paint.FillShape(gtx.Ops, fill, clip.UniformRRect(image.Rectangle{Max: size}, radiusPx).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

var _ core.Widget = Widget{}
