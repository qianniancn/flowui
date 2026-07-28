package alertdialog

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/components/icon"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui-icons-lucide"
)

type dialogHeader struct {
	title       string
	description string
	status      Status
	icon        frame.Widget
}

func (h dialogHeader) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.AlertDialog
	style := alertDialogStyleFor(frame.ActiveTheme(ctx), h.status)
	if h.title != "" {
		semantic.LabelOp(h.title).Add(gtx.Ops)
	}
	if h.description != "" {
		semantic.DescriptionOp(h.description).Add(gtx.Ops)
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return h.layoutIcon(ctx, gtx, style)
		}),
	}
	if h.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(h.title).
				Size(float32(tokens.TitleSize)).
				Weight(font.Medium).
				Color(frame.ActiveTheme(ctx).Palette.OverlayForegroundColor()).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
		Gap:       gtx.Dp(tokens.HeaderGap),
	}.Layout(gtx, children...)
}

func (h dialogHeader) layoutIcon(ctx *frame.Context, gtx layout.Context, style alertDialogStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.AlertDialog
	diameter := min(gtx.Dp(tokens.IconSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	if diameter <= 0 {
		return layout.Dimensions{}
	}
	size := image.Pt(diameter, diameter)
	gtx.Constraints = layout.Exact(size)
	rect := image.Rectangle{Max: size}
	radius := diameter / 2
	paint.FillShape(gtx.Ops, style.iconBackground, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	stack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		glyphSize := min(gtx.Dp(tokens.IconGlyphSize), diameter)
		gtx.Constraints = layout.Exact(image.Pt(glyphSize, glyphSize))
		restore := frame.PushColors(ctx, style.iconForeground, style.iconBackground)
		defer restore()
		if h.icon != nil {
			return h.icon.Layout(ctx, gtx)
		}
		return icon.Layout(alertDialogIcon(h.status), gtx, style.iconForeground)
	})
	stack.Pop()
	return layout.Dimensions{Size: size, Baseline: dims.Baseline}
}

func alertDialogIcon(status Status) []byte {
	switch status {
	case StatusSuccess:
		return lucide.CircleCheck
	case StatusWarning:
		return lucide.TriangleAlert
	case StatusDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
