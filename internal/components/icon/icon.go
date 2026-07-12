package icon

import (
	"image"
	"image/color"
	"sync"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
)

const defaultSize = unit.Dp(24)

type Widget struct {
	data     []byte
	size     unit.Dp
	color    color.NRGBA
	hasColor bool
}

type cacheKey struct {
	first  *byte
	length int
}

type renderer struct {
	mu   sync.Mutex
	icon *widget.Icon
}

var renderers sync.Map

func New(data []byte) Widget {
	return Widget{data: data}
}

func (w Widget) Size(dp float32) Widget {
	w.size = unit.Dp(max(dp, 0))
	return w
}

func (w Widget) Color(col color.NRGBA) Widget {
	w.color = col
	w.hasColor = true
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if w.data == nil {
		return layout.Dimensions{}
	}
	size := w.size
	if size == 0 {
		size = defaultSize
	}
	target := gtx.Dp(size)
	outerSize := gtx.Constraints.Constrain(image.Pt(target, target))
	diameter := min(outerSize.X, outerSize.Y)

	col := ctx.ForegroundColor()
	if w.hasColor {
		col = w.color
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	offset := image.Pt((outerSize.X-diameter)/2, (outerSize.Y-diameter)/2)
	stack := op.Offset(offset).Push(gtx.Ops)
	layoutIcon(w.data, iconGtx, col)
	stack.Pop()
	return layout.Dimensions{Size: outerSize}
}

func layoutIcon(data []byte, gtx layout.Context, col color.NRGBA) layout.Dimensions {
	if len(data) == 0 {
		return layout.Dimensions{}
	}
	key := cacheKey{first: &data[0], length: len(data)}
	valueRenderer, ok := renderers.Load(key)
	if !ok {
		resolved, err := widget.NewIcon(data)
		if err != nil {
			panic(err)
		}
		valueRenderer, _ = renderers.LoadOrStore(key, &renderer{icon: resolved})
	}
	cached := valueRenderer.(*renderer)
	cached.mu.Lock()
	defer cached.mu.Unlock()
	return cached.icon.Layout(gtx, col)
}
