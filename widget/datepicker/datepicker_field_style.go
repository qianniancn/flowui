package datepicker

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func resolveDatePickerField(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	custom style.Style,
	visual theme.DatePickerVisual,
) field.Resolved {
	// The owning widget already resolved and cached the complete visual bundle.
	// Read Defaults directly so field resolution does not maintain a second
	// cache for a value that has no independent invalidation boundary.
	defaults := visual.Defaults
	content := computed.ResolvePart(ctx, gtx, key, style.PartContent, state, defaults, style.Style{}, style.Style{}, custom)
	placeholder := computed.ResolvePart(ctx, gtx, key, style.PartPlaceholder, state, defaults, style.Style{}, style.Style{}, custom)
	selection := computed.ResolvePart(ctx, gtx, key, style.PartSelection, state, defaults, style.Style{}, style.Style{}, custom)
	return field.Resolved{
		Content: content,
		Colors:  datePickerFieldColors(content, placeholder, selection, visual),
	}
}

func datePickerFieldColors(content, placeholder, selection style.ResolvedStyle, visual theme.DatePickerVisual) field.Colors {
	resolved := visual.FieldColors(content, placeholder, selection)
	return field.Colors{
		Foreground:  resolved.Foreground,
		Placeholder: resolved.Placeholder,
		Selection:   resolved.Selection,
	}
}
