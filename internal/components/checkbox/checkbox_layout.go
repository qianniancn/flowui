package checkbox

import (
	"gioui.org/font"
	"gioui.org/layout"
	giotext "gioui.org/text"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (c CheckboxWidget) layoutContent(ctx *frame.Context, gtx layout.Context, options ControlOptions, indicatorState IndicatorState) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return c.layoutRow(ctx, gtx, options, indicatorState)
	}
	message := c.supportingText()
	if message == "" {
		return row(gtx)
	}
	tokens := frame.ActiveTheme(ctx).Components.Checkbox
	return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(tokens.DescriptionGap)}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: tokens.DescriptionIndent}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if c.invalid && c.errorMessage != "" {
					label := material.Label(frame.ActiveTheme(ctx).Material, frame.ActiveTheme(ctx).Components.Description.TextSize, c.errorMessage)
					label.Color = checkboxSupportingColor(frame.ActiveTheme(ctx), frame.ActiveTheme(ctx).Palette.Danger, options.Disabled)
					label.WrapPolicy = giotext.WrapHeuristically
					return label.Layout(gtx)
				}
				return description.Description(c.description).For(c.key).Disabled(options.Disabled).Layout(ctx, gtx)
			})
		}),
	)
}

func (c CheckboxWidget) layoutRow(ctx *frame.Context, gtx layout.Context, options ControlOptions, indicatorState IndicatorState) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return DrawControl(ctx, gtx, options)
	}
	if c.label == "" {
		return control(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(frame.ActiveTheme(ctx).Components.Checkbox.LabelGap),
	}.Layout(gtx,
		layout.Rigid(control),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return text.New(c.label).
						Size(float32(frame.ActiveTheme(ctx).Typography.ControlSize)).
						Color(checkboxLabelColor(frame.ActiveTheme(ctx), indicatorState.Disabled)).
						Weight(font.Medium).
						Layout(ctx, gtx)
				}),
			}
			if c.required {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					color := checkboxSupportingColor(frame.ActiveTheme(ctx), frame.ActiveTheme(ctx).Palette.Danger, indicatorState.Disabled)
					return text.New("*").Size(float32(frame.ActiveTheme(ctx).Typography.ControlSize)).Color(color).Layout(ctx, gtx)
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline, Gap: gtx.Dp(frame.ActiveTheme(ctx).Components.Label.RequiredMarkOffset)}.Layout(gtx, children...)
		}),
	)
}
