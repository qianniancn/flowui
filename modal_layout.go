package flowui

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

func (m ModalWidget) layoutOverlay(ctx *Context, gtx layout.Context, state *modalState, progress float32) layout.Dimensions {
	size := modalOverlaySize(gtx)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(size)

	style := modalStyleFor(ctx.Theme, m.backdrop, m.size)
	drawModalBackdrop(gtx, size, style, progress)
	m.layoutFocusTrap(ctx, gtx, state)

	dialogGtx := gtx
	dialogGtx.Constraints = m.dialogConstraints(ctx, gtx, size)
	macro := op.Record(gtx.Ops)
	dialogDims := m.layoutDialogFrame(ctx, dialogGtx, state)
	dialogCall := macro.Stop()

	dialogPos := m.dialogPosition(ctx, gtx, size, dialogDims.Size)
	dialogRect := image.Rectangle{
		Min: dialogPos,
		Max: dialogPos.Add(dialogDims.Size),
	}
	motion := m.dialogMotion(ctx, gtx, dialogRect, progress, state.opening())
	m.layoutDismissAreas(gtx, state, size, modalMotionBounds(dialogRect, motion.transform))
	transform := op.Affine(motion.transform).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, motion.opacity)
	offset := op.Offset(dialogPos).Push(gtx.Ops)
	m.layoutDialogBlocker(gtx, state, dialogDims.Size)
	dialogCall.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	transform.Pop()
	m.layoutFocusEnd(gtx, state)

	return layout.Dimensions{Size: size}
}

func (m ModalWidget) layoutFocusTrap(ctx *Context, gtx layout.Context, state *modalState) {
	m.redirectFocusedBoundary(ctx, gtx, &state.focusStart, state.tabFocusTag(m.showCloseButton()))
	m.redirectFocusedBoundary(ctx, gtx, &state.focusEnd, state.endFocusTag())
	for {
		_, ok := gtx.Event(key.FocusFilter{Target: &state.focusTarget})
		if !ok {
			break
		}
	}
	m.layoutFocusTag(gtx, &state.focusStart)
	m.layoutFocusTag(gtx, &state.focusTarget)
}

func (m ModalWidget) redirectFocusedBoundary(ctx *Context, gtx layout.Context, target event.Tag, redirect event.Tag) {
	for {
		e, ok := gtx.Event(key.FocusFilter{Target: target})
		if !ok {
			break
		}
		event, ok := e.(key.FocusEvent)
		if ok && event.Focus {
			ctx.requestFocus(redirect)
		}
	}
}

func (m ModalWidget) layoutFocusEnd(gtx layout.Context, state *modalState) {
	m.layoutFocusTag(gtx, &state.focusEnd)
}

func (m ModalWidget) layoutFocusTag(gtx layout.Context, target *widget.Clickable) {
	stack := clip.Rect{Max: image.Pt(1, 1)}.Push(gtx.Ops)
	target.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(1, 1)}
	})
	stack.Pop()
}

func modalOverlaySize(gtx layout.Context) image.Point {
	size := gtx.Constraints.Max
	if size.X < 0 {
		size.X = 0
	}
	if size.Y < 0 {
		size.Y = 0
	}
	return size
}

func (m ModalWidget) dialogConstraints(ctx *Context, gtx layout.Context, size image.Point) layout.Constraints {
	margin := m.dialogMargin(ctx, gtx, size)
	available := image.Pt(max(size.X-margin*2, 0), max(size.Y-margin*2, 0))
	if m.size == ModalFull {
		return layout.Exact(size)
	}
	if m.size == ModalCover {
		return layout.Exact(available)
	}

	width := min(modalWidth(ctx, gtx, m.size), available.X)
	if width <= 0 {
		width = available.X
	}
	return layout.Constraints{
		Min: image.Pt(width, 0),
		Max: image.Pt(width, available.Y),
	}
}

func modalWidth(ctx *Context, gtx layout.Context, size ModalSize) int {
	theme := ctx.Theme.Components.Modal
	switch size {
	case ModalXSmall:
		return gtx.Dp(theme.XSmallWidth)
	case ModalSmall:
		return gtx.Dp(theme.SmallWidth)
	case ModalLarge:
		return gtx.Dp(theme.LargeWidth)
	default:
		return gtx.Dp(theme.MediumWidth)
	}
}

func (m ModalWidget) dialogMargin(ctx *Context, gtx layout.Context, size image.Point) int {
	theme := ctx.Theme.Components.Modal
	margin := gtx.Dp(theme.Margin)
	if size.X >= gtx.Dp(theme.DesktopBreakpoint) {
		margin = gtx.Dp(theme.DesktopMargin)
	}
	if m.size == ModalFull {
		return 0
	}
	return min(max(margin, 0), min(size.X, size.Y)/2)
}

func (m ModalWidget) dialogPosition(ctx *Context, gtx layout.Context, overlay, dialog image.Point) image.Point {
	margin := m.dialogMargin(ctx, gtx, overlay)
	x := max((overlay.X-dialog.X)/2, 0)
	placement := m.resolvedPlacement(ctx, gtx, overlay)
	var y int
	switch placement {
	case ModalTop:
		y = margin
	case ModalBottom:
		y = overlay.Y - margin - dialog.Y
	default:
		y = (overlay.Y - dialog.Y) / 2
	}
	if y < margin && m.size != ModalFull {
		y = margin
	}
	if y+dialog.Y > overlay.Y-margin && m.size != ModalFull {
		y = max(overlay.Y-margin-dialog.Y, 0)
	}
	return image.Pt(x, max(y, 0))
}

func (m ModalWidget) resolvedPlacement(ctx *Context, gtx layout.Context, size image.Point) ModalPlacement {
	if m.placement != ModalAuto {
		return m.placement
	}
	if size.X < gtx.Dp(ctx.Theme.Components.Modal.DesktopBreakpoint) {
		return ModalBottom
	}
	return ModalCenter
}

type modalDialogMotion struct {
	transform f32.Affine2D
	opacity   float32
}

func (m ModalWidget) dialogMotion(ctx *Context, gtx layout.Context, rect image.Rectangle, progress float32, opening bool) modalDialogMotion {
	theme := ctx.Theme.Components.Modal
	transform := f32.AffineId()
	center := f32.Pt(float32(rect.Min.X)+float32(rect.Dx())/2, float32(rect.Min.Y)+float32(rect.Dy())/2)
	scale := float32(1)

	switch m.resolvedAnimation() {
	case ModalAnimationFade:
	case ModalAnimationSlideDown:
		distance := float32(gtx.Dp(theme.AnimationDistance))
		transform = transform.Offset(f32.Pt(0, -distance*(1-progress)))
	case ModalAnimationSlideUp:
		distance := float32(gtx.Dp(theme.AnimationDistance))
		transform = transform.Offset(f32.Pt(0, distance*(1-progress)))
	case ModalAnimationBounceScale:
		if opening {
			scale = modalDialogBounceScale(theme.AnimationScale, theme.AnimationBounceScale, progress)
		} else {
			scale = modalDialogScale(theme.AnimationScale, progress)
		}
	case ModalAnimationZoomOut:
		if opening {
			scale = modalDialogZoomOutScale(progress)
		} else {
			scale = modalDialogScale(theme.AnimationScale, progress)
		}
	case ModalAnimationPop:
		if opening {
			scale = modalDialogPopScale(progress)
		} else {
			scale = modalDialogScale(theme.AnimationScale, progress)
		}
	default:
		scale = modalDialogScale(theme.AnimationScale, progress)
	}
	if scale != 1 {
		transform = transform.Scale(center, f32.Pt(scale, scale))
	}
	return modalDialogMotion{
		transform: transform,
		opacity:   progress,
	}
}

func modalMotionBounds(rect image.Rectangle, transform f32.Affine2D) image.Rectangle {
	if rect.Empty() {
		return image.Rectangle{}
	}
	points := [4]f32.Point{
		transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Min.Y))),
		transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Min.Y))),
		transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Max.Y))),
		transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Max.Y))),
	}
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, point := range points[1:] {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	return image.Rect(
		int(math.Floor(float64(minX))),
		int(math.Floor(float64(minY))),
		int(math.Ceil(float64(maxX))),
		int(math.Ceil(float64(maxY))),
	)
}

func modalDialogScale(startScale, progress float32) float32 {
	if startScale <= 0 || startScale > 1 {
		startScale = 0.95
	}
	return startScale + (1-startScale)*progress
}

func modalDialogBounceScale(startScale, overshoot, progress float32) float32 {
	if startScale <= 0 || startScale > 1 {
		startScale = 0.95
	}
	if overshoot <= 1 {
		overshoot = 1.035
	}
	progress = clampProgress(progress)
	switch {
	case progress < 0.65:
		return lerp(startScale, overshoot, modalEaseOutCubic(progress/0.65))
	case progress < 0.86:
		return lerp(overshoot, 0.992, animationEase((progress-0.65)/0.21))
	default:
		return lerp(0.992, 1, modalEaseOutCubic((progress-0.86)/0.14))
	}
}

func modalDialogZoomOutScale(progress float32) float32 {
	return lerp(1.04, 1, modalEaseOutCubic(progress))
}

func modalDialogPopScale(progress float32) float32 {
	progress = clampProgress(progress)
	switch {
	case progress < 0.58:
		return lerp(0.92, 1.025, modalEaseOutCubic(progress/0.58))
	case progress < 0.82:
		return lerp(1.025, 0.997, animationEase((progress-0.58)/0.24))
	default:
		return lerp(0.997, 1, modalEaseOutCubic((progress-0.82)/0.18))
	}
}

func modalEaseOutCubic(progress float32) float32 {
	progress = clampProgress(progress)
	inverse := 1 - progress
	return 1 - inverse*inverse*inverse
}

func clampProgress(progress float32) float32 {
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

func (m ModalWidget) layoutDismissAreas(gtx layout.Context, state *modalState, overlay image.Point, dialog image.Rectangle) {
	areas := overlayDismissRects(image.Rectangle{Max: overlay}, dialog)
	for i, area := range areas {
		if area.Empty() {
			continue
		}
		areaGtx := gtx
		areaGtx.Constraints = layout.Exact(area.Size())
		stack := op.Offset(area.Min).Push(gtx.Ops)
		state.dismiss[i].Layout(areaGtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: area.Size()}
		})
		stack.Pop()
	}
}

func (m ModalWidget) layoutDialogFrame(ctx *Context, gtx layout.Context, state *modalState) layout.Dimensions {
	return m.layoutDialog(ctx, gtx, state)
}

func (m ModalWidget) layoutDialog(ctx *Context, gtx layout.Context, state *modalState) layout.Dimensions {
	return m.layoutDialogSurface(ctx, gtx, state)
}

func (m ModalWidget) layoutDialogBlocker(gtx layout.Context, state *modalState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}

func (m ModalWidget) layoutDialogSurface(ctx *Context, gtx layout.Context, state *modalState) layout.Dimensions {
	theme := ctx.Theme.Components.Modal
	padding := gtx.Dp(theme.Padding)
	contentGtx := gtx
	contentGtx.Constraints.Min = shrinkPoint(contentGtx.Constraints.Min, padding*2)
	contentGtx.Constraints.Max = shrinkPoint(contentGtx.Constraints.Max, padding*2)

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		restore := ctx.pushForeground(ctx.Theme.Palette.overlayForegroundColor())
		defer restore()
		contentDims = m.layoutDialogContent(ctx, contentGtx, state)
	}()
	contentCall := macro.Stop()

	size := contentDims.Size.Add(image.Pt(padding*2, padding*2))
	size = gtx.Constraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := modalDialogRadius(gtx, ctx.Theme, m.size, size)
	drawModalSurface(gtx, ctx.Theme, rect, radius, m.size)

	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	contentCall.Add(gtx.Ops)
	contentOffset.Pop()
	clipStack.Pop()

	if m.showCloseButton() {
		m.layoutCloseButton(ctx, gtx, state, size)
	}
	return layout.Dimensions{Size: size}
}

func (m ModalWidget) layoutDialogContent(ctx *Context, gtx layout.Context, state *modalState) layout.Dimensions {
	if m.scroll == ModalScrollOutside {
		state.outsideList.Axis = layout.Vertical
		state.outsideList.ScrollAnyAxis = false
		return state.outsideList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
			return m.layoutDialogSections(ctx, gtx, state)
		})
	}
	return m.layoutDialogSections(ctx, gtx, state)
}

func (m ModalWidget) layoutDialogSections(ctx *Context, gtx layout.Context, state *modalState) layout.Dimensions {
	sectionGtx := gtx
	sectionGtx.Constraints.Min = image.Point{}

	headerCall, headerDims := m.recordHeader(ctx, sectionGtx)
	footerCall, footerDims := m.recordFooter(ctx, sectionGtx)
	bodyGap, footerGap := m.dialogGaps(ctx, sectionGtx, headerDims.Size.Y > 0, footerDims.Size.Y > 0)

	bodyGtx := sectionGtx
	maxBodyY := max(gtx.Constraints.Max.Y-headerDims.Size.Y-footerDims.Size.Y-bodyGap-footerGap, 0)
	bodyGtx.Constraints.Max.Y = maxBodyY
	bodyCall, bodyDims := m.recordBody(ctx, bodyGtx, state)

	width := max(headerDims.Size.X, max(bodyDims.Size.X, footerDims.Size.X))
	height := headerDims.Size.Y + bodyDims.Size.Y + footerDims.Size.Y + bodyGap + footerGap
	size := gtx.Constraints.Constrain(image.Pt(width, height))

	y := 0
	if headerDims.Size.Y > 0 {
		headerCall.Add(gtx.Ops)
		y += headerDims.Size.Y
	}
	if bodyDims.Size.Y > 0 {
		y += bodyGap
		stack := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		bodyCall.Add(gtx.Ops)
		stack.Pop()
		y += bodyDims.Size.Y
	}
	if footerDims.Size.Y > 0 {
		y += footerGap
		pos := m.dialogFooterPosition(size, footerDims.Size, y)
		stack := op.Offset(pos).Push(gtx.Ops)
		footerCall.Add(gtx.Ops)
		stack.Pop()
	}
	return layout.Dimensions{Size: size}
}

func (m ModalWidget) dialogFooterPosition(content, footer image.Point, y int) image.Point {
	if m.size == ModalCover || m.size == ModalFull {
		y = max(y, content.Y-footer.Y)
	}
	return image.Pt(max(content.X-footer.X, 0), y)
}

func (m ModalWidget) dialogGaps(ctx *Context, gtx layout.Context, hasHeader, hasFooter bool) (bodyGap, footerGap int) {
	theme := ctx.Theme.Components.Modal
	if hasHeader && m.body != nil {
		bodyGap = gtx.Dp(theme.BodyGap)
	}
	if hasFooter {
		if m.body != nil || hasHeader {
			footerGap = gtx.Dp(theme.SectionGap)
		} else {
			footerGap = gtx.Dp(theme.FooterGap)
		}
	}
	return bodyGap, footerGap
}

func (m ModalWidget) recordHeader(ctx *Context, gtx layout.Context) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := m.layoutHeader(ctx, gtx)
	return macro.Stop(), dims
}

func (m ModalWidget) layoutHeader(ctx *Context, gtx layout.Context) layout.Dimensions {
	if m.header != nil {
		return m.header.Layout(ctx, gtx)
	}
	if m.title == "" && m.icon == nil {
		return layout.Dimensions{}
	}
	children := make([]layout.FlexChild, 0, 2)
	if m.icon != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutIcon(ctx, gtx)
		}))
	}
	if m.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(m.title).
				Size(float32(ctx.Theme.Components.Modal.TitleSize)).
				Weight(font.Medium).
				Color(ctx.Theme.Palette.overlayForegroundColor()).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(ctx.Theme.Components.Modal.HeaderGap),
	}.Layout(gtx, children...)
}

func (m ModalWidget) layoutIcon(ctx *Context, gtx layout.Context) layout.Dimensions {
	size := image.Pt(gtx.Dp(ctx.Theme.Components.Modal.IconSize), gtx.Dp(ctx.Theme.Components.Modal.IconSize))
	gtx.Constraints = layout.Exact(size)
	drawModalIconFrame(gtx, ctx.Theme, size)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return m.icon.Layout(ctx, gtx)
	})
}

func (m ModalWidget) recordBody(ctx *Context, gtx layout.Context, state *modalState) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := m.layoutBody(ctx, gtx, state)
	return macro.Stop(), dims
}

func (m ModalWidget) layoutBody(ctx *Context, gtx layout.Context, state *modalState) layout.Dimensions {
	if m.body == nil {
		return layout.Dimensions{}
	}
	body := m.styleBody(ctx, m.body)
	if m.scroll == ModalScrollOutside {
		return body.Layout(ctx, gtx)
	}
	state.bodyList.Axis = layout.Vertical
	state.bodyList.ScrollAnyAxis = false
	return state.bodyList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return body.Layout(ctx, gtx)
	})
}

func (m ModalWidget) styleBody(ctx *Context, body Widget) Widget {
	text, ok := body.(TextWidget)
	if !ok {
		return body
	}
	if text.size == 0 {
		text = text.Size(float32(ctx.Theme.Components.Modal.BodyTextSize))
	}
	if !text.hasColor {
		text = text.Color(ctx.Theme.Palette.MutedForeground)
	}
	return text
}

func (m ModalWidget) recordFooter(ctx *Context, gtx layout.Context) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := m.layoutFooter(ctx, gtx)
	return macro.Stop(), dims
}

func (m ModalWidget) layoutFooter(ctx *Context, gtx layout.Context) layout.Dimensions {
	if m.footer == nil {
		return layout.Dimensions{}
	}
	return m.footer.Layout(ctx, gtx)
}

func (m ModalWidget) layoutCloseButton(ctx *Context, gtx layout.Context, state *modalState, dialogSize image.Point) {
	theme := ctx.Theme.Components.Modal
	size := image.Pt(gtx.Dp(theme.CloseSize), gtx.Dp(theme.CloseSize))
	inset := gtx.Dp(theme.CloseInset)
	pos := image.Pt(max(dialogSize.X-inset-size.X, 0), min(inset, max(dialogSize.Y-size.Y, 0)))

	presses := activePresses(state.close.History())
	ctx.focusOnPress(&state.close, state.close.History(), presses)
	buttonGtx := gtx
	buttonGtx.Constraints = layout.Exact(size)
	stack := op.Offset(pos).Push(gtx.Ops)
	state.close.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(modalCloseLabel(ctx)).Add(gtx.Ops)
		focused := gtx.Focused(&state.close)
		focusVisible := state.closeFocus.focusVisible(focused, state.close.History())
		drawModalCloseButton(gtx, ctx.Theme, size, state.close.Hovered(), state.close.Pressed(), focusVisible)
		return layout.Dimensions{Size: size}
	})
	stack.Pop()
}

func shrinkPoint(p image.Point, amount int) image.Point {
	return image.Pt(max(p.X-amount, 0), max(p.Y-amount, 0))
}

func modalCloseLabel(ctx *Context) string {
	if ctx.Language == LanguageChinese {
		return "关闭"
	}
	return "Close"
}
