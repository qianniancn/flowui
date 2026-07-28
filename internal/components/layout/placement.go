package layoutui

import (
	"cmp"
	"image"
	"slices"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
)

func layoutTrackedInset(ctx *frame.Context, gtx layout.Context, inset layout.Inset, child layout.Widget) layout.Dimensions {
	var placement frame.OverlayPlacement
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var dims layout.Dimensions
		dims, placement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(gtx)
		})
		return dims
	})
	placement.PlaceOffset(insetOffset(gtx, inset))
	return dims
}

// LayoutTrackedInset keeps overlay anchors aligned with an inset child.
func LayoutTrackedInset(ctx *frame.Context, gtx layout.Context, inset layout.Inset, child layout.Widget) layout.Dimensions {
	return layoutTrackedInset(ctx, gtx, inset, child)
}

func insetOffset(gtx layout.Context, inset layout.Inset) image.Point {
	top := gtx.Dp(inset.Top)
	right := gtx.Dp(inset.Right)
	bottom := gtx.Dp(inset.Bottom)
	left := gtx.Dp(inset.Left)
	if gtx.Constraints.Max.X-left-right < 0 {
		left = 0
	}
	if gtx.Constraints.Max.Y-top-bottom < 0 {
		top = 0
	}
	return image.Pt(left, top)
}

func layoutTrackedDirection(ctx *frame.Context, gtx layout.Context, direction layout.Direction, child layout.Widget) layout.Dimensions {
	var childDims layout.Dimensions
	var placement frame.OverlayPlacement
	dims := direction.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		childDims, placement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(gtx)
		})
		return childDims
	})
	placement.PlaceOffset(direction.Position(childDims.Size, dims.Size))
	return dims
}

// LayoutTrackedDirection keeps overlay anchors aligned with a positioned child.
func LayoutTrackedDirection(ctx *frame.Context, gtx layout.Context, direction layout.Direction, child layout.Widget) layout.Dimensions {
	return layoutTrackedDirection(ctx, gtx, direction, child)
}

type listChildPlacement struct {
	index     int
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func layoutTrackedList(ctx *frame.Context, gtx layout.Context, state *layout.List, count int, item layout.ListElement) layout.Dimensions {
	var storage [32]listChildPlacement
	children := storage[:0]
	dims := state.Layout(gtx, count, func(gtx layout.Context, index int) layout.Dimensions {
		dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return item(gtx, index)
		})
		children = append(children, listChildPlacement{index: index, dims: dims, placement: placement})
		return dims
	})
	slices.SortStableFunc(children, func(a, b listChildPlacement) int {
		return cmp.Compare(a.index, b.index)
	})
	children = deduplicateListChildren(children)
	placeListChildren(state, children, image.Rectangle{Max: dims.Size})
	return dims
}

// LayoutTrackedList keeps overlay anchors aligned with scrolling list items.
func LayoutTrackedList(ctx *frame.Context, gtx layout.Context, state *layout.List, count int, item layout.ListElement) layout.Dimensions {
	return layoutTrackedList(ctx, gtx, state, count, item)
}

func deduplicateListChildren(children []listChildPlacement) []listChildPlacement {
	write := 0
	for _, child := range children {
		if write > 0 && children[write-1].index == child.index {
			children[write-1] = child
			continue
		}
		children[write] = child
		write++
	}
	return children[:write]
}

func placeListChildren(state *layout.List, children []listChildPlacement, viewport image.Rectangle) {
	first := state.Position.First
	count := state.Position.Count
	maxCross := 0
	for index := first; index < first+count; index++ {
		if child, ok := findListChild(children, index); ok {
			maxCross = max(maxCross, state.Axis.Convert(child.dims.Size).Y)
		}
	}

	place := func(child listChildPlacement, main int) {
		size := state.Axis.Convert(child.dims.Size)
		cross := 0
		switch state.Alignment {
		case layout.End:
			cross = maxCross - size.Y
		case layout.Middle:
			cross = (maxCross - size.Y) / 2
		}
		child.placement.PlaceOffset(state.Axis.Convert(image.Pt(main, cross)))
		child.placement.ClipTo(viewport)
	}

	main := -state.Position.Offset
	if child, ok := findListChild(children, first-1); ok {
		leadingSize := state.Axis.Convert(child.dims.Size).X
		place(child, main-leadingSize-state.Gap)
	}
	for index := first; index < first+count; index++ {
		if index > first {
			main += state.Gap
		}
		child, ok := findListChild(children, index)
		if !ok {
			continue
		}
		place(child, main)
		main += state.Axis.Convert(child.dims.Size).X
	}
	if child, ok := findListChild(children, first+count); ok {
		place(child, main+state.Gap)
	}
}

func findListChild(children []listChildPlacement, index int) (listChildPlacement, bool) {
	position, ok := slices.BinarySearchFunc(children, index, func(child listChildPlacement, target int) int {
		return cmp.Compare(child.index, target)
	})
	if !ok {
		return listChildPlacement{}, false
	}
	return children[position], true
}
