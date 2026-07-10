package text

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type Widget struct {
	text     string
	size     unit.Sp
	color    color.NRGBA
	hasColor bool
	weight   font.Weight
}

func New(text string) Widget {
	return Widget{text: text}
}

func (t Widget) Size(sp float32) Widget {
	t.size = unit.Sp(sp)
	return t
}

func (t Widget) Color(c color.NRGBA) Widget {
	t.color = c
	t.hasColor = true
	return t
}

func (t Widget) Weight(w font.Weight) Widget {
	t.weight = w
	return t
}

func (t Widget) DefaultSize(sp float32) Widget {
	if t.size == 0 {
		t.size = unit.Sp(sp)
	}
	return t
}

func (t Widget) DefaultColor(c color.NRGBA) Widget {
	if !t.hasColor {
		t.color = c
		t.hasColor = true
	}
	return t
}

func (t Widget) DefaultWeight(w font.Weight) Widget {
	if t.weight == 0 {
		t.weight = w
	}
	return t
}

func (t Widget) ConfiguredSize() unit.Sp {
	return t.size
}

func (t Widget) ConfiguredColor() (color.NRGBA, bool) {
	return t.color, t.hasColor
}

func (t Widget) ConfiguredWeight() font.Weight {
	return t.weight
}

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	size := t.size
	if size == 0 {
		size = frame.ActiveTheme(ctx).Typography.BodySize
	}
	label := material.Label(frame.ActiveTheme(ctx).Material, size, t.text)
	if t.hasColor {
		label.Color = t.color
	} else {
		label.Color = ctx.ForegroundColor()
	}
	if t.weight != 0 {
		label.Font.Weight = t.weight
	}
	return label.Layout(gtx)
}
