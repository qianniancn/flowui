package modal

import (
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
)

func (m ModalWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *modalState, progress float32, contentEnabled bool) layout.Dimensions {
	size := modalOverlaySize(gtx)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(size)

	style := modalStyleFor(frame.ActiveTheme(ctx), m.backdrop, m.size)
	drawModalBackdrop(gtx, size, style, progress)
	contentGtx := gtx
	if !contentEnabled {
		contentGtx = contentGtx.Disabled()
	}
	m.layoutFocusTrap(ctx, gtx, state, contentEnabled)

	dialogGtx := contentGtx
	dialogGtx.Constraints = m.dialogConstraints(ctx, gtx, size)
	macro := op.Record(gtx.Ops)
	dialogDims, dialogPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutDialogFrame(ctx, dialogGtx, state)
	})
	dialogCall := macro.Stop()

	dialogPos := m.dialogPosition(ctx, gtx, size, dialogDims.Size)
	dialogRect := image.Rectangle{
		Min: dialogPos,
		Max: dialogPos.Add(dialogDims.Size),
	}
	motion := m.dialogMotion(ctx, gtx, dialogRect, progress, state.opening())
	dialogPlacement.PlaceTransform(motion.transform.Mul(f32.AffineId().Offset(f32.Pt(float32(dialogPos.X), float32(dialogPos.Y)))))
	dialogPlacement.SetOpacity(motion.opacity)
	m.layoutDismissAreas(gtx, state, size, overlay.AffineRectBounds(dialogRect, motion.transform))
	transform := op.Affine(motion.transform).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, motion.opacity)
	offset := op.Offset(dialogPos).Push(gtx.Ops)
	m.layoutDialogBlocker(gtx, state, dialogDims.Size)
	dialogCall.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	transform.Pop()
	return layout.Dimensions{Size: size}
}

func (m ModalWidget) layoutFocusTrap(ctx *frame.Context, gtx layout.Context, state *modalState, contentEnabled bool) {
	m.redirectFocusedBoundary(ctx, gtx, &state.focusStart, state.tabFocusTag(m.showCloseButton() && contentEnabled))
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

func (m ModalWidget) redirectFocusedBoundary(ctx *frame.Context, gtx layout.Context, target event.Tag, redirect event.Tag) {
	for {
		e, ok := gtx.Event(key.FocusFilter{Target: target})
		if !ok {
			break
		}
		event, ok := e.(key.FocusEvent)
		if ok && event.Focus {
			frame.RequestFocus(ctx, redirect)
		}
	}
}

func (m ModalWidget) layoutFocusEnd(gtx layout.Context, state *modalState) {
	m.layoutFocusTag(gtx, &state.focusEnd)
}

func (m ModalWidget) layoutFocusTag(gtx layout.Context, target event.Tag) {
	stack := clip.Rect{Max: image.Pt(1, 1)}.Push(gtx.Ops)
	event.Op(gtx.Ops, target)
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

func (m ModalWidget) dialogConstraints(ctx *frame.Context, gtx layout.Context, size image.Point) layout.Constraints {
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

func modalWidth(ctx *frame.Context, gtx layout.Context, size ModalSize) int {
	theme := frame.ActiveTheme(ctx).Components.Modal
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

func (m ModalWidget) dialogMargin(ctx *frame.Context, gtx layout.Context, size image.Point) int {
	theme := frame.ActiveTheme(ctx).Components.Modal
	margin := gtx.Dp(theme.Margin)
	if size.X >= gtx.Dp(theme.DesktopBreakpoint) {
		margin = gtx.Dp(theme.DesktopMargin)
	}
	if m.size == ModalFull {
		return 0
	}
	return min(max(margin, 0), min(size.X, size.Y)/2)
}

func (m ModalWidget) dialogPosition(ctx *frame.Context, gtx layout.Context, overlay, dialog image.Point) image.Point {
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

func (m ModalWidget) resolvedPlacement(ctx *frame.Context, gtx layout.Context, size image.Point) ModalPlacement {
	if m.placement != ModalAuto {
		return m.placement
	}
	if size.X < gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.DesktopBreakpoint) {
		return ModalBottom
	}
	return ModalCenter
}

type modalDialogMotion struct {
	transform f32.Affine2D
	opacity   float32
}

func (m ModalWidget) dialogMotion(ctx *frame.Context, gtx layout.Context, rect image.Rectangle, progress float32, opening bool) modalDialogMotion {
	theme := frame.ActiveTheme(ctx).Components.Modal
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
		return render.Lerp(startScale, overshoot, modalEaseOutCubic(progress/0.65))
	case progress < 0.86:
		return render.Lerp(overshoot, 0.992, render.Ease((progress-0.65)/0.21))
	default:
		return render.Lerp(0.992, 1, modalEaseOutCubic((progress-0.86)/0.14))
	}
}

func modalDialogZoomOutScale(progress float32) float32 {
	return render.Lerp(1.04, 1, modalEaseOutCubic(progress))
}

func modalDialogPopScale(progress float32) float32 {
	progress = clampProgress(progress)
	switch {
	case progress < 0.58:
		return render.Lerp(0.92, 1.025, modalEaseOutCubic(progress/0.58))
	case progress < 0.82:
		return render.Lerp(1.025, 0.997, render.Ease((progress-0.58)/0.24))
	default:
		return render.Lerp(0.997, 1, modalEaseOutCubic((progress-0.82)/0.18))
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

func (m ModalWidget) layoutDismissAreas(gtx layout.Context, state *modalState, bounds image.Point, dialog image.Rectangle) {
	areas := overlay.DismissRects(image.Rectangle{Max: bounds}, dialog)
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

func (m ModalWidget) layoutDialogFrame(ctx *frame.Context, gtx layout.Context, state *modalState) layout.Dimensions {
	return m.layoutDialog(ctx, gtx, state)
}

func (m ModalWidget) layoutDialog(ctx *frame.Context, gtx layout.Context, state *modalState) layout.Dimensions {
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

func (m ModalWidget) layoutDialogSurface(ctx *frame.Context, gtx layout.Context, state *modalState) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Modal
	padding := gtx.Dp(theme.Padding)
	contentGtx := gtx
	contentGtx.Constraints.Min = shrinkPoint(contentGtx.Constraints.Min, padding*2)
	contentGtx.Constraints.Max = shrinkPoint(contentGtx.Constraints.Max, padding*2)

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	var contentPlacement frame.OverlayPlacement
	func() {
		restore := frame.PushColors(ctx, frame.ActiveTheme(ctx).Palette.OverlayForegroundColor(), frame.ActiveTheme(ctx).Palette.OverlayColor())
		defer restore()
		contentDims, contentPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return m.layoutDialogContent(ctx, contentGtx, state)
		})
	}()
	contentCall := macro.Stop()

	size := contentDims.Size.Add(image.Pt(padding*2, padding*2))
	size = gtx.Constraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := modalDialogRadius(gtx, frame.ActiveTheme(ctx), m.size, size)
	drawModalSurface(gtx, frame.ActiveTheme(ctx), rect, radius, m.size)

	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	contentPlacement.PlaceOffset(image.Pt(padding, padding))
	contentCall.Add(gtx.Ops)
	contentOffset.Pop()
	clipStack.Pop()

	if m.showCloseButton() {
		m.layoutCloseButton(ctx, gtx, state, size)
	}
	return layout.Dimensions{Size: size}
}

func (m ModalWidget) layoutDialogContent(ctx *frame.Context, gtx layout.Context, state *modalState) layout.Dimensions {
	if m.scroll == ModalScrollOutside {
		state.outsideList.Axis = layout.Vertical
		state.outsideList.ScrollAnyAxis = false
		return layoutui.LayoutTrackedList(ctx, gtx, &state.outsideList, 1, func(gtx layout.Context, _ int) layout.Dimensions {
			return m.layoutDialogSections(ctx, gtx, state)
		})
	}
	return m.layoutDialogSections(ctx, gtx, state)
}

func (m ModalWidget) layoutDialogSections(ctx *frame.Context, gtx layout.Context, state *modalState) layout.Dimensions {
	sectionGtx := gtx
	sectionGtx.Constraints.Min = image.Point{}

	headerCall, headerDims, headerPlacement := m.recordHeader(ctx, sectionGtx)
	footerCall, footerDims, footerPlacement := m.recordFooter(ctx, sectionGtx)
	bodyGap, footerGap := m.dialogGaps(ctx, sectionGtx, headerDims.Size.Y > 0, footerDims.Size.Y > 0)

	bodyGtx := sectionGtx
	maxBodyY := max(gtx.Constraints.Max.Y-headerDims.Size.Y-footerDims.Size.Y-bodyGap-footerGap, 0)
	bodyGtx.Constraints.Max.Y = maxBodyY
	bodyCall, bodyDims, bodyPlacement := m.recordBody(ctx, bodyGtx, state)

	width := max(headerDims.Size.X, max(bodyDims.Size.X, footerDims.Size.X))
	height := headerDims.Size.Y + bodyDims.Size.Y + footerDims.Size.Y + bodyGap + footerGap
	size := gtx.Constraints.Constrain(image.Pt(width, height))

	y := 0
	if headerDims.Size.Y > 0 {
		headerPlacement.PlaceOffset(image.Point{})
		headerCall.Add(gtx.Ops)
		y += headerDims.Size.Y
	}
	if bodyDims.Size.Y > 0 {
		y += bodyGap
		bodyPlacement.PlaceOffset(image.Pt(0, y))
		stack := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		bodyCall.Add(gtx.Ops)
		stack.Pop()
		y += bodyDims.Size.Y
	}
	if footerDims.Size.Y > 0 {
		y += footerGap
		pos := m.dialogFooterPosition(size, footerDims.Size, y)
		footerPlacement.PlaceOffset(pos)
		stack := op.Offset(pos).Push(gtx.Ops)
		footerCall.Add(gtx.Ops)
		stack.Pop()
	}
	if headerDims.Size.Y == 0 {
		headerPlacement.PlaceOffset(image.Point{})
	}
	if bodyDims.Size.Y == 0 {
		bodyPlacement.PlaceOffset(image.Pt(0, y))
	}
	if footerDims.Size.Y == 0 {
		footerPlacement.PlaceOffset(image.Pt(0, y))
	}
	return layout.Dimensions{Size: size}
}

func (m ModalWidget) dialogFooterPosition(content, footer image.Point, y int) image.Point {
	if m.size == ModalCover || m.size == ModalFull {
		y = max(y, content.Y-footer.Y)
	}
	return image.Pt(max(content.X-footer.X, 0), y)
}

func (m ModalWidget) dialogGaps(ctx *frame.Context, gtx layout.Context, hasHeader, hasFooter bool) (bodyGap, footerGap int) {
	theme := frame.ActiveTheme(ctx).Components.Modal
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

func (m ModalWidget) recordHeader(ctx *frame.Context, gtx layout.Context) (op.CallOp, layout.Dimensions, frame.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutHeader(ctx, gtx)
	})
	return macro.Stop(), dims, placement
}

func (m ModalWidget) layoutHeader(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
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
			return text.New(m.title).
				Size(float32(frame.ActiveTheme(ctx).Components.Modal.TitleSize)).
				Weight(font.Medium).
				Color(frame.ActiveTheme(ctx).Palette.OverlayForegroundColor()).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.HeaderGap),
	}.Layout(gtx, children...)
}

func (m ModalWidget) layoutIcon(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	size := image.Pt(gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.IconSize), gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.IconSize))
	gtx.Constraints = layout.Exact(size)
	drawModalIconFrame(gtx, frame.ActiveTheme(ctx), size)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return m.icon.Layout(ctx, gtx)
	})
}

func (m ModalWidget) recordBody(ctx *frame.Context, gtx layout.Context, state *modalState) (op.CallOp, layout.Dimensions, frame.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutBody(ctx, gtx, state)
	})
	return macro.Stop(), dims, placement
}

func (m ModalWidget) layoutBody(ctx *frame.Context, gtx layout.Context, state *modalState) layout.Dimensions {
	if m.body == nil {
		return layout.Dimensions{}
	}
	body := m.styleBody(ctx, m.body)
	if m.scroll == ModalScrollOutside {
		return body.Layout(ctx, gtx)
	}
	state.bodyList.Axis = layout.Vertical
	state.bodyList.ScrollAnyAxis = false
	return layoutui.LayoutTrackedScrollbar(ctx, gtx, &state.bodyList, &state.bodyBar, 1, !gtx.Enabled(), false, func(gtx layout.Context, _ int) layout.Dimensions {
		return body.Layout(ctx, gtx)
	})
}

func (m ModalWidget) styleBody(ctx *frame.Context, body frame.Widget) frame.Widget {
	text, ok := body.(text.Widget)
	if !ok {
		return body
	}
	text = text.DefaultSize(float32(frame.ActiveTheme(ctx).Components.Modal.BodyTextSize))
	text = text.DefaultColor(frame.ActiveTheme(ctx).Palette.MutedForeground)
	return text
}

func (m ModalWidget) recordFooter(ctx *frame.Context, gtx layout.Context) (op.CallOp, layout.Dimensions, frame.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutFooter(ctx, gtx)
	})
	return macro.Stop(), dims, placement
}

func (m ModalWidget) layoutFooter(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if m.footer == nil {
		return layout.Dimensions{}
	}
	return m.footer.Layout(ctx, gtx)
}

func (m ModalWidget) layoutCloseButton(ctx *frame.Context, gtx layout.Context, modalStateValue *modalState, dialogSize image.Point) {
	theme := frame.ActiveTheme(ctx).Components.Modal
	size := image.Pt(gtx.Dp(theme.CloseSize), gtx.Dp(theme.CloseSize))
	inset := gtx.Dp(theme.CloseInset)
	pos := image.Pt(max(dialogSize.X-inset-size.X, 0), min(inset, max(dialogSize.Y-size.Y, 0)))

	buttonGtx := gtx
	buttonGtx.Constraints = layout.Exact(size)
	stack := op.Offset(pos).Push(gtx.Ops)
	modalStateValue.close.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(modalCloseLabel(ctx)).Add(gtx.Ops)
		focused := gtx.Focused(&modalStateValue.close)
		focusVisible := frame.FocusVisible(ctx, &modalStateValue.close, focused)
		drawModalCloseButton(gtx, frame.ActiveTheme(ctx), size, modalStateValue.close.Hovered(), modalStateValue.close.Pressed(), focusVisible)
		return layout.Dimensions{Size: size}
	})
	stack.Pop()
}

func shrinkPoint(p image.Point, amount int) image.Point {
	return image.Pt(max(p.X-amount, 0), max(p.Y-amount, 0))
}

func modalCloseLabel(ctx *frame.Context) string {
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "关闭"
	}
	return "Close"
}
