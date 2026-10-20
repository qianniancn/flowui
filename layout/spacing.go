package layout

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

// SpacerWidget occupies fixed space.
type SpacerWidget struct {
	width  unit.Dp
	height unit.Dp
}

// Spacer returns a fixed-size empty widget.
func Spacer(width, height int) SpacerWidget {
	return SpacerWidget{width: unit.Dp(width), height: unit.Dp(height)}
}

// Layout implements Widget.
func (s SpacerWidget) Layout(_ *Context, gtx layout.Context) layout.Dimensions {
	return layout.Spacer{Width: s.width, Height: s.height}.Layout(gtx)
}
