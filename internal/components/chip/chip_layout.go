package chip

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	textui "github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

func (c Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	resolved := c.resolveStyle(ctx, gtx)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if c.label != "" {
			semantic.LabelOp(c.label).Add(gtx.Ops)
		}
		return c.layoutContent(ctx, gtx, resolved)
	})
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, content)
}

func (c Widget) layoutContent(ctx *frame.Context, gtx layout.Context, style chipResolvedStyle) layout.Dimensions {
	gap := max(gtx.Dp(style.contentGap), 0)
	gapCount := 0
	if c.startContent != nil {
		gapCount++
	}
	if c.endContent != nil {
		gapCount++
	}
	gapsWidth := gap * gapCount
	labelMinimum := min(gtx.Dp(style.labelPaddingX)*2, max(gtx.Constraints.Max.X-gapsWidth, 0))
	accessoryWidth := max(gtx.Constraints.Max.X-gapsWidth-labelMinimum, 0)

	accessoryGtx := gtx
	accessoryGtx.Constraints.Min = image.Point{}
	accessoryGtx.Constraints.Max.X = accessoryWidth
	end := recordChipWidget(ctx, accessoryGtx, style.icon, c.endContent)
	accessoryGtx.Constraints.Max.X = max(accessoryWidth-end.dims.Size.X, 0)
	start := recordChipWidget(ctx, accessoryGtx, style.icon, c.startContent)

	labelGtx := gtx
	labelGtx.Constraints.Min = image.Point{}
	labelGtx.Constraints.Max.X = max(gtx.Constraints.Max.X-gapsWidth-start.dims.Size.X-end.dims.Size.X, 0)
	label := c.recordLabel(ctx, labelGtx, style)

	width := start.dims.Size.X + label.dims.Size.X + end.dims.Size.X + gapsWidth
	height := max(gtx.Constraints.Min.Y, max(start.dims.Size.Y, max(label.dims.Size.Y, end.dims.Size.Y)))
	size := gtx.Constraints.Constrain(image.Pt(width, height))

	x := 0
	if c.startContent != nil {
		placeChipChild(gtx, start, image.Pt(x, max((size.Y-start.dims.Size.Y)/2, 0)))
		x += start.dims.Size.X + gap
	}
	labelPosition := image.Pt(x, max((size.Y-label.dims.Size.Y)/2, 0))
	placeChipChild(gtx, label, labelPosition)
	x += label.dims.Size.X
	if c.endContent != nil {
		x += gap
		placeChipChild(gtx, end, image.Pt(x, max((size.Y-end.dims.Size.Y)/2, 0)))
	}

	baseline := size.Y - labelPosition.Y - label.dims.Size.Y + label.dims.Baseline
	return layout.Dimensions{Size: size, Baseline: max(baseline, 0)}
}

type recordedChipChild struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func recordChipWidget(ctx *frame.Context, gtx layout.Context, style flowstyle.ResolvedStyle, child frame.Widget) recordedChipChild {
	if child == nil {
		return recordedChipChild{}
	}
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layoutui.LayoutResolved(ctx, gtx, style, child)
	})
	return recordedChipChild{call: macro.Stop(), dims: dims, placement: placement}
}

func (c Widget) recordLabel(ctx *frame.Context, gtx layout.Context, style chipResolvedStyle) recordedChipChild {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layout.Inset{Left: style.labelPaddingX, Right: style.labelPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, style.label, textui.New(c.label))
		})
	})
	return recordedChipChild{call: macro.Stop(), dims: dims, placement: placement}
}

func placeChipChild(gtx layout.Context, child recordedChipChild, position image.Point) {
	child.placement.PlaceOffset(position)
	stack := op.Offset(position).Push(gtx.Ops)
	child.call.Add(gtx.Ops)
	stack.Pop()
}
