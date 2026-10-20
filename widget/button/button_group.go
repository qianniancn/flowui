package button

import (
	"github.com/qianniancn/flowui/style"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type GroupOrientation uint8

const (
	GroupHorizontal GroupOrientation = iota
	GroupVertical
)

type buttonGroupPosition uint8

const (
	buttonGroupSingle buttonGroupPosition = iota
	buttonGroupStart
	buttonGroupMiddle
	buttonGroupEnd
)

type buttonGroupItemStyle struct {
	grouped     bool
	orientation GroupOrientation
	position    buttonGroupPosition
}

type GroupWidget struct {
	override    *theme.ButtonOverride
	buttons     []Widget
	orientation GroupOrientation
	variant     Variant
	size        Size
	disabled    bool
	fullWidth   bool
	separators  bool
}

func Group(buttons ...Widget) GroupWidget {
	return GroupWidget{buttons: append([]Widget(nil), buttons...)}
}

// GroupStart applies the leading corner treatment used by Group to a
// button rendered by another composite widget.
func GroupStart(value Widget) Widget {
	value.group = buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupStart}
	return value
}

// GroupEnd applies the trailing corner treatment used by Group to a
// button rendered by another composite widget.
func GroupEnd(value Widget) Widget {
	value.group = buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupEnd}
	return value
}

func (g GroupWidget) Orientation(value GroupOrientation) GroupWidget {
	g.orientation = value
	return g
}

func (g GroupWidget) Variant(value Variant) GroupWidget {
	g.variant = value
	return g
}

func (g GroupWidget) Size(value Size) GroupWidget {
	g.size = value
	return g
}

func (g GroupWidget) Disabled(value bool) GroupWidget {
	g.disabled = value
	return g
}

func (g GroupWidget) FullWidth() GroupWidget {
	g.fullWidth = true
	return g
}

func (g GroupWidget) Separators(visible bool) GroupWidget {
	g.separators = visible
	return g
}

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return g.layoutWithOverride(ctx, gtx, g.override)
}

func (g GroupWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ButtonOverride) layout.Dimensions {
	if len(g.buttons) == 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints.Min = image.Point{}
	var items []buttonGroupItemLayout
	if g.separators {
		items = make([]buttonGroupItemLayout, len(g.buttons))
	}
	children := make([]layout.FlexChild, len(g.buttons))
	axis := layout.Horizontal
	if g.orientation == GroupVertical {
		axis = layout.Vertical
	}
	if g.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	// Only the vertical intrinsic-width pass needs prepared values across its
	// measurement and layout passes. Keep them local to this layout call so the
	// behavior Widget does not retain frame-specific visual operations.
	if axis == layout.Vertical && !g.fullWidth {
		type preparedButton struct {
			value   Widget
			content buttonPreparedContent
		}
		buttons := make([]preparedButton, len(g.buttons))
		verticalWidth := 0
		for index := range g.buttons {
			value := g.prepareButton(g.buttons[index], index, len(g.buttons))
			buttons[index] = preparedButton{value: value, content: value.prepareContent(ctx, gtx, override)}
			verticalWidth = max(verticalWidth, buttons[index].content.width)
		}
		verticalWidth = min(verticalWidth, gtx.Constraints.Max.X)
		for index := range buttons {
			preparedContent := &buttons[index].content
			layoutButton := func(gtx layout.Context) layout.Dimensions {
				if verticalWidth > 0 {
					gtx.Constraints.Min.X = verticalWidth
					gtx.Constraints.Max.X = verticalWidth
				}
				dims, resolved := layoutWithClickableResolved(buttons[index].value, ctx, gtx, nil, true, override, preparedContent)
				if g.separators {
					items[index] = buttonGroupItemLayout{
						dims:       dims,
						foreground: buttonGroupForeground(resolved),
						opacity:    buttonOpacity(resolved.root),
						group:      resolved.visual.Group,
					}
				}
				return dims
			}
			children[index] = layout.Rigid(layoutButton)
		}

		dims := layout.Flex{Axis: axis, Alignment: layout.Middle}.Layout(gtx, children...)
		root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		if g.separators {
			g.drawSeparators(gtx, dims.Size, items)
		}
		root.Pop()
		return dims
	}

	for index := range g.buttons {
		layoutButton := func(gtx layout.Context) layout.Dimensions {
			button := g.prepareButton(g.buttons[index], index, len(g.buttons))
			dims, resolved := layoutWithClickableResolved(button, ctx, gtx, nil, true, override, nil)
			if g.separators {
				items[index] = buttonGroupItemLayout{
					dims:       dims,
					foreground: buttonGroupForeground(resolved),
					opacity:    buttonOpacity(resolved.root),
					group:      resolved.visual.Group,
				}
			}
			return dims
		}
		if g.fullWidth && axis == layout.Horizontal {
			children[index] = layout.Flexed(1, layoutButton)
		} else {
			children[index] = layout.Rigid(layoutButton)
		}
	}

	dims := layout.Flex{Axis: axis, Alignment: layout.Middle}.Layout(gtx, children...)
	root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	if g.separators {
		g.drawSeparators(gtx, dims.Size, items)
	}
	root.Pop()
	return dims
}

func (g GroupWidget) prepareButton(button Widget, index, count int) Widget {
	if !button.variantSet {
		button.variant = g.variant
	}
	if !button.sizeSet {
		button.size = g.size
	}
	if !button.disabledSet {
		button.disabled = g.disabled
	}
	if g.fullWidth {
		button.fullWidth = true
	}
	position := buttonGroupMiddle
	if count == 1 {
		position = buttonGroupSingle
	} else if index == 0 {
		position = buttonGroupStart
	} else if index == count-1 {
		position = buttonGroupEnd
	}
	button.group = buttonGroupItemStyle{grouped: true, orientation: g.orientation, position: position}
	return button
}

type buttonGroupItemLayout struct {
	dims       layout.Dimensions
	foreground color.NRGBA
	opacity    float32
	group      theme.ButtonGroupVisual
}

func buttonGroupForeground(resolved buttonStyle) color.NRGBA {
	return buttonForeground(resolved.visual, resolved.content)
}

func (g GroupWidget) drawSeparators(gtx layout.Context, size image.Point, items []buttonGroupItemLayout) {
	if len(items) == 0 {
		return
	}
	main := 0
	for index := 1; index < len(items); index++ {
		previous := g.orientationSize(items[index-1].dims.Size)
		main += previous.X
		axis := theme.ButtonGroupHorizontal
		if g.orientation == GroupVertical {
			axis = theme.ButtonGroupVertical
		}
		// A separator belongs to the item after the boundary. Resolve its
		// renderer from that item so a provider may vary separator treatment by
		// semantic variant or size without leaking the first item's strategy.
		group := items[index].group
		group.SeparatorRenderer(gtx, theme.ButtonGroupSeparatorRequest{
			Size:       size,
			MainOffset: main,
			Axis:       axis,
			Foreground: items[index].foreground,
			Opacity:    items[index].opacity,
		})
	}
}

func (g GroupWidget) orientationSize(value image.Point) image.Point {
	if g.orientation == GroupVertical {
		return image.Pt(value.Y, value.X)
	}
	return value
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w GroupWidget) Style(values ...style.Style) GroupWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ButtonOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
