package card

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type CardHeaderWidget struct {
	children []frame.Widget
	gap      unit.Dp
	hasGap   bool
}

func CardHeader(children ...frame.Widget) CardHeaderWidget {
	return CardHeaderWidget{children: children}
}

func (h CardHeaderWidget) Gap(dp int) CardHeaderWidget {
	h.gap = unit.Dp(max(dp, 0))
	h.hasGap = true
	return h
}

func (h CardHeaderWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	gap := frame.ActiveTheme(ctx).Components.Card.HeaderGap
	if h.hasGap {
		gap = h.gap
	}
	return layoutui.Column(nonNilWidgets(h.children)...).Gap(int(gap)).Layout(ctx, gtx)
}

type CardContentWidget struct {
	children []frame.Widget
	gap      unit.Dp
	hasGap   bool
}

func CardContent(children ...frame.Widget) CardContentWidget {
	return CardContentWidget{children: children}
}

func (c CardContentWidget) Gap(dp int) CardContentWidget {
	c.gap = unit.Dp(max(dp, 0))
	c.hasGap = true
	return c
}

func (c CardContentWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	gap := frame.ActiveTheme(ctx).Components.Card.ContentGap
	if c.hasGap {
		gap = c.gap
	}
	return layoutui.Column(nonNilWidgets(c.children)...).Gap(int(gap)).Layout(ctx, gtx)
}

type CardFooterWidget struct {
	children []frame.Widget
	gap      unit.Dp
	hasGap   bool
}

func CardFooter(children ...frame.Widget) CardFooterWidget {
	return CardFooterWidget{children: children}
}

func (f CardFooterWidget) Gap(dp int) CardFooterWidget {
	f.gap = unit.Dp(max(dp, 0))
	f.hasGap = true
	return f
}

func (f CardFooterWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	gap := frame.ActiveTheme(ctx).Components.Card.FooterGap
	if f.hasGap {
		gap = f.gap
	}
	return layoutui.Row(nonNilWidgets(f.children)...).Gap(int(gap)).AlignMiddle().Layout(ctx, gtx)
}

type CardTitleWidget struct {
	value      string
	size       unit.Sp
	lineHeight unit.Sp
	color      color.NRGBA
	weight     font.Weight
	hasColor   bool
}

func CardTitle(value string) CardTitleWidget {
	return CardTitleWidget{value: value}
}

func (t CardTitleWidget) Size(sp float32) CardTitleWidget {
	t.size = unit.Sp(max(sp, 0))
	return t
}

func (t CardTitleWidget) LineHeight(sp float32) CardTitleWidget {
	t.lineHeight = unit.Sp(max(sp, 0))
	return t
}

func (t CardTitleWidget) Color(value color.NRGBA) CardTitleWidget {
	t.color = value
	t.hasColor = true
	return t
}

func (t CardTitleWidget) Weight(weight font.Weight) CardTitleWidget {
	t.weight = weight
	return t
}

func (t CardTitleWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Card
	size := tokens.TitleSize
	if t.size > 0 {
		size = t.size
	}
	lineHeight := tokens.TitleLineHeight
	if t.lineHeight > 0 {
		lineHeight = t.lineHeight
	}
	label := material.Label(frame.ActiveTheme(ctx).Material, size, t.value)
	label.Color = ctx.ForegroundColor()
	label.Font.Weight = font.Medium
	label.LineHeight = lineHeight
	label.WrapPolicy = text.WrapHeuristically
	if t.hasColor {
		label.Color = t.color
	}
	if t.weight != 0 {
		label.Font.Weight = t.weight
	}
	return label.Layout(gtx)
}

type CardDescriptionWidget struct {
	value      string
	size       unit.Sp
	lineHeight unit.Sp
	color      color.NRGBA
	hasColor   bool
}

func CardDescription(value string) CardDescriptionWidget {
	return CardDescriptionWidget{value: value}
}

func (d CardDescriptionWidget) Size(sp float32) CardDescriptionWidget {
	d.size = unit.Sp(max(sp, 0))
	return d
}

func (d CardDescriptionWidget) LineHeight(sp float32) CardDescriptionWidget {
	d.lineHeight = unit.Sp(max(sp, 0))
	return d
}

func (d CardDescriptionWidget) Color(value color.NRGBA) CardDescriptionWidget {
	d.color = value
	d.hasColor = true
	return d
}

func (d CardDescriptionWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	size := activeTheme.Components.Card.DescriptionSize
	if d.size > 0 {
		size = d.size
	}
	lineHeight := activeTheme.Components.Card.DescriptionLineHeight
	if d.lineHeight > 0 {
		lineHeight = d.lineHeight
	}
	label := material.Label(activeTheme.Material, size, d.value)
	label.Color = activeTheme.Palette.MutedForeground
	label.LineHeight = lineHeight
	label.WrapPolicy = text.WrapHeuristically
	if d.hasColor {
		label.Color = d.color
	}
	return label.Layout(gtx)
}
