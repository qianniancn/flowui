package checkbox

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/text"
)

func (c Widget) layoutContent(ctx *core.Context, gtx layout.Context, options ControlOptions, indicatorState IndicatorState, resolved checkboxResolvedStyle) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return c.layoutRow(ctx, gtx, options, indicatorState, resolved)
	}
	message := c.supportingText()
	if message == "" {
		return row(gtx)
	}
	return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(resolved.metrics.DescriptionGap)}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: resolved.metrics.DescriptionIndent}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layoutui.Resolved(ctx, gtx, resolved.description, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					return description.LayoutWithResolvedText(ctx, gtx, description.New(message).
						For(c.key).
						Disabled(options.Disabled), resolved.description.Text)
				}))
			})
		}),
	)
}

func (c Widget) layoutRow(ctx *core.Context, gtx layout.Context, options ControlOptions, indicatorState IndicatorState, resolved checkboxResolvedStyle) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return DrawControl(ctx, gtx, options)
	}
	if c.label == "" {
		return control(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(resolved.metrics.LabelGap),
	}.Layout(gtx,
		layout.Rigid(control),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved.label, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				children := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return text.WithResolvedText(text.New(c.label), resolved.label.Text).Layout(ctx, gtx)
				})}
				if c.required {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						color := resolved.colors.RequiredMark
						if indicatorState.Disabled {
							color = resolved.colors.DisabledRequiredMark
						}
						return text.New("*").Color(color).Layout(ctx, gtx)
					}))
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline, Gap: gtx.Dp(resolved.metrics.RequiredMarkOffset)}.Layout(gtx, children...)
			}))
		}),
	)
}
