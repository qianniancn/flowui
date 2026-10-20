package layout

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

const stateSlotVisualOutset = "visual-outset"

// VisualOutsetState retains visual overflow seen in a clipping viewport. A
// virtualized list keeps its previous maximum until every item is observed
// again; non-virtualized content updates exactly each core.
type VisualOutsetState struct {
	outset                 core.VisualOutset
	scrollbarDefaultsCache computed.DeclarationCache[scrollbarDeclarationKey, style.Style]
}

func visualOutsetStateFor(ctx *Context, kind core.Kind, owner string) *VisualOutsetState {
	key := core.ClaimDerivedKey(ctx, kind, owner, "visual-outset")
	return core.UseState[VisualOutsetState](ctx, key, stateSlotVisualOutset)
}

func (s *VisualOutsetState) observe(outset core.VisualOutset, complete bool) bool {
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

func visualOutsetForListItem(axis layout.Axis, index, count int, outset core.VisualOutset) core.VisualOutset {
	if outset.Empty() || count <= 0 {
		return core.VisualOutset{}
	}
	first := index == 0
	last := index == count-1
	if axis == layout.Horizontal {
		return core.VisualOutset{
			Top:    outset.Top,
			Right:  whenVisualOutset(last, outset.Right),
			Bottom: outset.Bottom,
			Left:   whenVisualOutset(first, outset.Left),
		}
	}
	return core.VisualOutset{
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
func layoutVisualOutset(ctx *Context, gtx layout.Context, outset core.VisualOutset, child layout.Widget) layout.Dimensions {
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
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
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

func layoutVisualOutsetList(ctx *Context, gtx layout.Context, list *layout.List, count int, visual *VisualOutsetState, item layout.ListElement) layout.Dimensions {
	collector := new(core.VisualOverflowCollector)
	restoreCollector := core.PushVisualOverflowCollector(ctx, collector)
	collecting := core.CollectingVisualOverflow(ctx)
	dims := layoutTrackedList(ctx, gtx, list, count, func(itemGtx layout.Context, index int) layout.Dimensions {
		outset := core.VisualOutset{}
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
