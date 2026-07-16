package button

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ButtonGroupOrientation uint8

const (
	ButtonGroupHorizontal ButtonGroupOrientation = iota
	ButtonGroupVertical
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
	orientation ButtonGroupOrientation
	position    buttonGroupPosition
}

type ButtonGroupWidget struct {
	buttons     []ButtonWidget
	orientation ButtonGroupOrientation
	variant     ButtonVariant
	size        ButtonSize
	disabled    bool
	fullWidth   bool
	separators  bool
}

func ButtonGroup(buttons ...ButtonWidget) ButtonGroupWidget {
	return ButtonGroupWidget{buttons: append([]ButtonWidget(nil), buttons...)}
}

func (g ButtonGroupWidget) Orientation(value ButtonGroupOrientation) ButtonGroupWidget {
	g.orientation = value
	return g
}

func (g ButtonGroupWidget) Variant(value ButtonVariant) ButtonGroupWidget {
	g.variant = value
	return g
}

func (g ButtonGroupWidget) Size(value ButtonSize) ButtonGroupWidget {
	g.size = value
	return g
}

func (g ButtonGroupWidget) Disabled(value bool) ButtonGroupWidget {
	g.disabled = value
	return g
}

func (g ButtonGroupWidget) FullWidth() ButtonGroupWidget {
	g.fullWidth = true
	return g
}

func (g ButtonGroupWidget) Separators(visible bool) ButtonGroupWidget {
	g.separators = visible
	return g
}

func (g ButtonGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if len(g.buttons) == 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints.Min = image.Point{}
	buttons := append([]ButtonWidget(nil), g.buttons...)
	items := make([]buttonGroupItemLayout, len(buttons))
	children := make([]layout.FlexChild, len(buttons))
	axis := layout.Horizontal
	if g.orientation == ButtonGroupVertical {
		axis = layout.Vertical
	}
	if g.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	for index := range buttons {
		buttons[index] = g.prepareButton(buttons[index], index, len(buttons))
	}
	verticalWidth := 0
	if axis == layout.Vertical && !g.fullWidth {
		for index := range buttons {
			buttons[index].prepared = buttons[index].prepareContent(ctx, gtx)
			buttons[index].preparedSet = true
			verticalWidth = max(verticalWidth, buttons[index].prepared.width)
		}
		verticalWidth = min(verticalWidth, gtx.Constraints.Max.X)
	}
	for index := range buttons {
		index := index
		layoutButton := func(gtx layout.Context) layout.Dimensions {
			if verticalWidth > 0 {
				gtx.Constraints.Min.X = verticalWidth
				gtx.Constraints.Max.X = verticalWidth
			}
			dims := buttons[index].Layout(ctx, gtx)
			items[index] = buttonGroupItemLayout{dims: dims, foreground: buttonGroupForeground(ctx, buttons[index])}
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
		g.drawSeparators(ctx, gtx, dims.Size, items)
	}
	root.Pop()
	return dims
}

func (g ButtonGroupWidget) prepareButton(button ButtonWidget, index, count int) ButtonWidget {
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
}

func buttonGroupForeground(ctx *frame.Context, button ButtonWidget) color.NRGBA {
	activeTheme := frame.ActiveTheme(ctx)
	foreground := buttonColors(activeTheme, button.variant).fg
	if button.disabled {
		foreground = activeTheme.DisabledColor(foreground)
	}
	return foreground
}

func (g ButtonGroupWidget) drawSeparators(ctx *frame.Context, gtx layout.Context, size image.Point, items []buttonGroupItemLayout) {
	tokens := frame.ActiveTheme(ctx).Components.ButtonGroup
	thickness := max(gtx.Dp(tokens.SeparatorWidth), 1)
	ratio := min(max(tokens.SeparatorLength, 0), 1)
	opacity := min(max(tokens.SeparatorOpacity, 0), 1)
	main := 0
	for index := 1; index < len(items); index++ {
		previous := g.orientationSize(items[index-1].dims.Size)
		main += previous.X
		foreground := items[index].foreground
		foreground.A = byte(float32(foreground.A)*opacity + 0.5)
		if g.orientation == ButtonGroupVertical {
			length := max(int(float32(items[index].dims.Size.X)*ratio), 1)
			rect := image.Rect((size.X-length)/2, main-thickness/2, (size.X+length)/2, main-thickness/2+thickness)
			paint.FillShape(gtx.Ops, foreground, clip.Rect(rect).Op())
			continue
		}
		length := max(int(float32(items[index].dims.Size.Y)*ratio), 1)
		rect := image.Rect(main-thickness/2, (size.Y-length)/2, main-thickness/2+thickness, (size.Y+length)/2)
		paint.FillShape(gtx.Ops, foreground, clip.Rect(rect).Op())
	}
}

func (g ButtonGroupWidget) orientationSize(value image.Point) image.Point {
	if g.orientation == ButtonGroupVertical {
		return image.Pt(value.Y, value.X)
	}
	return value
}
