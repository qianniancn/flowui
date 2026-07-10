package flowui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type TextWidget struct {
	text     string
	size     unit.Sp
	color    color.NRGBA
	hasColor bool
	weight   font.Weight
}

func Text(text string) TextWidget {
	return TextWidget{text: text}
}

func (t TextWidget) Size(sp float32) TextWidget {
	t.size = unit.Sp(sp)
	return t
}

func (t TextWidget) Color(c color.NRGBA) TextWidget {
	t.color = c
	t.hasColor = true
	return t
}

func (t TextWidget) Weight(w font.Weight) TextWidget {
	t.weight = w
	return t
}

func (t TextWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	size := t.size
	if size == 0 {
		size = ctx.Theme.Typography.BodySize
	}
	label := material.Label(ctx.Theme.Material, size, t.text)
	if t.hasColor {
		label.Color = t.color
	} else {
		label.Color = ctx.foregroundColor()
	}
	if t.weight != 0 {
		label.Font.Weight = t.weight
	}
	return label.Layout(gtx)
}
