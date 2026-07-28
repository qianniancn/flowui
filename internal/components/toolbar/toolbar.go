package toolbar

import (
	"image"
	"strconv"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
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
	customStyle flowstyle.Style
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

func (w Widget) Style(value flowstyle.Style) Widget {
	w.customStyle = value
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if !gtx.Enabled() {
		w.disabled = true
	}
	gtx.Constraints.Min = image.Point{}
	resolved := w.resolveStyle(ctx, gtx)
	group := new(frame.FocusGroup)
	dims := layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		restoreGroup := frame.PushFocusGroup(ctx, group)
		defer restoreGroup()
		if w.disabled {
			gtx = gtx.Disabled()
		}
		macro := op.Record(gtx.Ops)
		dims := w.layoutChildren(ctx, gtx)
		call := macro.Stop()
		root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.EnabledOp(!w.disabled).Add(gtx.Ops)
		label := w.alt
		if label == "" {
			label = "Toolbar"
		}
		semantic.DescriptionOp(label).Add(gtx.Ops)
		call.Add(gtx.Ops)
		root.Pop()
		return dims
	}))

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
			separator.key = "toolbar-separator:" + strconv.Itoa(index)
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
	key         string
	customStyle flowstyle.Style
}

func Separator() SeparatorWidget {
	return SeparatorWidget{}
}

func (s SeparatorWidget) Style(value flowstyle.Style) SeparatorWidget {
	s.customStyle = value
	return s
}

func (s SeparatorWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Toolbar
	defaultStyle := flowstyle.Style{}.
		Width(tokens.SeparatorWidth).
		Height(tokens.SeparatorLength).
		Background(flowstyle.SolidColor{Color: frame.ActiveTheme(ctx).Palette.SeparatorColor()})

	if s.orientation == Vertical {
		defaultStyle = flowstyle.Style{}.
			Width(tokens.SeparatorLength * 2).
			Height(tokens.SeparatorWidth).
			Background(flowstyle.SolidColor{Color: frame.ActiveTheme(ctx).Palette.SeparatorColor()})

	}
	resolved := styleruntime.ResolveStatic(ctx, flowstyle.StyleState{}, defaultStyle, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	if len(resolved.Transitions) != 0 {
		key := s.key
		if key == "" {
			key = "toolbar-separator"
		}
		key = frame.ClaimKey(ctx, stateutil.KindStyle, key)
		resolved = styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.LayoutResolved(ctx, gtx, resolved, nil)
}

func (w Widget) resolveStyle(ctx *frame.Context, gtx layout.Context) flowstyle.ResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := flowstyle.Style{}
	if w.attached {
		defaults = defaults.
			Background(flowstyle.SolidColor{Color: activeTheme.Palette.Surface}).
			TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceForeground}).
			Padding(activeTheme.Components.Toolbar.Padding).
			Radius(activeTheme.Components.Toolbar.Radius).
			Shadow(flowstyle.ShadowOverlay)
	}
	resolved := styleruntime.ResolveStatic(
		ctx,
		flowstyle.StyleState{Disabled: w.disabled},
		defaults,
		flowstyle.Style{},
		flowstyle.Style{},
		w.customStyle,
	)
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	key := frame.ClaimKey(ctx, stateutil.KindStyle, "toolbar")
	return styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
}
