package toolbar

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

// Widget groups related controls and provides directional keyboard navigation.
type Widget struct {
	children    []frame.Widget
	orientation Orientation
	attached    bool
	disabled    bool
	loopFocus   bool
	alt         string
}

func New(children ...frame.Widget) Widget {
	return Widget{children: children, loopFocus: true}
}

func (w Widget) Orientation(orientation Orientation) Widget {
	w.orientation = orientation
	return w
}

func (w Widget) Attached(attached bool) Widget {
	w.attached = attached
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) LoopFocus(loop bool) Widget {
	w.loopFocus = loop
	return w
}

func (w Widget) Alt(alt string) Widget {
	w.alt = alt
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if !gtx.Enabled() {
		w.disabled = true
	}
	gtx.Constraints.Min = image.Point{}
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Toolbar
	group := new(frame.FocusGroup)

	macro := op.Record(gtx.Ops)
	dims := func() layout.Dimensions {
		restoreGroup := frame.PushFocusGroup(ctx, group)
		defer restoreGroup()
		if w.disabled {
			gtx = gtx.Disabled()
		}
		content := func(gtx layout.Context) layout.Dimensions {
			return w.layoutChildren(ctx, gtx)
		}
		if !w.attached {
			return content(gtx)
		}
		restoreColors := frame.PushColors(
			ctx,
			theme.ColorOr(activeTheme.Palette.SurfaceForeground, activeTheme.Palette.Foreground),
			activeTheme.Palette.Surface,
		)
		defer restoreColors()
		return layout.UniformInset(tokens.Padding).Layout(gtx, content)
	}()
	call := macro.Stop()

	if w.attached && dims.Size.X > 0 && dims.Size.Y > 0 {
		rect := image.Rectangle{Max: dims.Size}
		radius := min(max(gtx.Dp(tokens.Radius), 0), min(dims.Size.X, dims.Size.Y)/2)
		render.DrawShadow(
			gtx,
			rect,
			render.RoundedShadowCorners(tokens.Radius, tokens.Radius, tokens.Radius, tokens.Radius),
			render.PopupShadow(activeTheme.Palette.OverlayShadow),
		)
		paint.FillShape(gtx.Ops, activeTheme.Palette.Surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	semantic.EnabledOp(!w.disabled).Add(gtx.Ops)
	label := w.alt
	if label == "" {
		label = "Toolbar"
	}
	semantic.DescriptionOp(label).Add(gtx.Ops)
	call.Add(gtx.Ops)
	root.Pop()

	w.updateKeys(ctx, gtx, group.Items)
	return dims
}

func (w Widget) layoutChildren(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	axis := layout.Horizontal
	align := layout.Middle
	if w.orientation == Vertical {
		axis = layout.Vertical
		align = layout.Start
	}
	children := append([]frame.Widget(nil), w.children...)
	for index, child := range children {
		if separator, ok := child.(SeparatorWidget); ok {
			separator.orientation = w.orientation
			children[index] = separator
		}
	}
	return layoutui.LayoutTrackedFlex(
		ctx,
		gtx,
		axis,
		frame.ActiveTheme(ctx).Components.Toolbar.Gap,
		align,
		children...,
	)
}

func (w Widget) updateKeys(ctx *frame.Context, gtx layout.Context, items []frame.FocusGroupItem) {
	if len(items) == 0 {
		return
	}
	previous, next := key.NameLeftArrow, key.NameRightArrow
	if w.orientation == Vertical {
		previous, next = key.NameUpArrow, key.NameDownArrow
	}
	filters := make([]event.Filter, 0, len(items)*4)
	for _, item := range items {
		filters = append(filters,
			key.Filter{Focus: item.Tag, Name: previous},
			key.Filter{Focus: item.Tag, Name: next},
			key.Filter{Focus: item.Tag, Name: key.NameHome},
			key.Filter{Focus: item.Tag, Name: key.NameEnd},
		)
	}
	for {
		value, ok := gtx.Event(filters...)
		if !ok {
			return
		}
		e, ok := value.(key.Event)
		if !ok || e.State != key.Press {
			continue
		}
		current := focusedIndex(gtx, items)
		target := current
		switch e.Name {
		case previous:
			target = moveIndex(current, -1, len(items), w.loopFocus)
		case next:
			target = moveIndex(current, 1, len(items), w.loopFocus)
		case key.NameHome:
			target = 0
		case key.NameEnd:
			target = len(items) - 1
		}
		if target < 0 || target >= len(items) || target == current {
			continue
		}
		if items[target].Prepare != nil {
			items[target].Prepare(true)
		}
		frame.RequestFocusVisible(ctx, items[target].Tag, true)
	}
}

func focusedIndex(gtx layout.Context, items []frame.FocusGroupItem) int {
	for index, item := range items {
		if gtx.Focused(item.Tag) {
			return index
		}
	}
	return -1
}

func moveIndex(current, delta, count int, loop bool) int {
	if count == 0 {
		return -1
	}
	if current < 0 {
		if delta < 0 {
			return count - 1
		}
		return 0
	}
	next := current + delta
	if loop {
		return (next + count) % count
	}
	return min(max(next, 0), count-1)
}

type SeparatorWidget struct {
	orientation Orientation
}

func Separator() SeparatorWidget {
	return SeparatorWidget{}
}

func (s SeparatorWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Toolbar
	length := max(gtx.Dp(tokens.SeparatorLength), 1)
	thickness := max(gtx.Dp(tokens.SeparatorWidth), 1)
	size := image.Pt(thickness, length)
	line := image.Rectangle{Max: size}
	if s.orientation == Vertical {
		size = image.Pt(length*2, thickness)
		line = image.Rect(length/2, 0, length+length/2, thickness)
	}
	size = gtx.Constraints.Constrain(size)
	paint.FillShape(gtx.Ops, frame.ActiveTheme(ctx).Palette.SeparatorColor(), clip.Rect(line).Op())
	return layout.Dimensions{Size: size}
}
