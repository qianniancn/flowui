package layout

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// StyleFunc computes an instance-style override from the current interaction
// state. It is evaluated before the normal style cascade.
type StyleFunc func(*Context, layout.Context, style.State) style.Style

// BoxWidget is a layout host whose visual and box-model properties come only
// from Style (including style scopes). Identity, interaction, and content
// alignment stay on the widget.
type BoxWidget struct {
	child       Widget
	key         string
	label       string
	onClick     func()
	disabled    bool
	interactive bool
	customStyle style.Style
	styleFunc   StyleFunc
	hasAlign    bool
	align       Align
}

// Box returns a host for child. Geometry comes only from Style.
func Box(child Widget) BoxWidget {
	return BoxWidget{child: child}
}

// Child returns the hosted widget.
func (b BoxWidget) Child() Widget {
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

// Style is the legacy alias for MergeStyle. Every declaration is merged in
// order, so one call can carry several styles (or one chained style):
//
//	layout.Box(content).Style(style.New().FillWidth().MaxWidth(760).Padding(24))
//	layout.Box(content).Style(style.New().FillWidth(), style.New().Padding(24))
//
// Use ReplaceStyle when the whole instance layer should be replaced.
func (b BoxWidget) Style(values ...style.Style) BoxWidget {
	return b.MergeStyle(values...)
}

// MergeStyle merges every value into this box's instance style. Later
// properties override earlier properties while unspecified properties are
// preserved.
func (b BoxWidget) MergeStyle(values ...style.Style) BoxWidget {
	b.customStyle = style.Join(append([]style.Style{b.customStyle}, values...)...)
	return b
}

// ReplaceStyle replaces the box's complete instance style layer with the given
// declarations joined in order. This is useful when a reusable style should not
// inherit an earlier Box style.
func (b BoxWidget) ReplaceStyle(values ...style.Style) BoxWidget {
	b.customStyle = style.Join(values...)
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

func (b BoxWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	if b.interactive {
		click := core.Begin(ctx, gtx, b.key, nil, !b.disabled, true, b.onClick)
		customStyle := b.customStyle
		if b.styleFunc != nil {
			customStyle = style.Join(customStyle, b.styleFunc(ctx, gtx, click.State))
		}
		resolved := computed.Resolve(
			ctx,
			gtx,
			click.Key,
			click.State,
			style.Style{},
			style.Style{},
			style.Style{},
			customStyle,
		)
		return b.layoutInteractiveResolved(ctx, gtx, resolved, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
			return click.Layout(gtx, visual, b.label)
		}, nil, nil)
	}
	runAssociationPreparer(ctx, b.child)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		style.Style{},
		style.Style{},
		style.Style{},
		b.customStyle,
	)
	if len(resolved.Transitions) != 0 {
		key := ""
		if b.key != "" {
			key = core.ClaimKey(ctx, core.KindStyle, b.key)
		}
		resolved = computed.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return b.layoutResolved(ctx, gtx, resolved)
}

// Styled resolves an interactive component root through
// internal/computed, then renders it with Box semantics.
func Styled(ctx *Context, gtx layout.Context, key string, state style.State, value style.Style, child Widget) layout.Dimensions {
	resolved := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		style.Style{},
		style.Style{},
		style.Style{},
		value,
	)
	return Resolved(ctx, gtx, resolved, child)
}

// Resolved renders a style that has already passed through the common
// cascade. Compound components use it to share Box semantics without resolving
// their default, variant, and instance layers twice.
func Resolved(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle, child Widget) layout.Dimensions {
	runSingleAssociationPreparer(ctx, child)
	return (BoxWidget{child: child}).layoutResolved(ctx, gtx, resolved)
}

// InteractiveResolved keeps margin outside an interaction host while
// applying every other Box property to the host's visual and hit area.
func InteractiveResolved(
	ctx *Context,
	gtx layout.Context,
	resolved style.ResolvedStyle,
	child Widget,
	host func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return (BoxWidget{child: child}).layoutInteractiveResolved(ctx, gtx, resolved, host, nil, nil)
}

// InteractiveResolvedWithDraw is the visual extension point used by
// style packages. draw receives the default root visual and may add or replace
// its paint while interaction and the box model remain component-owned.
func InteractiveResolvedWithDraw(
	ctx *Context,
	gtx layout.Context,
	resolved style.ResolvedStyle,
	child Widget,
	host func(layout.Context, layout.Widget) layout.Dimensions,
	draw func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return (BoxWidget{child: child}).layoutInteractiveResolved(ctx, gtx, resolved, host, draw, nil)
}

// InteractiveResolvedWithDrawAndOverlay renders a visual draw callback
// inside the box and a post-layout overlay outside the interaction clip.
func InteractiveResolvedWithDrawAndOverlay(
	ctx *Context,
	gtx layout.Context,
	resolved style.ResolvedStyle,
	child Widget,
	host func(layout.Context, layout.Widget) layout.Dimensions,
	draw func(layout.Context, layout.Widget) layout.Dimensions,
	overlay func(layout.Context, layout.Dimensions),
) layout.Dimensions {
	return (BoxWidget{child: child}).layoutInteractiveResolved(ctx, gtx, resolved, host, draw, overlay)
}

func (b BoxWidget) layoutInteractiveResolved(
	ctx *Context,
	gtx layout.Context,
	resolved style.ResolvedStyle,
	host func(layout.Context, layout.Widget) layout.Dimensions,
	draw func(layout.Context, layout.Widget) layout.Dimensions,
	overlay func(layout.Context, layout.Dimensions),
) layout.Dimensions {
	runAssociationPreparer(ctx, b.child)
	transformStyle := resolved.Trans
	visualStyle := resolved
	visualStyle.Trans = nil
	visual := func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, visualStyle)
	}
	if draw != nil {
		defaultVisual := visual
		visual = func(gtx layout.Context) layout.Dimensions {
			return draw(gtx, defaultVisual)
		}
	}
	layoutHost := func(gtx layout.Context) layout.Dimensions {
		var dims layout.Dimensions
		if host == nil {
			dims = visual(gtx)
		} else {
			dims = host(gtx, visual)
		}
		if overlay != nil {
			overlay(gtx, dims)
		}
		return dims
	}
	if !boxTransformIsIdentity(transformStyle) {
		untransformedHost := layoutHost
		layoutHost = func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return untransformedHost(gtx)
			})
			call := macro.Stop()
			transform := boxTransform(gtx, transformStyle, dims.Size)
			placement.PlaceTransform(transform)
			stack := op.Affine(transform).Push(gtx.Ops)
			call.Add(gtx.Ops)
			stack.Pop()
			return dims
		}
	}
	margin := boxInsets(resolved.Box, false)
	if hasInset(margin) {
		return layoutBoxInset(ctx, gtx, margin, layoutHost)
	}
	return layoutHost(gtx)
}

func (b BoxWidget) layoutResolved(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {
	layoutBox := func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, resolved)
	}
	margin := boxInsets(resolved.Box, false)
	if hasInset(margin) {
		return layoutBoxInset(ctx, gtx, margin, layoutBox)
	}
	return layoutBox(gtx)
}

func (b BoxWidget) layoutVisual(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {
	b.applyConstraints(&gtx, resolved.Box)
	if core.CollectingVisualOverflow(ctx) {
		core.ReportVisualOverflow(ctx, boxVisualOutset(gtx, resolved.Paint))
	}
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
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
		draw.FillRRect(gtx, shape, brush)
	}
	content.Add(gtx.Ops)
	boxBorder(gtx, shape, resolved.Paint)
	boxOutline(gtx, shape, resolved.Paint)
	return dims
}

func (b BoxWidget) layoutContent(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {

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

func layoutBoxInset(ctx *Context, gtx layout.Context, inset layout.Inset, child layout.Widget) layout.Dimensions {
	horizontal := gtx.Dp(inset.Left) + gtx.Dp(inset.Right)
	vertical := gtx.Dp(inset.Top) + gtx.Dp(inset.Bottom)
	gtx.Constraints.Min.X = max(gtx.Constraints.Min.X-horizontal, 0)
	gtx.Constraints.Min.Y = max(gtx.Constraints.Min.Y-vertical, 0)
	return layoutTrackedInset(ctx, gtx, inset, child)
}

func (b BoxWidget) layoutChild(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {
	if b.child == nil {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	foreground, background := ctx.ForegroundColor(), ctx.BackgroundColor()
	if resolved.Text != nil {
		if color, ok := computed.Color(resolved.Text.Color); ok {
			foreground = color
		}
	}
	if brush, ok := boxBrush(resolved.Paint); ok {
		if sampled := brush.ColorAt(.5); sampled.A != 0 {
			background = sampled
		}
	}
	restoreColors := core.PushColors(ctx, foreground, background)
	defer restoreColors()
	if resolved.Text != nil {
		restoreStyle := core.PushInheritedStyle(ctx, style.TextDeclaration(resolved.Text))
		defer restoreStyle()
	}
	return b.child.Layout(ctx, gtx)
}

func (b BoxWidget) layoutOverflow(ctx *Context, gtx layout.Context, paintStyle *style.PaintStyle, boxStyle *style.BoxStyle, child layout.Widget) layout.Dimensions {
	if boxStyle == nil || boxStyle.Overflow == nil || *boxStyle.Overflow == style.OverflowVisible {
		return child(gtx)
	}
	// This box owns the clip, so visual overflow from descendants must not be
	// reserved again by an ancestor viewport.
	restoreCollector := core.PushVisualOverflowCollector(ctx, new(core.VisualOverflowCollector))
	defer restoreCollector()
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
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

func (b BoxWidget) applyConstraints(gtx *layout.Context, box *style.BoxStyle) {
	if box == nil {
		box = &style.BoxStyle{}
	}
	if box.MaxWidth != nil {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, max(gtx.Dp(*box.MaxWidth), 0))
		gtx.Constraints.Min.X = min(gtx.Constraints.Min.X, gtx.Constraints.Max.X)
	}
	if box.MaxHeight != nil {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, max(gtx.Dp(*box.MaxHeight), 0))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
	if box.MinWidth != nil {
		gtx.Constraints.Min.X = min(max(gtx.Constraints.Min.X, gtx.Dp(*box.MinWidth)), gtx.Constraints.Max.X)
	}
	if box.MinHeight != nil {
		gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(*box.MinHeight)), gtx.Constraints.Max.Y)
	}
	aspectBounds := gtx.Constraints
	if box.FillWidth != nil && *box.FillWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if box.FillHeight != nil && *box.FillHeight {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if box.Width != nil {
		width := min(max(max(gtx.Dp(*box.Width), 0), gtx.Constraints.Min.X), gtx.Constraints.Max.X)
		gtx.Constraints.Min.X = width
		gtx.Constraints.Max.X = width
	}
	if box.Height != nil {
		height := min(max(max(gtx.Dp(*box.Height), 0), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
		gtx.Constraints.Min.Y = height
		gtx.Constraints.Max.Y = height
	}
	applyAspectRatio(gtx, box, aspectBounds)
}

func applyAspectRatio(gtx *layout.Context, box *style.BoxStyle, bounds layout.Constraints) {
	if box.AspectRatio == nil || *box.AspectRatio <= 0 {
		return
	}
	widthFixed := box.Width != nil || box.FillWidth != nil && *box.FillWidth
	heightFixed := box.Height != nil || box.FillHeight != nil && *box.FillHeight
	if widthFixed == heightFixed {
		return
	}
	ratio := *box.AspectRatio
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

func boxInsets(box *style.BoxStyle, padding bool) layout.Inset {
	if box == nil {
		return layout.Inset{}
	}
	value := box.Margin
	if padding {
		value = box.Padding
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

func boxTransform(gtx layout.Context, transform *style.TransformStyle, size image.Point) f32.Affine2D {
	matrix := f32.AffineId()
	if transform == nil {
		return matrix
	}
	translateX, translateY := float32(0), float32(0)
	if transform.TranslateX != nil {
		translateX = float32(gtx.Dp(*transform.TranslateX))
	}
	if transform.TranslateY != nil {
		translateY = float32(gtx.Dp(*transform.TranslateY))
	}
	matrix = matrix.Offset(f32.Pt(translateX, translateY))
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	scaleX, scaleY := float32(1), float32(1)
	if transform.ScaleX != nil {
		scaleX = *transform.ScaleX
	}
	if transform.ScaleY != nil {
		scaleY = *transform.ScaleY
	}
	matrix = matrix.Scale(center, f32.Pt(scaleX, scaleY))
	if transform.Rotate != nil {
		matrix = matrix.Rotate(center, *transform.Rotate)
	}
	return matrix
}

func boxTransformIsIdentity(transform *style.TransformStyle) bool {
	if transform == nil {
		return true
	}
	return (transform.TranslateX == nil || *transform.TranslateX == 0) &&
		(transform.TranslateY == nil || *transform.TranslateY == 0) &&
		(transform.ScaleX == nil || *transform.ScaleX == 1) &&
		(transform.ScaleY == nil || *transform.ScaleY == 1) &&
		(transform.Rotate == nil || *transform.Rotate == 0)
}

func boxOpacity(paint *style.PaintStyle) float32 {
	if paint == nil || paint.Opacity == nil {
		return 1
	}
	return min(max(*paint.Opacity, 0), 1)
}

func boxBrush(paint *style.PaintStyle) (draw.Brush, bool) {
	if paint == nil {
		return draw.Brush{}, false
	}
	return computed.Brush(paint.Background)
}

func boxRRect(gtx layout.Context, paint *style.PaintStyle, rect image.Rectangle) clip.RRect {
	shape := clip.RRect{Rect: rect}
	if paint == nil {
		return shape
	}
	limit := min(rect.Dx(), rect.Dy()) / 2
	toPixels := func(value unit.Dp) int { return min(max(gtx.Dp(value), 0), limit) }
	radii := boxCornerRadii(paint)
	shape.NW = toPixels(radii.TopLeft)
	shape.NE = toPixels(radii.TopRight)
	shape.SE = toPixels(radii.BottomRight)
	shape.SW = toPixels(radii.BottomLeft)
	return shape
}

func boxCornerRadii(paint *style.PaintStyle) style.CornerRadii {
	var radii style.CornerRadii
	if paint == nil {
		return radii
	}
	if paint.Radius != nil {
		radii = style.CornerRadii{
			TopLeft: *paint.Radius, TopRight: *paint.Radius,
			BottomRight: *paint.Radius, BottomLeft: *paint.Radius,
		}
	}
	if paint.Radii != nil {
		radii = *paint.Radii
	}
	return radii
}

func boxShadows(gtx layout.Context, rect image.Rectangle, paint *style.PaintStyle) {
	if paint == nil || rect.Empty() {
		return
	}
	radii := boxCornerRadii(paint)
	shape := draw.RoundedShadowCorners(radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft)
	for _, shadow := range paint.Shadows {
		col, ok := computed.Color(shadow.Color)
		if !ok {
			continue
		}
		draw.Shadow(gtx, rect, shape, draw.BoxShadow{
			OffsetX: float32(shadow.OffsetX),
			OffsetY: float32(shadow.OffsetY),
			Blur:    float32(shadow.Blur),
			Spread:  float32(shadow.Spread),
			Color:   col,
		})
	}
}

func boxVisualOutset(gtx layout.Context, paint *style.PaintStyle) core.VisualOutset {
	if paint == nil {
		return core.VisualOutset{}
	}
	result := core.VisualOutset{}
	if paint.VisualOutset != nil {
		result = result.Max(core.VisualOutset{
			Top:    max(gtx.Dp(paint.VisualOutset.Top), 0),
			Right:  max(gtx.Dp(paint.VisualOutset.Right), 0),
			Bottom: max(gtx.Dp(paint.VisualOutset.Bottom), 0),
			Left:   max(gtx.Dp(paint.VisualOutset.Left), 0),
		})
	}
	for _, shadow := range paint.Shadows {
		if shadow.Transient {
			continue
		}
		col, ok := computed.Color(shadow.Color)
		if !ok || col.A == 0 {
			continue
		}
		blur := gtx.Dp(shadow.Blur)
		if blur < 0 {
			continue
		}
		padX := draw.ShadowRasterPadding(blur, gtx.Dp(shadow.Spread), gtx.Dp(shadow.OffsetX))
		padY := draw.ShadowRasterPadding(blur, gtx.Dp(shadow.Spread), gtx.Dp(shadow.OffsetY))
		result = result.Max(core.VisualOutset{
			Top:    padY,
			Right:  padX,
			Bottom: padY,
			Left:   padX,
		})
	}
	if paint.Outline != nil {
		width := gtx.Dp(paint.Outline.Width)
		if width > 0 {
			expand := saturatedAdd(width, max(gtx.Dp(paint.Outline.Offset), 0))
			result = result.Max(core.VisualOutset{Top: expand, Right: expand, Bottom: expand, Left: expand})
		}
	}
	return result
}

func saturatedAdd(values ...int) int {
	limit := int(^uint(0) >> 1)
	result := 0
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if result > limit-value {
			return limit
		}
		result += value
	}
	return result
}

func boxBorder(gtx layout.Context, shape clip.RRect, paint *style.PaintStyle) {
	if paint == nil {
		return
	}
	if paint.Border != nil && paint.Border.Width != nil {
		color, ok := computed.Color(paint.Border.Color)
		if ok {
			width := max(gtx.Dp(*paint.Border.Width), 0)
			on, off := style.LineSolid.DashPattern()
			if paint.Border.LineStyle != nil {
				on, off = paint.Border.LineStyle.DashPattern()
			}
			if on > 0 {
				draw.DashedRRectBorder(gtx, shape, width, max(gtx.Dp(on), 1), gtx.Dp(off), color)
			} else {
				draw.RRectBorder(gtx, shape, width, color)
			}
		}
	}
	if paint.BorderBottom != nil && paint.BorderBottom.Width != nil {
		color, ok := computed.Color(paint.BorderBottom.Color)
		if ok {
			clipStack := shape.Push(gtx.Ops)
			draw.BottomBorder(gtx, shape.Rect, max(gtx.Dp(*paint.BorderBottom.Width), 0), color)
			clipStack.Pop()
		}
	}
}

func boxOutline(gtx layout.Context, shape clip.RRect, paint *style.PaintStyle) {
	if paint == nil || paint.Outline == nil || paint.Outline.Width <= 0 {
		return
	}
	col, ok := computed.Color(paint.Outline.Color)
	if !ok {
		return
	}
	width := max(gtx.Dp(paint.Outline.Width), 1)
	gap := max(gtx.Dp(paint.Outline.Offset), 0)
	expand := width + gap
	shape.Rect = shape.Rect.Inset(-expand)
	shape.NW += expand
	shape.NE += expand
	shape.SE += expand
	shape.SW += expand
	draw.RRectBorder(gtx, shape, width, col)
}
