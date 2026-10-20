package alertdialog

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
)

type dialogHeader struct {
	title       string
	description string
	status      Status
	icon        core.Widget
	cache       *alertDialogVisualCache
	size        Size
	backdrop    BackdropVariant
	override    *theme.ModalOverride
}

func (h dialogHeader) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := h.visual(ctx)
	if h.title != "" {
		semantic.LabelOp(h.title).Add(gtx.Ops)
	}
	if h.description != "" {
		semantic.DescriptionOp(h.description).Add(gtx.Ops)
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return h.layoutIcon(ctx, gtx, visual)
		}),
	}
	if h.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.WithDefaults(text.New(h.title), visual.Title).Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
		Gap:       gtx.Dp(visual.Metrics.HeaderGap),
	}.Layout(gtx, children...)
}

func (h dialogHeader) layoutIcon(ctx *core.Context, gtx layout.Context, visual theme.AlertDialogVisual) layout.Dimensions {
	diameter := min(gtx.Dp(visual.Metrics.IconSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	if diameter <= 0 {
		return layout.Dimensions{}
	}
	size := image.Pt(diameter, diameter)
	gtx.Constraints = layout.Exact(size)
	return visual.IconFrame(gtx, theme.AlertDialogIconFrameRequest{
		Size: size,
		Content: func(gtx layout.Context) layout.Dimensions {
			glyphSize := min(gtx.Dp(visual.Metrics.IconGlyphSize), diameter)
			gtx.Constraints = layout.Exact(image.Pt(glyphSize, glyphSize))
			restore := core.PushColors(ctx, visual.IconForeground, visual.IconBackground)
			defer restore()
			if h.icon != nil {
				return h.icon.Layout(ctx, gtx)
			}
			if len(visual.Icon) == 0 {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}
			return icon.Layout(visual.Icon, gtx, visual.IconForeground)
		},
	})
}
