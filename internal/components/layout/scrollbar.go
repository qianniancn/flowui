package layoutui

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

const (
	stateSlotScrollbar    = "scrollbar"
	stateSlotScrollbarBar = "bar"
)

type ScrollbarWidget struct {
	key           string
	child         frame.Widget
	axis          layout.Axis
	align         layout.Alignment
	disabled      bool
	stickToEnd    bool
	scrollAnyAxis bool
	overlay       bool
	customStyle   flowstyle.Style
}

type scrollbarState struct {
	list layout.List
	bar  widget.Scrollbar
}

func Scrollbar(key string, child frame.Widget) ScrollbarWidget {
	return ScrollbarWidget{key: key, child: child, axis: layout.Vertical}
}

func (s ScrollbarWidget) Vertical() ScrollbarWidget {
	s.axis = layout.Vertical
	return s
}

func (s ScrollbarWidget) Horizontal() ScrollbarWidget {
	s.axis = layout.Horizontal
	return s
}

func (s ScrollbarWidget) Disabled(disabled bool) ScrollbarWidget {
	s.disabled = disabled
	return s
}

func (s ScrollbarWidget) AlignStart() ScrollbarWidget {
	s.align = layout.Start
	return s
}

func (s ScrollbarWidget) AlignMiddle() ScrollbarWidget {
	s.align = layout.Middle
	return s
}

func (s ScrollbarWidget) AlignEnd() ScrollbarWidget {
	s.align = layout.End
	return s
}

func (s ScrollbarWidget) StickToEnd() ScrollbarWidget {
	s.stickToEnd = true
	return s
}

func (s ScrollbarWidget) ScrollAnyAxis() ScrollbarWidget {
	s.scrollAnyAxis = true
	return s
}

func (s ScrollbarWidget) Overlay() ScrollbarWidget {
	s.overlay = true
	return s
}

func (s ScrollbarWidget) Style(value flowstyle.Style) ScrollbarWidget {
	s.customStyle = value
	return s
}

func (s ScrollbarWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, s.child)
	key := frame.ClaimKey(ctx, stateutil.KindScrollbar, s.key)
	state := frame.UseState[scrollbarState](ctx, key, stateSlotScrollbar)
	state.list.Axis = s.axis
	state.list.Gap = 0
	state.list.Alignment = s.align
	state.list.ScrollToEnd = s.stickToEnd
	state.list.ScrollAnyAxis = s.scrollAnyAxis
	if s.disabled {
		gtx = gtx.Disabled()
	}
	styleState := scrollbarStyleState(&state.bar, s.disabled)
	return LayoutStyled(ctx, gtx, key, styleState, s.customStyle, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return layoutScrollbarList(ctx, gtx, key, &state.list, &state.bar, 1, s.disabled, s.overlay, s.customStyle, func(gtx layout.Context, _ int) layout.Dimensions {
			return s.child.Layout(ctx, gtx)
		})
	}))
}

func derivedScrollbarState(ctx *frame.Context, owner string) *widget.Scrollbar {
	key := frame.ClaimDerivedKey(ctx, stateutil.KindScrollbar, owner, "bar")
	return frame.UseState[widget.Scrollbar](ctx, key, stateSlotScrollbarBar)
}

func layoutScrollbarList(ctx *frame.Context, gtx layout.Context, key string, list *layout.List, bar *widget.Scrollbar, count int, disabled, overlay bool, custom flowstyle.Style, item layout.ListElement) layout.Dimensions {
	if disabled {
		gtx = gtx.Disabled()
	}
	activeTheme := frame.ActiveTheme(ctx)
	style := resolvedScrollbarStyle(ctx, gtx, key, bar, disabled, custom)
	axis := list.Axis
	originalConstraints := gtx.Constraints
	barWidth := gtx.Dp(style.Width())
	reservedWidth := barWidth + gtx.Dp(max(activeTheme.Components.Scrollbar.ContentGap, 0))
	majorAxisLimit := axis.Convert(originalConstraints.Max).X
	// ponytail: overflow changes settle on the next frame to avoid laying out interactive children twice.
	reserveScrollbar := !overlay && count > 0 && list.Position.Length > majorAxisLimit
	if reserveScrollbar {
		maxConstraints := axis.Convert(gtx.Constraints.Max)
		minConstraints := axis.Convert(gtx.Constraints.Min)
		maxConstraints.Y = max(maxConstraints.Y-reservedWidth, 0)
		minConstraints.Y = max(minConstraints.Y-reservedWidth, 0)
		gtx.Constraints.Max = axis.Convert(maxConstraints)
		gtx.Constraints.Min = axis.Convert(minConstraints)
	}

	listDims := layoutTrackedList(ctx, gtx, list, count, item)
	gtx.Constraints = originalConstraints
	majorAxisSize := axis.Convert(listDims.Size).X
	start, end := scrollbarViewport(list.Position, count, majorAxisSize)
	scrollable := start != 0 || end != 1
	if !overlay && reserveScrollbar != scrollable {
		gtx.Execute(op.InvalidateCmd{})
	}

	gtx.Constraints.Min = listDims.Size
	if reserveScrollbar {
		minor := axis.Convert(gtx.Constraints.Min)
		minor.Y += reservedWidth
		gtx.Constraints.Min = axis.Convert(minor)
	}
	anchor := layout.E
	if axis == layout.Horizontal {
		anchor = layout.S
	}
	anchor.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return style.Layout(gtx, axis, start, end)
	})

	if scrollable {
		if distance := bar.ScrollDistance(); distance != 0 {
			list.ScrollBy(distance * float32(count))
		}
	}
	if reserveScrollbar {
		minor := axis.Convert(listDims.Size)
		minor.Y += reservedWidth
		listDims.Size = axis.Convert(minor)
	}
	return listDims
}

// LayoutTrackedScrollbar lays out a Gio list with FlowUI's scrollbar style.
func LayoutTrackedScrollbar(ctx *frame.Context, gtx layout.Context, list *layout.List, bar *widget.Scrollbar, count int, disabled, overlay bool, item layout.ListElement) layout.Dimensions {
	return layoutScrollbarList(ctx, gtx, "", list, bar, count, disabled, overlay, flowstyle.Style{}, item)
}

func resolvedScrollbarStyle(ctx *frame.Context, gtx layout.Context, key string, bar *widget.Scrollbar, disabled bool, custom flowstyle.Style) material.ScrollbarStyle {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Scrollbar
	base := scrollbarStyleFor(activeTheme, bar, disabled)
	state := scrollbarStyleState(bar, disabled)
	defaults := flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.
			Width(tokens.TrackWidth).
			Background(flowstyle.SolidColor{Color: base.Track.Color})).
		Part(flowstyle.PartThumb, flowstyle.Style{}.
			Width(tokens.ThumbWidth).
			Background(flowstyle.SolidColor{Color: base.Indicator.Color}).
			Radius(tokens.Radius))

	track := resolveScrollbarPart(ctx, gtx, key, flowstyle.PartTrack, state, defaults, custom)
	thumb := resolveScrollbarPart(ctx, gtx, key, flowstyle.PartThumb, state, defaults, custom)
	trackWidth := tokens.TrackWidth
	thumbWidth := tokens.ThumbWidth
	if track.Box != nil && track.Box.Width != nil {
		trackWidth = max(*track.Box.Width, 1)
	}
	if thumb.Box != nil && thumb.Box.Width != nil {
		thumbWidth = max(*thumb.Box.Width, 1)
	}
	base.Indicator.MinorWidth = thumbWidth
	base.Track.MinorPadding = max((trackWidth-thumbWidth)/2, 0)
	if track.Paint != nil {
		if brush, ok := styleruntime.Brush(track.Paint.Background); ok {
			base.Track.Color = styleruntime.ApplyOpacity(brush.ColorAt(.5), track.Paint.Opacity)
		}
	}
	if thumb.Paint != nil {
		if brush, ok := styleruntime.Brush(thumb.Paint.Background); ok {
			indicator := styleruntime.ApplyOpacity(brush.ColorAt(.5), thumb.Paint.Opacity)
			base.Indicator.Color = indicator
			base.Indicator.HoverColor = indicator
		}
		if thumb.Paint.Radius != nil {
			base.Indicator.CornerRadius = *thumb.Paint.Radius
		}
	}
	return base
}

func scrollbarStyleState(bar *widget.Scrollbar, disabled bool) flowstyle.StyleState {
	dragging := !disabled && bar.Dragging()
	return flowstyle.StyleState{Pressed: dragging, Dragging: dragging, Disabled: disabled}
}

func resolveScrollbarPart(ctx *frame.Context, gtx layout.Context, key string, part flowstyle.Part, state flowstyle.StyleState, defaults, custom flowstyle.Style) flowstyle.ResolvedStyle {
	if key == "" {
		return styleruntime.ResolvePartStatic(ctx, part, state, defaults, flowstyle.Style{}, flowstyle.Style{}, custom)
	}
	return styleruntime.ResolvePart(ctx, gtx, key, part, state, defaults, flowstyle.Style{}, flowstyle.Style{}, custom)
}

func scrollbarStyleFor(activeTheme *theme.Theme, state *widget.Scrollbar, disabled bool) material.ScrollbarStyle {
	tokens := activeTheme.Components.Scrollbar
	style := material.Scrollbar(theme.MaterialOf(activeTheme), state)
	style.Track.MajorPadding = tokens.MajorPadding
	style.Track.MinorPadding = max((tokens.TrackWidth-tokens.ThumbWidth)/2, 0)
	style.Track.Color = color.NRGBA{}
	style.Indicator.MajorMinLen = tokens.MinThumbLength
	style.Indicator.MinorWidth = tokens.ThumbWidth
	style.Indicator.CornerRadius = tokens.Radius
	style.Indicator.Color = scrollbarColor(activeTheme.Palette.Foreground, tokens.ThumbOpacity)
	style.Indicator.HoverColor = scrollbarColor(activeTheme.Palette.Foreground, tokens.HoverOpacity)
	if disabled {
		style.Indicator.Color = activeTheme.DisabledColor(style.Indicator.Color)
		style.Indicator.HoverColor = style.Indicator.Color
	} else if state.Dragging() {
		style.Indicator.Color = style.Indicator.HoverColor
	}
	return style
}

func scrollbarColor(value color.NRGBA, opacity float32) color.NRGBA {
	opacity = min(max(opacity, 0), 1)
	value.A = byte(float32(value.A)*opacity + .5)
	return value
}

// scrollbarViewport maps Gio's estimated list position to the normalized range used by widget.Scrollbar.
func scrollbarViewport(position layout.Position, elements, majorAxisSize int) (start, end float32) {
	if elements <= 0 || position.Length <= 0 || majorAxisSize >= position.Length {
		return 0, 1
	}
	length := float32(position.Length)
	elementLength := length / float32(elements)
	start = min(max((float32(position.First)*elementLength+float32(position.Offset))/length, 0), 1)
	end = min(max((float32(position.First+position.Count)*elementLength+float32(position.OffsetLast))/length, 0), 1)
	visibleFraction := min(max(float32(majorAxisSize)/length, 0), 1)
	viewportFraction := end - start
	if viewportFraction < 1 {
		start -= start / (1 - viewportFraction) * (visibleFraction - viewportFraction)
		end += (1 - end) / (1 - viewportFraction) * (visibleFraction - viewportFraction)
	}
	return min(max(start, 0), 1), min(max(end, 0), 1)
}
