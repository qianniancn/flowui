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
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

// Widget groups related controls and provides directional keyboard navigation.
type Widget struct {
	override    *theme.ToolbarOverride
	children    []core.Widget
	orientation Orientation
	attached    bool
	disabled    bool
	loopFocus   bool
	alt         string
	visualCache *toolbarVisualCache
}

func New(children ...core.Widget) Widget {
	return Widget{children: children, loopFocus: true, visualCache: &toolbarVisualCache{}}
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

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ToolbarOverride) layout.Dimensions {
	if !gtx.Enabled() {
		w.disabled = true
	}
	gtx.Constraints.Min = image.Point{}
	visual := w.visualWithCache(ctx, override)
	resolved := w.resolveStyle(ctx, gtx, visual)
	group := &core.FocusGroup{}
	dims := layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		restoreGroup := core.PushFocusGroup(ctx, group)
		defer restoreGroup()
		if w.disabled {
			gtx = gtx.Disabled()
		}
		macro := op.Record(gtx.Ops)
		dims := w.layoutChildren(ctx, gtx, visual)
		call := macro.Stop()
		root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.EnabledOp(!w.disabled).Add(gtx.Ops)
		label := w.alt
		if label == "" {
			label = ctx.Text(i18n.ToolbarDefault)
		}
		semantic.DescriptionOp(label).Add(gtx.Ops)
		call.Add(gtx.Ops)
		root.Pop()
		return dims
	}))

	w.updateKeys(ctx, gtx, group.Items)
	return dims
}

func (w Widget) layoutChildren(ctx *core.Context, gtx layout.Context, visual theme.ToolbarVisual) layout.Dimensions {
	axis := layout.Horizontal
	align := layout.Middle
	if w.orientation == Vertical {
		axis = layout.Vertical
		align = layout.Start
	}
	children := append([]core.Widget(nil), w.children...)
	for index, child := range children {
		if separator, ok := child.(SeparatorWidget); ok {
			separator.orientation = w.orientation
			separator.key = "toolbar-separator:" + strconv.Itoa(index)
			separator.visual = &visual
			children[index] = separator
			continue
		}
	}
	return layoutui.TrackedFlex(
		ctx,
		gtx,
		axis,
		visual.Metrics.Gap,
		align,
		children...,
	)
}

func (w Widget) updateKeys(ctx *core.Context, gtx layout.Context, items []core.FocusGroupItem) {
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
		core.RequestFocusVisible(ctx, items[target].Tag, true)
	}
}

func focusedIndex(gtx layout.Context, items []core.FocusGroupItem) int {
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
	override    *theme.ToolbarSeparatorOverride
	visual      *theme.ToolbarVisual
}

func Separator() SeparatorWidget {
	return SeparatorWidget{}
}

func (s SeparatorWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := s.visual
	if visual == nil {
		resolvedVisual := resolveToolbarVisual(core.ActiveTheme(ctx), theme.ToolbarRequest{Orientation: toolbarOrientation(s.orientation)})
		visual = &resolvedVisual
	}
	declaration := visual.Separator
	if s.orientation == Vertical {
		declaration = visual.VerticalSeparator
	}
	resolved := computed.ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, theme.ToolbarSeparatorOverrideStyle(s.override))
	if len(resolved.Transitions) != 0 {
		key := s.key
		if key == "" {
			key = "toolbar-separator"
		}
		key = core.ClaimKey(ctx, core.KindStyle, key)
		resolved = computed.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, nil)
}

func (w Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual theme.ToolbarVisual) style.ResolvedStyle {
	resolved := computed.ResolveStatic(
		ctx,
		style.State{Disabled: w.disabled},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		visual.Override,
	)
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	key := core.ClaimKey(ctx, core.KindStyle, "toolbar")
	return computed.ApplyTransitions(ctx, gtx, key, resolved)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ToolbarOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}

// Style merges instance declarations into the separator.
func (s SeparatorWidget) Style(values ...style.Style) SeparatorWidget {
	declaration := style.Join(values...)
	if s.override == nil {
		s.override = &theme.ToolbarSeparatorOverride{}
	}
	s.override.Declaration = style.Join(s.override.Declaration, declaration)
	return s
}
