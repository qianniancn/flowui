package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type StackWidget struct {
	layers []StackLayer
	align  Align
}

type StackLayer struct {
	child    frame.Widget
	align    Align
	hasAlign bool
	overlay  bool
	expanded bool
}

func Stack(layers ...StackLayer) StackWidget {
	return StackWidget{layers: layers}
}

func Stacked(child frame.Widget) StackLayer {
	return StackLayer{child: child}
}

func Overlay(child frame.Widget) StackLayer {
	return StackLayer{child: child, overlay: true}
}

func (s StackWidget) Align(align Align) StackWidget {
	s.align = align
	return s
}

func (l StackLayer) Align(align Align) StackLayer {
	l.align = align
	l.hasAlign = true
	return l
}

func (l StackLayer) Expanded() StackLayer {
	l.expanded = true
	return l
}

func (s StackWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareStackFieldAssociations(ctx, s.layers)
	layers := make([]stackLayerLayout, len(s.layers))
	stackSize := image.Point{}
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}

	for i, layer := range s.layers {
		if layer.overlay || layer.expanded {
			continue
		}
		layers[i] = layoutStackLayer(ctx, childGtx, layer)
		stackSize.X = max(stackSize.X, layers[i].dims.Size.X)
		stackSize.Y = max(stackSize.Y, layers[i].dims.Size.Y)
	}

	size := gtx.Constraints.Constrain(stackSize)
	for i, layer := range s.layers {
		if !layer.overlay && !layer.expanded {
			continue
		}
		layerGtx := childGtx
		if layer.expanded {
			layerGtx.Constraints.Min = size
		}
		layers[i] = layoutStackLayer(ctx, layerGtx, layer)
	}

	var baseline int
	for i, layer := range layers {
		align := s.layers[i].align
		if !s.layers[i].hasAlign {
			align = s.align
		}
		pos := align.Position(layer.dims.Size, size)
		trans := op.Offset(pos).Push(gtx.Ops)
		layer.call.Add(gtx.Ops)
		trans.Pop()
		if baseline == 0 && layer.dims.Baseline != 0 {
			baseline = layer.dims.Baseline + size.Y - layer.dims.Size.Y - pos.Y
		}
	}
	return layout.Dimensions{
		Size:     size,
		Baseline: baseline,
	}
}

func layoutStackLayer(ctx *frame.Context, gtx layout.Context, layer StackLayer) stackLayerLayout {
	macro := op.Record(gtx.Ops)
	dims := layer.child.Layout(ctx, gtx)
	call := macro.Stop()
	return stackLayerLayout{
		call: call,
		dims: dims,
	}
}

type stackLayerLayout struct {
	call op.CallOp
	dims layout.Dimensions
}
