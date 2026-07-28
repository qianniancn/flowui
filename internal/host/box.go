package host

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/interact"
	flowlayout "github.com/qianniancn/FlowUI/internal/layout"
	"github.com/qianniancn/FlowUI/internal/render"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
)

type Align = flowlayout.Align

// StyleFunc computes an instance-style override from the current interaction
// state. It is evaluated before the normal style cascade.
type StyleFunc func(*frame.Context, layout.Context, flowstyle.StyleState) flowstyle.Style

// BoxWidget is a layout host whose visual and box-model properties come only
// from Style (including StyleScope). Identity, interaction, and content
// alignment stay on the widget.
type BoxWidget struct {
	child       frame.Widget
	key         string
	label       string
	onClick     func()
	disabled    bool
	interactive bool
	customStyle flowstyle.Style
	styleFunc   StyleFunc
	hasAlign    bool
	align       Align
}

func Box(child frame.Widget) BoxWidget {
	return BoxWidget{child: child}
}

// Child returns the hosted widget.
func (b BoxWidget) Child() frame.Widget {
	return b.child
}

// Key sets the stable identity used by an interactive or transitioning Box.
func (b BoxWidget) Key(key string) BoxWidget {
	b.key = key
	return b
}

// Label sets the accessible name of an interactive Box.
func (b BoxWidget) Label(label string) BoxWidget {
	b.label = label
	return b
}

func (b BoxWidget) Disabled(disabled bool) BoxWidget {
	b.disabled = disabled
	return b
}

func (b BoxWidget) OnClick(fn func()) BoxWidget {
	b.onClick = fn
	b.interactive = true
	return b
}

// Align sets how the child is placed inside the box. Alignment is layout
// policy, not Style.
func (b BoxWidget) Align(align Align) BoxWidget {
	b.align = align
	b.hasAlign = true
	return b
}

// Style merges value into this box's instance style. Later Style calls override
// earlier properties via Join. Box geometry, paint, and overflow belong here.
func (b BoxWidget) Style(value flowstyle.Style) BoxWidget {
	b.customStyle = flowstyle.Join(b.customStyle, value)
	return b
}

// StyleFunc adds a state-aware instance-style override and makes the box
// interactive so hover, press, and focus state are available to fn.
func (b BoxWidget) StyleFunc(fn StyleFunc) BoxWidget {
	b.styleFunc = fn
	if fn != nil {
		b.interactive = true
	}
	return b
}

func (b BoxWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if b.interactive {
		click := interact.Begin(ctx, gtx, b.key, nil, !b.disabled, true, b.onClick)
		customStyle := b.customStyle
		if b.styleFunc != nil {
			customStyle = flowstyle.Join(customStyle, b.styleFunc(ctx, gtx, click.StyleState))
		}
		resolved := styleruntime.Resolve(
			ctx,
			gtx,
			click.Key,
			click.StyleState,
			flowstyle.Style{},
			flowstyle.Style{},
			flowstyle.Style{},
			customStyle,
		)
		return b.layoutInteractiveResolved(ctx, gtx, resolved, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
			return click.Layout(gtx, visual, b.label)
		})
	}
	runAssociationPreparer(ctx, b.child)
	resolved := styleruntime.ResolveStatic(
		ctx,
		flowstyle.StyleState{},
		flowstyle.Style{},
		flowstyle.Style{},
		flowstyle.Style{},
		b.customStyle,
	)
	if len(resolved.Transitions) != 0 {
		key := ""
		if b.key != "" {
			key = frame.ClaimKey(ctx, stateutil.KindStyle, b.key)
		}
		resolved = styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return b.layoutResolved(ctx, gtx, resolved)
}

// LayoutStyled resolves an interactive component root through the common
// style runtime, then renders it with Box semantics.
func LayoutStyled(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState, value flowstyle.Style, child frame.Widget) layout.Dimensions {
	resolved := styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		flowstyle.Style{},
		flowstyle.Style{},
		flowstyle.Style{},
		value,
	)
	return LayoutResolved(ctx, gtx, resolved, child)
}

// LayoutResolved renders a style that has already passed through the common
// cascade. Compound components use it to share Box semantics without resolving
// their default, variant, and instance layers twice.
func LayoutResolved(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle, child frame.Widget) layout.Dimensions {
	runAssociationPreparer(ctx, child)
	return (BoxWidget{child: child}).layoutResolved(ctx, gtx, resolved)
}

// LayoutInteractiveResolved keeps margin outside an interaction host while
// applying every other Box property to the host's visual and hit area.
func LayoutInteractiveResolved(
	ctx *frame.Context,
	gtx layout.Context,
	resolved flowstyle.ResolvedStyle,
	child frame.Widget,
	host func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return (BoxWidget{child: child}).layoutInteractiveResolved(ctx, gtx, resolved, host)
}

func (b BoxWidget) layoutInteractiveResolved(
	ctx *frame.Context,
	gtx layout.Context,
	resolved flowstyle.ResolvedStyle,
	host func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	runAssociationPreparer(ctx, b.child)
	visual := func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, resolved)
	}
	layoutHost := func(gtx layout.Context) layout.Dimensions {
		if host == nil {
			return visual(gtx)
		}
		return host(gtx, visual)
	}
	margin := boxInsets(resolved.Box, false)
	if hasInset(margin) {
		return layoutBoxInset(ctx, gtx, margin, layoutHost)
	}
	return layoutHost(gtx)
}

func (b BoxWidget) layoutResolved(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle) layout.Dimensions {
	layoutBox := func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, resolved)
	}
	margin := boxInsets(resolved.Box, false)
	if hasInset(margin) {
		return layoutBoxInset(ctx, gtx, margin, layoutBox)
	}
	return layoutBox(gtx)
}

func (b BoxWidget) layoutVisual(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle) layout.Dimensions {
	b.applyConstraints(&gtx, resolved.Box)
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return b.layoutContent(ctx, gtx, resolved)
	})
	content := macro.Stop()

	transform := boxTransform(gtx, resolved.Trans, dims.Size)
	placement.PlaceTransform(transform)
	opacityValue := boxOpacity(resolved.Paint)
	placement.SetOpacity(opacityValue)
	transformStack := op.Affine(transform).Push(gtx.Ops)
	defer transformStack.Pop()
	opacity := paint.PushOpacity(gtx.Ops, opacityValue)
	defer opacity.Pop()

	rect := image.Rectangle{Max: dims.Size}
	shape := boxRRect(gtx, resolved.Paint, rect)
	if resolved.Box != nil && resolved.Box.Cursor != nil {
		cursorClip := shape.Push(gtx.Ops)
		resolved.Box.Cursor.Add(gtx.Ops)
		cursorClip.Pop()
	}
	boxShadows(gtx, rect, resolved.Paint)
	if brush, ok := boxBrush(resolved.Paint); ok {
		render.DrawBrushRRect(gtx, shape, brush)
	}
	content.Add(gtx.Ops)
	boxBorder(gtx, shape, resolved.Paint)
	boxOutline(gtx, shape, resolved.Paint)
	return dims
}

func (b BoxWidget) layoutContent(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle) layout.Dimensions {

	layoutChild := func(gtx layout.Context) layout.Dimensions {
		return b.layoutChild(ctx, gtx, resolved)
	}
	if b.hasAlign {
		layoutChild = func(gtx layout.Context) layout.Dimensions {
			return layoutTrackedDirection(ctx, gtx, b.align.Direction(), func(gtx layout.Context) layout.Dimensions {
				return b.layoutChild(ctx, gtx, resolved)
			})
		}
	}
	padding := boxInsets(resolved.Box, true)
	if !hasInset(padding) {
		return b.layoutOverflow(ctx, gtx, resolved.Paint, resolved.Box, layoutChild)
	}
	return b.layoutOverflow(ctx, gtx, resolved.Paint, resolved.Box, func(gtx layout.Context) layout.Dimensions {
		return layoutBoxInset(ctx, gtx, padding, layoutChild)
	})
}

func layoutBoxInset(ctx *frame.Context, gtx layout.Context, inset layout.Inset, child layout.Widget) layout.Dimensions {
	horizontal := gtx.Dp(inset.Left) + gtx.Dp(inset.Right)
	vertical := gtx.Dp(inset.Top) + gtx.Dp(inset.Bottom)
	gtx.Constraints.Min.X = max(gtx.Constraints.Min.X-horizontal, 0)
	gtx.Constraints.Min.Y = max(gtx.Constraints.Min.Y-vertical, 0)
	return layoutTrackedInset(ctx, gtx, inset, child)
}

func (b BoxWidget) layoutChild(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle) layout.Dimensions {
	if b.child == nil {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	foreground, background := ctx.ForegroundColor(), ctx.BackgroundColor()
	if resolved.Text != nil {
		if color, ok := styleruntime.Color(resolved.Text.Color); ok {
			foreground = color
		}
	}
	if brush, ok := boxBrush(resolved.Paint); ok {
		if sampled := brush.ColorAt(.5); sampled.A != 0 {
			background = sampled
		}
	}
	restoreColors := frame.PushColors(ctx, foreground, background)
	defer restoreColors()
	if resolved.Text != nil {
		restoreStyle := frame.PushInheritedStyle(ctx, flowstyle.TextDeclaration(resolved.Text))
		defer restoreStyle()
	}
	return b.child.Layout(ctx, gtx)
}

func (b BoxWidget) layoutOverflow(ctx *frame.Context, gtx layout.Context, paintStyle *flowstyle.PaintStyle, boxStyle *flowstyle.BoxStyle, child layout.Widget) layout.Dimensions {
	if boxStyle == nil || boxStyle.Overflow == nil || *boxStyle.Overflow == flowstyle.OverflowVisible {
		return child(gtx)
	}
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child(gtx)
	})
	call := macro.Stop()
	dims.Size = gtx.Constraints.Constrain(dims.Size)
	placement.PlaceOffset(image.Point{})
	placement.ClipTo(image.Rectangle{Max: dims.Size})
	shape := boxRRect(gtx, paintStyle, image.Rectangle{Max: dims.Size})
	defer shape.Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
	return dims
}

func (b BoxWidget) applyConstraints(gtx *layout.Context, style *flowstyle.BoxStyle) {
	if style == nil {
		style = &flowstyle.BoxStyle{}
	}
	if style.MaxWidth != nil {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, max(gtx.Dp(*style.MaxWidth), 0))
		gtx.Constraints.Min.X = min(gtx.Constraints.Min.X, gtx.Constraints.Max.X)
	}
	if style.MaxHeight != nil {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, max(gtx.Dp(*style.MaxHeight), 0))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
	if style.MinWidth != nil {
		gtx.Constraints.Min.X = min(max(gtx.Constraints.Min.X, gtx.Dp(*style.MinWidth)), gtx.Constraints.Max.X)
	}
	if style.MinHeight != nil {
		gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(*style.MinHeight)), gtx.Constraints.Max.Y)
	}
	aspectBounds := gtx.Constraints
	if style.FillWidth != nil && *style.FillWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if style.FillHeight != nil && *style.FillHeight {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if style.Width != nil {
		width := min(max(max(gtx.Dp(*style.Width), 0), gtx.Constraints.Min.X), gtx.Constraints.Max.X)
		gtx.Constraints.Min.X = width
		gtx.Constraints.Max.X = width
	}
	if style.Height != nil {
		height := min(max(max(gtx.Dp(*style.Height), 0), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
		gtx.Constraints.Min.Y = height
		gtx.Constraints.Max.Y = height
	}
	applyAspectRatio(gtx, style, aspectBounds)
}

func applyAspectRatio(gtx *layout.Context, style *flowstyle.BoxStyle, bounds layout.Constraints) {
	if style.AspectRatio == nil || *style.AspectRatio <= 0 {
		return
	}
	widthFixed := style.Width != nil || style.FillWidth != nil && *style.FillWidth
	heightFixed := style.Height != nil || style.FillHeight != nil && *style.FillHeight
	if widthFixed == heightFixed {
		return
	}
	ratio := *style.AspectRatio
	if widthFixed {
		width := gtx.Constraints.Max.X
		height := int(float32(width)/ratio + .5)
		if height > bounds.Max.Y {
			height = bounds.Max.Y
			width = int(float32(height)*ratio + .5)
		}
		height = min(max(height, bounds.Min.Y), bounds.Max.Y)
		width = min(max(width, bounds.Min.X), bounds.Max.X)
		gtx.Constraints.Min = image.Pt(width, height)
		gtx.Constraints.Max = gtx.Constraints.Min
		return
	}
	height := gtx.Constraints.Max.Y
	width := int(float32(height)*ratio + .5)
	if width > bounds.Max.X {
		width = bounds.Max.X
		height = int(float32(width)/ratio + .5)
	}
	width = min(max(width, bounds.Min.X), bounds.Max.X)
	height = min(max(height, bounds.Min.Y), bounds.Max.Y)
	gtx.Constraints.Min = image.Pt(width, height)
	gtx.Constraints.Max = gtx.Constraints.Min
}

func boxInsets(style *flowstyle.BoxStyle, padding bool) layout.Inset {
	if style == nil {
		return layout.Inset{}
	}
	value := style.Margin
	if padding {
		value = style.Padding
	}
	if value == nil {
		return layout.Inset{}
	}
	return layout.Inset{
		Top:    value.Top,
		Right:  value.Right,
		Bottom: value.Bottom,
		Left:   value.Left,
	}
}

func hasInset(value layout.Inset) bool {
	return value.Top != 0 || value.Right != 0 || value.Bottom != 0 || value.Left != 0
}

func boxTransform(gtx layout.Context, style *flowstyle.TransformStyle, size image.Point) f32.Affine2D {
	transform := f32.AffineId()
	if style == nil {
		return transform
	}
	translateX, translateY := float32(0), float32(0)
	if style.TranslateX != nil {
		translateX = float32(gtx.Dp(*style.TranslateX))
	}
	if style.TranslateY != nil {
		translateY = float32(gtx.Dp(*style.TranslateY))
	}
	transform = transform.Offset(f32.Pt(translateX, translateY))
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	scaleX, scaleY := float32(1), float32(1)
	if style.ScaleX != nil {
		scaleX = *style.ScaleX
	}
	if style.ScaleY != nil {
		scaleY = *style.ScaleY
	}
	transform = transform.Scale(center, f32.Pt(scaleX, scaleY))
	if style.Rotate != nil {
		transform = transform.Rotate(center, *style.Rotate)
	}
	return transform
}

func boxOpacity(style *flowstyle.PaintStyle) float32 {
	if style == nil || style.Opacity == nil {
		return 1
	}
	return min(max(*style.Opacity, 0), 1)
}

func boxBrush(style *flowstyle.PaintStyle) (render.Brush, bool) {
	if style == nil {
		return render.Brush{}, false
	}
	return styleruntime.Brush(style.Background)
}

func boxRRect(gtx layout.Context, style *flowstyle.PaintStyle, rect image.Rectangle) clip.RRect {
	shape := clip.RRect{Rect: rect}
	if style == nil {
		return shape
	}
	limit := min(rect.Dx(), rect.Dy()) / 2
	toPixels := func(value unit.Dp) int { return min(max(gtx.Dp(value), 0), limit) }
	radii := boxCornerRadii(style)
	shape.NW = toPixels(radii.TopLeft)
	shape.NE = toPixels(radii.TopRight)
	shape.SE = toPixels(radii.BottomRight)
	shape.SW = toPixels(radii.BottomLeft)
	return shape
}

func boxCornerRadii(style *flowstyle.PaintStyle) flowstyle.CornerRadii {
	var radii flowstyle.CornerRadii
	if style == nil {
		return radii
	}
	if style.Radius != nil {
		radii = flowstyle.CornerRadii{
			TopLeft: *style.Radius, TopRight: *style.Radius,
			BottomRight: *style.Radius, BottomLeft: *style.Radius,
		}
	}
	if style.Radii != nil {
		radii = *style.Radii
	}
	return radii
}

func boxShadows(gtx layout.Context, rect image.Rectangle, style *flowstyle.PaintStyle) {
	if style == nil || rect.Empty() {
		return
	}
	radii := boxCornerRadii(style)
	shape := render.RoundedShadowCorners(radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft)
	for _, shadow := range style.Shadows {
		col, ok := styleruntime.Color(shadow.Color)
		if !ok {
			continue
		}
		render.DrawShadow(gtx, rect, shape, render.BoxShadow{
			OffsetX: float32(shadow.OffsetX),
			OffsetY: float32(shadow.OffsetY),
			Blur:    float32(shadow.Blur),
			Spread:  float32(shadow.Spread),
			Color:   col,
		})
	}
}

func boxBorder(gtx layout.Context, shape clip.RRect, style *flowstyle.PaintStyle) {
	if style == nil || style.Border == nil || style.Border.Width == nil {
		return
	}
	color, ok := styleruntime.Color(style.Border.Color)
	if !ok {
		return
	}
	width := max(gtx.Dp(*style.Border.Width), 0)
	render.DrawRRectBorder(gtx, shape, width, color)
}

func boxOutline(gtx layout.Context, shape clip.RRect, style *flowstyle.PaintStyle) {
	if style == nil || style.Outline == nil || style.Outline.Width <= 0 {
		return
	}
	col, ok := styleruntime.Color(style.Outline.Color)
	if !ok {
		return
	}
	width := max(gtx.Dp(style.Outline.Width), 1)
	gap := max(gtx.Dp(style.Outline.Offset), 0)
	expand := width + gap
	shape.Rect = shape.Rect.Inset(-expand)
	shape.NW += expand
	shape.NE += expand
	shape.SE += expand
	shape.SW += expand
	render.DrawRRectBorder(gtx, shape, width, col)
}
