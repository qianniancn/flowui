package layoutui

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotSplitPane = "split-pane"

const (
	splitPaneKeyboardStep     = float32(0.02)
	splitPaneKeyboardPageStep = float32(0.1)
)

type SplitPaneOrientation uint8

const (
	SplitPaneHorizontal SplitPaneOrientation = iota
	SplitPaneVertical
)

type SplitPaneWidget struct {
	key             string
	first           frame.Widget
	second          frame.Widget
	orientation     SplitPaneOrientation
	ratio           float32
	defaultRatio    float32
	hasRatio        bool
	hasDefaultRatio bool
	minFirst        unit.Dp
	minSecond       unit.Dp
	label           string
	onRatioChange   func(float32)
	disabled        bool
}

type splitPaneState struct {
	ratio       float32
	initialized bool
	hovered     bool
	dragging    bool
	pointerID   pointer.ID
	dragOffset  float32
	orientation SplitPaneOrientation
	axisReady   bool
	focus       state.FocusAnimation
}

func SplitPane(key string, first, second frame.Widget) SplitPaneWidget {
	return SplitPaneWidget{
		key:    key,
		first:  first,
		second: second,
		label:  "Resize panels",
	}
}

func (s SplitPaneWidget) Orientation(orientation SplitPaneOrientation) SplitPaneWidget {
	s.orientation = orientation
	return s
}

func (s SplitPaneWidget) Horizontal() SplitPaneWidget {
	s.orientation = SplitPaneHorizontal
	return s
}

func (s SplitPaneWidget) Vertical() SplitPaneWidget {
	s.orientation = SplitPaneVertical
	return s
}

func (s SplitPaneWidget) Ratio(ratio float32) SplitPaneWidget {
	s.ratio = splitPaneRatio(ratio)
	s.hasRatio = true
	return s
}

func (s SplitPaneWidget) DefaultRatio(ratio float32) SplitPaneWidget {
	s.defaultRatio = splitPaneRatio(ratio)
	s.hasDefaultRatio = true
	return s
}

func (s SplitPaneWidget) MinFirst(dp int) SplitPaneWidget {
	s.minFirst = unit.Dp(max(dp, 0))
	return s
}

func (s SplitPaneWidget) MinSecond(dp int) SplitPaneWidget {
	s.minSecond = unit.Dp(max(dp, 0))
	return s
}

func (s SplitPaneWidget) Label(label string) SplitPaneWidget {
	s.label = label
	return s
}

func (s SplitPaneWidget) OnRatioChange(fn func(float32)) SplitPaneWidget {
	s.onRatioChange = fn
	return s
}

func (s SplitPaneWidget) Disabled(disabled bool) SplitPaneWidget {
	s.disabled = disabled
	return s
}

func (s SplitPaneWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, s.first, s.second)
	key := frame.ClaimKey(ctx, state.KindSplitPane, s.key)
	value := frame.UseState[splitPaneState](ctx, key, stateSlotSplitPane)
	value.setOrientation(s.orientation)

	size := gtx.Constraints.Constrain(gtx.Constraints.Max)
	axis := s.axis()
	axisSize := axis.Convert(size)
	tokens := frame.ActiveTheme(ctx).Components.SplitPane
	divider := min(max(gtx.Dp(tokens.DividerWidth), 1), axisSize.X)
	available := max(axisSize.X-divider, 0)
	minFirst, maxFirst := s.bounds(gtx, available)
	first := splitPaneFirstSize(s.resolveRatio(value), available, minFirst, maxFirst)
	ratio := splitPaneSizeRatio(first, available)

	enabled := gtx.Enabled() && !s.disabled && available > 0 && minFirst < maxFirst
	if next, changed := s.update(ctx, gtx, value, enabled, axis, available, first, divider, minFirst, maxFirst); changed {
		first = splitPaneFirstSize(next, available, minFirst, maxFirst)
		next = splitPaneSizeRatio(first, available)
		if next != ratio && s.onRatioChange != nil {
			s.onRatioChange(next)
		}
		ratio = next
	}
	value.ratio = ratio
	value.initialized = true

	second := available - first
	firstSize := axis.Convert(image.Pt(first, axisSize.Y))
	secondSize := axis.Convert(image.Pt(second, axisSize.Y))
	secondOffset := axis.Convert(image.Pt(first+divider, 0))
	s.layoutChild(ctx, gtx, s.first, firstSize, image.Point{})
	s.layoutChild(ctx, gtx, s.second, secondSize, secondOffset)

	focusVisible := frame.FocusVisible(ctx, value, gtx.Focused(value))
	focusOpacity := value.focus.Opacity(gtx, focusVisible && enabled)
	drawSplitPaneDivider(ctx, gtx, axis, size, first, divider, value.hovered, value.dragging, focusOpacity, s.disabled || !gtx.Enabled())
	s.addHandle(gtx, value, enabled, axis, size, first, divider, max(gtx.Dp(tokens.HitSize), divider))
	return layout.Dimensions{Size: size}
}

func (s SplitPaneWidget) axis() layout.Axis {
	if s.orientation == SplitPaneVertical {
		return layout.Vertical
	}
	return layout.Horizontal
}

func (s SplitPaneWidget) resolveRatio(value *splitPaneState) float32 {
	if s.hasRatio {
		return s.ratio
	}
	if value.initialized {
		return value.ratio
	}
	if s.hasDefaultRatio {
		return s.defaultRatio
	}
	return 0.5
}

func (s SplitPaneWidget) bounds(gtx layout.Context, available int) (int, int) {
	minFirst := min(max(gtx.Dp(s.minFirst), 0), available)
	minSecond := min(max(gtx.Dp(s.minSecond), 0), available)
	if minFirst+minSecond <= available {
		return minFirst, available - minSecond
	}
	total := minFirst + minSecond
	if total == 0 {
		return 0, available
	}
	fixed := int(math.Round(float64(available) * float64(minFirst) / float64(total)))
	return fixed, fixed
}

func (s SplitPaneWidget) update(ctx *frame.Context, gtx layout.Context, value *splitPaneState, enabled bool, axis layout.Axis, available, first, divider, minFirst, maxFirst int) (float32, bool) {
	if !enabled {
		value.hovered = false
		value.dragging = false
		return splitPaneSizeRatio(first, available), false
	}
	next, pointerChanged := value.updatePointer(ctx, gtx, axis, available, first, divider)
	if pointerChanged {
		return splitPaneClampRatio(next, available, minFirst, maxFirst), true
	}
	next, keyboardChanged := value.updateKeyboard(gtx, s.orientation, splitPaneSizeRatio(first, available), splitPaneSizeRatio(minFirst, available), splitPaneSizeRatio(maxFirst, available))
	return next, keyboardChanged
}

func (s SplitPaneWidget) layoutChild(ctx *frame.Context, gtx layout.Context, child frame.Widget, size, offset image.Point) {
	childGtx := gtx
	childGtx.Constraints = layout.Exact(size)
	macro := op.Record(gtx.Ops)
	_, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child.Layout(ctx, childGtx)
	})
	call := macro.Stop()
	placement.PlaceOffset(offset)
	placement.ClipTo(image.Rectangle{Min: offset, Max: offset.Add(size)})
	transform := op.Offset(offset).Push(gtx.Ops)
	clipped := clip.Rect{Max: size}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipped.Pop()
	transform.Pop()
}

func (s SplitPaneWidget) addHandle(gtx layout.Context, value *splitPaneState, enabled bool, axis layout.Axis, size image.Point, first, divider, hitSize int) {
	hit := splitPaneHitRect(axis, size, first, divider, hitSize)
	if hit.Empty() {
		return
	}
	clipped := clip.Rect(hit).Push(gtx.Ops)
	if s.label != "" {
		semantic.LabelOp(s.label).Add(gtx.Ops)
	}
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if enabled {
		if axis == layout.Horizontal {
			pointer.CursorColResize.Add(gtx.Ops)
		} else {
			pointer.CursorRowResize.Add(gtx.Ops)
		}
		event.Op(gtx.Ops, value)
	}
	clipped.Pop()
}

func splitPaneHitRect(axis layout.Axis, size image.Point, first, divider, hitSize int) image.Rectangle {
	center := first + divider/2
	start := center - hitSize/2
	end := start + hitSize
	if axis == layout.Horizontal {
		return image.Rect(max(start, 0), 0, min(end, size.X), size.Y)
	}
	return image.Rect(0, max(start, 0), size.X, min(end, size.Y))
}

func (s *splitPaneState) setOrientation(orientation SplitPaneOrientation) {
	if s.axisReady && s.orientation != orientation {
		s.hovered = false
		s.dragging = false
	}
	s.orientation = orientation
	s.axisReady = true
}

func (s *splitPaneState) updatePointer(ctx *frame.Context, gtx layout.Context, axis layout.Axis, available, first, divider int) (float32, bool) {
	next := splitPaneSizeRatio(first, available)
	changed := false
	for {
		e, ok := gtx.Event(pointer.Filter{Target: s, Kinds: pointer.Enter | pointer.Leave | pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			return next, changed
		}
		event, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		switch event.Kind {
		case pointer.Enter:
			s.hovered = true
		case pointer.Leave:
			if !s.dragging {
				s.hovered = false
			}
		case pointer.Press:
			if event.Source != pointer.Touch && event.Buttons&pointer.ButtonPrimary == 0 {
				continue
			}
			s.dragging = true
			s.hovered = true
			s.pointerID = event.PointerID
			s.dragOffset = splitPanePointerPosition(axis, event.Position) - (float32(first) + float32(divider)/2)
			frame.RequestFocusVisible(ctx, s, false)
			gtx.Execute(pointer.GrabCmd{Tag: s, ID: event.PointerID})
		case pointer.Drag:
			if !s.dragging || event.PointerID != s.pointerID {
				continue
			}
			position := splitPanePointerPosition(axis, event.Position) - s.dragOffset - float32(divider)/2
			next = position / float32(available)
			changed = true
		case pointer.Release:
			if event.PointerID == s.pointerID {
				s.dragging = false
			}
		case pointer.Cancel:
			if event.PointerID == s.pointerID {
				s.dragging = false
				s.hovered = false
			}
		}
	}
}

func (s *splitPaneState) updateKeyboard(gtx layout.Context, orientation SplitPaneOrientation, current, minimum, maximum float32) (float32, bool) {
	filters := []event.Filter{
		key.Filter{Focus: s, Name: key.NameHome},
		key.Filter{Focus: s, Name: key.NameEnd},
		key.Filter{Focus: s, Name: key.NamePageDown},
		key.Filter{Focus: s, Name: key.NamePageUp},
	}
	if orientation == SplitPaneVertical {
		filters = append(filters,
			key.Filter{Focus: s, Name: key.NameUpArrow},
			key.Filter{Focus: s, Name: key.NameDownArrow},
		)
	} else {
		filters = append(filters,
			key.Filter{Focus: s, Name: key.NameLeftArrow},
			key.Filter{Focus: s, Name: key.NameRightArrow},
		)
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return current, false
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		next := current
		switch event.Name {
		case key.NameLeftArrow, key.NameUpArrow:
			next -= splitPaneKeyboardStep
		case key.NameRightArrow, key.NameDownArrow:
			next += splitPaneKeyboardStep
		case key.NamePageDown:
			next -= splitPaneKeyboardPageStep
		case key.NamePageUp:
			next += splitPaneKeyboardPageStep
		case key.NameHome:
			next = minimum
		case key.NameEnd:
			next = maximum
		}
		next = min(max(next, minimum), maximum)
		return next, next != current
	}
}

func splitPanePointerPosition(axis layout.Axis, point f32.Point) float32 {
	if axis == layout.Vertical {
		return float32(point.Y)
	}
	return float32(point.X)
}

func splitPaneFirstSize(ratio float32, available, minimum, maximum int) int {
	first := int(math.Round(float64(splitPaneRatio(ratio) * float32(available))))
	return min(max(first, minimum), maximum)
}

func splitPaneClampRatio(ratio float32, available, minimum, maximum int) float32 {
	return splitPaneSizeRatio(splitPaneFirstSize(ratio, available, minimum, maximum), available)
}

func splitPaneSizeRatio(size, available int) float32 {
	if available <= 0 {
		return 0.5
	}
	return float32(size) / float32(available)
}

func splitPaneRatio(ratio float32) float32 {
	if math.IsNaN(float64(ratio)) || math.IsInf(float64(ratio), 0) {
		return 0.5
	}
	return min(max(ratio, 0), 1)
}

func drawSplitPaneDivider(ctx *frame.Context, gtx layout.Context, axis layout.Axis, size image.Point, first, divider int, hovered, dragging bool, focus float32, disabled bool) {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.SplitPane
	base := activeTheme.Palette.SeparatorColor()
	if disabled {
		base = activeTheme.DisabledColor(base)
	}
	dividerRect := splitPaneDividerRect(axis, size, first, divider)
	paint.FillShape(gtx.Ops, base, clip.Rect(dividerRect).Op())

	active := focus
	if hovered || dragging {
		active = 1
	}
	if active <= 0 || disabled {
		return
	}
	col := activeTheme.Palette.Accent
	col.A = byte(float32(col.A)*min(max(active, 0), 1) + 0.5)
	activeWidth := max(gtx.Dp(tokens.ActiveWidth), divider)
	handleLength := max(gtx.Dp(tokens.HandleLength), activeWidth)
	handle := splitPaneHandleRect(axis, size, first, divider, activeWidth, handleLength).Intersect(image.Rectangle{Max: size})
	if handle.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(handle, max(activeWidth/2, 1)).Op(gtx.Ops))
}

func splitPaneDividerRect(axis layout.Axis, size image.Point, first, divider int) image.Rectangle {
	if axis == layout.Vertical {
		return image.Rect(0, first, size.X, first+divider)
	}
	return image.Rect(first, 0, first+divider, size.Y)
}

func splitPaneHandleRect(axis layout.Axis, size image.Point, first, divider, width, length int) image.Rectangle {
	mainCenter := first + divider/2
	crossSize := axis.Convert(size).Y
	length = min(length, crossSize)
	crossStart := (crossSize - length) / 2
	mainStart := mainCenter - width/2
	return axisRect(axis, mainStart, crossStart, width, length)
}

func axisRect(axis layout.Axis, main, cross, mainSize, crossSize int) image.Rectangle {
	minPoint := axis.Convert(image.Pt(main, cross))
	return image.Rectangle{Min: minPoint, Max: minPoint.Add(axis.Convert(image.Pt(mainSize, crossSize)))}
}
