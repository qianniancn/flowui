package menu

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type menuWidthCache struct {
	ready           bool
	version         uint64
	themeGeneration uint64
	maxWidth        int
	maxHeight       int
	pxPerDp         float32
	pxPerSp         float32
	styleHash       uint64
	compact         bool
	autoSeparate    bool
	hasMinWidth     bool
	minWidth        int
	minWidthPx      int
	hasMaxWidth     bool
	maxWidthLimit   int
	emptyText       string
	width           int
}

func (m Widget) preferredWidthPxCachedWithVisual(ctx *core.Context, gtx layout.Context, state *menuState, visual theme.MenuVisual) int {
	if state == nil || !m.hasDataVersion || visual.Override.CacheUnsafe() {
		return m.preferredWidthPxWithVisual(ctx, gtx, visual)
	}
	key := menuWidthCache{
		version:         m.dataVersion,
		themeGeneration: core.ThemeGeneration(ctx),
		maxWidth:        gtx.Constraints.Max.X,
		maxHeight:       gtx.Constraints.Max.Y,
		pxPerDp:         gtx.Metric.PxPerDp,
		pxPerSp:         gtx.Metric.PxPerSp,
		styleHash:       visual.Override.Hash64(),
		compact:         m.compact,
		autoSeparate:    m.autoSeparateSections,
		hasMinWidth:     m.hasMinWidth,
		minWidth:        gtx.Dp(m.minWidth),
		minWidthPx:      m.minWidthPx,
		hasMaxWidth:     m.hasMaxWidth,
		maxWidthLimit:   gtx.Dp(m.maxWidth),
		emptyText:       m.resolvedEmptyText(ctx),
	}
	if state.widthCache.ready && state.widthCache.version == key.version &&
		state.widthCache.themeGeneration == key.themeGeneration &&
		state.widthCache.maxWidth == key.maxWidth && state.widthCache.maxHeight == key.maxHeight &&
		state.widthCache.pxPerDp == key.pxPerDp && state.widthCache.pxPerSp == key.pxPerSp &&
		state.widthCache.styleHash == key.styleHash && state.widthCache.compact == key.compact &&
		state.widthCache.autoSeparate == key.autoSeparate && state.widthCache.hasMinWidth == key.hasMinWidth &&
		state.widthCache.minWidth == key.minWidth && state.widthCache.minWidthPx == key.minWidthPx &&
		state.widthCache.hasMaxWidth == key.hasMaxWidth && state.widthCache.maxWidthLimit == key.maxWidthLimit &&
		state.widthCache.emptyText == key.emptyText {
		return state.widthCache.width
	}
	key.width = m.preferredWidthPxWithVisual(ctx, gtx, visual)
	key.ready = true
	state.widthCache = key
	return key.width
}

// preferredWidthPx measures the widest menu entry without painting it. A
// separate pass keeps AutoWidth independent from the fixed-width Flex layout
// used for the actual menu.
func (m Widget) preferredWidthPxWithVisual(ctx *core.Context, gtx layout.Context, visual theme.MenuVisual) int {
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		return 0
	}
	entries := m.entries()
	metrics := visual.Metrics
	width := 0
	if len(entries) == 0 {
		width = visual.TextWidth(gtx, theme.MenuTextRequest{Text: m.resolvedEmptyText(ctx), Color: visual.Colors.Description, Role: theme.MenuTextEmpty})
	}
	for _, entry := range entries {
		if entry.separator {
			continue
		}
		if entry.sectionTitle != "" {
			titleWidth := visual.TextWidth(gtx, theme.MenuTextRequest{Text: entry.sectionTitle, Color: visual.Colors.Description, Role: theme.MenuTextSection})
			titleWidth += 2 * gtx.Dp(metrics.SectionPaddingX)
			width = max(width, titleWidth)
			continue
		}
		width = max(width, m.preferredItemWidthPxWithVisual(ctx, gtx, entry.item, visual))
	}
	if m.beforeContent != nil {
		width = max(width, m.measureWidgetWidth(ctx, gtx, m.beforeContent))
	}
	if m.afterContent != nil {
		width = max(width, m.measureWidgetWidth(ctx, gtx, m.afterContent))
	}
	return width + 2*gtx.Dp(metrics.Padding)
}

func (m Widget) preferredItemWidthPxWithVisual(ctx *core.Context, gtx layout.Context, item Item, visual theme.MenuVisual) int {
	metrics := visual.Metrics
	itemVisual := visual.Item(theme.MenuItemRequest{Variant: uint8(item.Variant)})
	measurementStyle := resolveItemMeasurementStyleForVisual(ctx, itemVisual, visual.Override)
	if itemVisual.Colors == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	itemColors := itemVisual.Colors(measurementStyle)
	content := 0
	pendingGap := 0
	add := func(width, gapAfter int) {
		if width <= 0 {
			return
		}
		content += pendingGap + width
		pendingGap = gapAfter
	}
	if item.Kind == ItemCheckbox || item.Kind == ItemRadio || item.Indicator != nil || item.IndicatorType != IndicatorNone {
		add(gtx.Dp(metrics.IndicatorSize), gtx.Dp(metrics.IndicatorContentGap))
	}
	if item.Leading != nil {
		add(m.measureWidgetWidth(ctx, gtx, item.Leading), gtx.Dp(metrics.ItemContentGap))
	}
	labelWidth := visual.TextWidth(gtx, theme.MenuTextRequest{Text: item.Label, Color: itemColors.Foreground, Role: theme.MenuTextItem, Style: measurementStyle})
	if item.Description != "" {
		labelWidth = max(labelWidth, visual.TextWidth(gtx, theme.MenuTextRequest{Text: item.Description, Color: itemColors.Description, Role: theme.MenuTextDescription, Style: measurementStyle}))
	}
	add(labelWidth, gtx.Dp(metrics.ItemContentGap))
	if item.Shortcut != "" {
		shortcut := visual.TextWidth(gtx, theme.MenuTextRequest{Text: item.Shortcut, Color: itemColors.Shortcut, Role: theme.MenuTextShortcut})
		shortcut += 2 * gtx.Dp(metrics.ShortcutPaddingX)
		add(shortcut, gtx.Dp(metrics.ItemContentGap))
	}
	if item.Trailing != nil {
		add(m.measureWidgetWidth(ctx, gtx, item.Trailing), gtx.Dp(metrics.ItemContentGap))
	}
	if itemHasSubmenu(item) {
		indicatorWidth := gtx.Dp(metrics.SubmenuIndicatorSize)
		if item.SubmenuIndicator != nil {
			indicatorWidth = m.measureWidgetWidth(ctx, gtx, item.SubmenuIndicator)
		}
		add(indicatorWidth, 0)
	}
	paddingLeft := metrics.ItemPaddingX
	paddingRight := metrics.ItemPaddingX
	if measurementStyle.Box != nil && measurementStyle.Box.Padding != nil {
		paddingLeft = measurementStyle.Box.Padding.Left
		paddingRight = measurementStyle.Box.Padding.Right
	}
	return content + gtx.Dp(paddingLeft) + gtx.Dp(paddingRight)
}

func (m Widget) resolveItemMeasurementStyleWithVisual(ctx *core.Context, item Item, visual theme.MenuVisual) style.ResolvedStyle {
	return resolveItemMeasurementStyleForVisual(ctx, visual.Item(theme.MenuItemRequest{Variant: uint8(item.Variant)}), visual.Override)
}

func resolveItemMeasurementStyleForVisual(ctx *core.Context, itemVisual theme.MenuItemVisual, override style.Style) style.ResolvedStyle {
	return computed.ResolvePartStatic(
		ctx,
		style.PartItem,
		style.State{},
		itemVisual.Defaults,
		itemVisual.Variant,
		style.Style{},
		override,
	)
}

func (m Widget) measureWidgetWidth(ctx *core.Context, gtx layout.Context, value core.Widget) int {
	if value == nil {
		return 0
	}
	measureGtx := gtx
	measureGtx.Constraints = layout.Constraints{Max: image.Pt(gtx.Constraints.Max.X, max(gtx.Constraints.Max.Y, gtx.Dp(1024)))}
	return core.MeasureWidget(ctx, measureGtx, value).Size.X
}
