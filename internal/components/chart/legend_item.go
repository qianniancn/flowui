package chart

import (
	"gioui.org/gesture"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// LegendItem tracks pointer-only legend interaction without entering the
// keyboard focus order.
type LegendItem struct {
	click           gesture.Click
	requestedClicks int
}

func (c *LegendItem) Click() {
	c.requestedClicks++
}

func (c *LegendItem) Clicked(gtx layout.Context) bool {
	if c.requestedClicks > 0 {
		c.requestedClicks--
		return true
	}
	for {
		event, ok := c.click.Update(gtx.Source)
		if !ok {
			return false
		}
		if event.Kind == gesture.KindClick {
			return true
		}
	}
}

func (c *LegendItem) Hovered() bool {
	return c.click.Hovered()
}

func (c *LegendItem) Layout(gtx layout.Context, child layout.Widget) layout.Dimensions {
	for {
		_, ok := c.click.Update(gtx.Source)
		if !ok {
			break
		}
	}
	macro := op.Record(gtx.Ops)
	dims := child(gtx)
	call := macro.Stop()
	area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	c.click.Add(gtx.Ops)
	call.Add(gtx.Ops)
	area.Pop()
	return dims
}
