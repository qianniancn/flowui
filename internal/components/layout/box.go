package layoutui

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowlayout "github.com/qianniancn/FlowUI/internal/layout"
)

type BoxWidget struct {
	child         frame.Widget
	width         unit.Dp
	height        unit.Dp
	minWidth      unit.Dp
	maxWidth      unit.Dp
	minHeight     unit.Dp
	maxHeight     unit.Dp
	paddingTop    unit.Dp
	paddingRight  unit.Dp
	paddingBottom unit.Dp
	paddingLeft   unit.Dp
	marginTop     unit.Dp
	marginRight   unit.Dp
	marginBottom  unit.Dp
	marginLeft    unit.Dp
	hasMinWidth   bool
	hasMaxWidth   bool
	hasMinHeight  bool
	hasMaxHeight  bool
	hasAlign      bool
	fillWidth     bool
	fillHeight    bool
	align         Align
	overflow      Overflow
}

func Box(child frame.Widget) BoxWidget {
	return BoxWidget{child: child}
}

func (b BoxWidget) Width(dp int) BoxWidget {
	b.width = unit.Dp(dp)
	return b
}

func (b BoxWidget) Height(dp int) BoxWidget {
	b.height = unit.Dp(dp)
	return b
}

func (b BoxWidget) MinWidth(dp int) BoxWidget {
	b.minWidth = unit.Dp(dp)
	b.hasMinWidth = true
	return b
}

func (b BoxWidget) MaxWidth(dp int) BoxWidget {
	b.maxWidth = unit.Dp(dp)
	b.hasMaxWidth = true
	return b
}

func (b BoxWidget) MinHeight(dp int) BoxWidget {
	b.minHeight = unit.Dp(dp)
	b.hasMinHeight = true
	return b
}

func (b BoxWidget) MaxHeight(dp int) BoxWidget {
	b.maxHeight = unit.Dp(dp)
	b.hasMaxHeight = true
	return b
}

func (b BoxWidget) FillWidth() BoxWidget {
	b.fillWidth = true
	return b
}

func (b BoxWidget) FillHeight() BoxWidget {
	b.fillHeight = true
	return b
}

func (b BoxWidget) Padding(dp int) BoxWidget {
	padding := unit.Dp(dp)
	b.paddingTop = padding
	b.paddingRight = padding
	b.paddingBottom = padding
	b.paddingLeft = padding
	return b
}

func (b BoxWidget) PaddingTop(dp int) BoxWidget {
	b.paddingTop = unit.Dp(dp)
	return b
}

func (b BoxWidget) PaddingRight(dp int) BoxWidget {
	b.paddingRight = unit.Dp(dp)
	return b
}

func (b BoxWidget) PaddingBottom(dp int) BoxWidget {
	b.paddingBottom = unit.Dp(dp)
	return b
}

func (b BoxWidget) PaddingLeft(dp int) BoxWidget {
	b.paddingLeft = unit.Dp(dp)
	return b
}

func (b BoxWidget) Margin(dp int) BoxWidget {
	margin := unit.Dp(dp)
	b.marginTop = margin
	b.marginRight = margin
	b.marginBottom = margin
	b.marginLeft = margin
	return b
}

func (b BoxWidget) MarginTop(dp int) BoxWidget {
	b.marginTop = unit.Dp(dp)
	return b
}

func (b BoxWidget) MarginRight(dp int) BoxWidget {
	b.marginRight = unit.Dp(dp)
	return b
}

func (b BoxWidget) MarginBottom(dp int) BoxWidget {
	b.marginBottom = unit.Dp(dp)
	return b
}

func (b BoxWidget) MarginLeft(dp int) BoxWidget {
	b.marginLeft = unit.Dp(dp)
	return b
}

func (b BoxWidget) Align(align Align) BoxWidget {
	b.align = align
	b.hasAlign = true
	return b
}

func (b BoxWidget) Clip() BoxWidget {
	b.overflow = OverflowHidden
	return b
}

func (b BoxWidget) Overflow(overflow Overflow) BoxWidget {
	b.overflow = overflow
	return b
}

func (b BoxWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, b.child)
	layoutBox := func(gtx layout.Context) layout.Dimensions {
		return b.layoutContent(ctx, gtx)
	}
	if b.hasMargin() {
		return b.marginInset().Layout(gtx, layoutBox)
	}
	return layoutBox(gtx)
}

func (b BoxWidget) layoutContent(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	b.applyConstraints(&gtx)

	layoutChild := func(gtx layout.Context) layout.Dimensions {
		return b.child.Layout(ctx, gtx)
	}
	if b.hasAlign {
		layoutChild = func(gtx layout.Context) layout.Dimensions {
			return b.align.Direction().Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return b.child.Layout(ctx, gtx)
			})
		}
	}
	if b.paddingTop == 0 && b.paddingRight == 0 && b.paddingBottom == 0 && b.paddingLeft == 0 {
		return b.layoutOverflow(gtx, layoutChild)
	}
	return b.layoutOverflow(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.paddingInset().Layout(gtx, layoutChild)
	})
}

func (b BoxWidget) layoutOverflow(gtx layout.Context, child layout.Widget) layout.Dimensions {
	if b.overflow != OverflowHidden {
		return child(gtx)
	}
	macro := op.Record(gtx.Ops)
	dims := child(gtx)
	call := macro.Stop()
	dims.Size = gtx.Constraints.Constrain(dims.Size)
	defer clip.Rect{Max: dims.Size}.Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
	return dims
}

func (b BoxWidget) applyConstraints(gtx *layout.Context) {
	if b.hasMaxWidth {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(b.maxWidth))
		gtx.Constraints.Min.X = min(gtx.Constraints.Min.X, gtx.Constraints.Max.X)
	}
	if b.hasMaxHeight {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(b.maxHeight))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
	if b.hasMinWidth {
		gtx.Constraints.Min.X = min(max(gtx.Constraints.Min.X, gtx.Dp(b.minWidth)), gtx.Constraints.Max.X)
	}
	if b.hasMinHeight {
		gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(b.minHeight)), gtx.Constraints.Max.Y)
	}
	if b.fillWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if b.fillHeight {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if b.width > 0 {
		width := gtx.Dp(b.width)
		gtx.Constraints.Min.X = width
		gtx.Constraints.Max.X = width
	}
	if b.height > 0 {
		height := gtx.Dp(b.height)
		gtx.Constraints.Min.Y = height
		gtx.Constraints.Max.Y = height
	}
}

func (b BoxWidget) paddingInset() layout.Inset {
	return layout.Inset{
		Top:    b.paddingTop,
		Right:  b.paddingRight,
		Bottom: b.paddingBottom,
		Left:   b.paddingLeft,
	}
}

func (b BoxWidget) marginInset() layout.Inset {
	return layout.Inset{
		Top:    b.marginTop,
		Right:  b.marginRight,
		Bottom: b.marginBottom,
		Left:   b.marginLeft,
	}
}

func (b BoxWidget) hasMargin() bool {
	return b.marginTop != 0 || b.marginRight != 0 || b.marginBottom != 0 || b.marginLeft != 0
}

type Overflow = flowlayout.Overflow

const (
	OverflowVisible = flowlayout.OverflowVisible
	OverflowHidden  = flowlayout.OverflowHidden
)

type Align = flowlayout.Align

const (
	AlignTopStart    = flowlayout.AlignTopStart
	AlignTop         = flowlayout.AlignTop
	AlignTopEnd      = flowlayout.AlignTopEnd
	AlignStart       = flowlayout.AlignStart
	AlignCenter      = flowlayout.AlignCenter
	AlignEnd         = flowlayout.AlignEnd
	AlignBottomStart = flowlayout.AlignBottomStart
	AlignBottom      = flowlayout.AlignBottom
	AlignBottomEnd   = flowlayout.AlignBottomEnd
)

type SpacerWidget struct {
	width  unit.Dp
	height unit.Dp
}

func Spacer(width, height int) SpacerWidget {
	return SpacerWidget{
		width:  unit.Dp(width),
		height: unit.Dp(height),
	}
}

func (s SpacerWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Spacer{
		Width:  s.width,
		Height: s.height,
	}.Layout(gtx)
}
