package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
)

const stateSlotVisualOutset = "visual-outset"

// VisualOutsetState retains visual overflow seen in a clipping viewport. A
// virtualized list keeps its previous maximum until every item is observed
// again; non-virtualized content updates exactly each frame.
type VisualOutsetState struct {
	outset frame.VisualOutset
}

func visualOutsetStateFor(ctx *frame.Context, kind stateutil.Kind, owner string) *VisualOutsetState {
	key := frame.ClaimDerivedKey(ctx, kind, owner, "visual-outset")
	return frame.UseState[VisualOutsetState](ctx, key, stateSlotVisualOutset)
}

func (s *VisualOutsetState) observe(outset frame.VisualOutset, complete bool) bool {
	if s == nil {
		return false
	}
	next := outset
	if !complete {
		next = s.outset.Max(outset)
	}
	if next == s.outset {
		return false
	}
	s.outset = next
	return true
}

func visualOutsetForListItem(axis layout.Axis, index, count int, outset frame.VisualOutset) frame.VisualOutset {
	if outset.Empty() || count <= 0 {
		return frame.VisualOutset{}
	}
	first := index == 0
	last := index == count-1
	if axis == layout.Horizontal {
		return frame.VisualOutset{
			Top:    outset.Top,
			Right:  whenVisualOutset(last, outset.Right),
			Bottom: outset.Bottom,
			Left:   whenVisualOutset(first, outset.Left),
		}
	}
	return frame.VisualOutset{
		Top:    whenVisualOutset(first, outset.Top),
		Right:  outset.Right,
		Bottom: whenVisualOutset(last, outset.Bottom),
		Left:   outset.Left,
	}
}

func whenVisualOutset(enabled bool, value int) int {
	if !enabled {
		return 0
	}
	return value
}

// layoutVisualOutset reserves pixels around child while preserving the
// parent's coordinate system. The child is shifted inside the returned box,
// so its pointer regions remain limited to the actual visual surface.
func layoutVisualOutset(ctx *frame.Context, gtx layout.Context, outset frame.VisualOutset, child layout.Widget) layout.Dimensions {
	if child == nil || outset.Empty() {
		if child == nil {
			return layout.Dimensions{}
		}
		return child(gtx)
	}
	left := max(outset.Left, 0)
	top := max(outset.Top, 0)
	right := max(outset.Right, 0)
	bottom := max(outset.Bottom, 0)
	horizontal := visualOutsetSum(left, right)
	vertical := visualOutsetSum(top, bottom)
	original := gtx.Constraints
	childGtx := gtx
	childGtx.Constraints.Min = image.Pt(
		max(original.Min.X-horizontal, 0),
		max(original.Min.Y-vertical, 0),
	)
	childGtx.Constraints.Max = image.Pt(
		max(original.Max.X-horizontal, 0),
		max(original.Max.Y-vertical, 0),
	)

	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child(childGtx)
	})
	call := macro.Stop()
	placement.PlaceOffset(image.Pt(left, top))
	offset := op.Offset(image.Pt(left, top)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()

	dims.Size = original.Constrain(image.Pt(
		visualOutsetSum(dims.Size.X, horizontal),
		visualOutsetSum(dims.Size.Y, vertical),
	))
	if dims.Baseline != 0 {
		// Gio stores Baseline as the distance from the bottom edge.
		dims.Baseline = visualOutsetSum(dims.Baseline, bottom)
	}
	return dims
}

func visualOutsetSum(first, second int) int {
	if first <= 0 {
		return max(second, 0)
	}
	if second <= 0 {
		return first
	}
	limit := int(^uint(0) >> 1)
	if first > limit-second {
		return limit
	}
	return first + second
}

func layoutVisualOutsetList(ctx *frame.Context, gtx layout.Context, list *layout.List, count int, visual *VisualOutsetState, item layout.ListElement) layout.Dimensions {
	collector := new(frame.VisualOverflowCollector)
	restoreCollector := frame.PushVisualOverflowCollector(ctx, collector)
	collecting := frame.CollectingVisualOverflow(ctx)
	dims := layoutTrackedList(ctx, gtx, list, count, func(itemGtx layout.Context, index int) layout.Dimensions {
		outset := frame.VisualOutset{}
		if visual != nil {
			outset = visualOutsetForListItem(list.Axis, index, count, visual.outset)
		}
		return layoutVisualOutset(ctx, itemGtx, outset, func(childGtx layout.Context) layout.Dimensions {
			return item(childGtx, index)
		})
	})
	restoreCollector()
	complete := list.Position.First == 0 && list.Position.Count >= count
	if collecting && visual != nil && visual.observe(collector.Outset(), complete) && ctx != nil {
		ctx.Invalidate()
	}
	return dims
}
