package menu

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
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

func (m Widget) preferredWidthPxCached(ctx *frame.Context, gtx layout.Context, state *menuState) int {
	if state == nil || !m.hasDataVersion || m.customStyle.CacheUnsafe() {
		return m.preferredWidthPx(ctx, gtx)
	}
	key := menuWidthCache{
		version:         m.dataVersion,
		themeGeneration: frame.ThemeGeneration(ctx),
		maxWidth:        gtx.Constraints.Max.X,
		maxHeight:       gtx.Constraints.Max.Y,
		pxPerDp:         gtx.Metric.PxPerDp,
		pxPerSp:         gtx.Metric.PxPerSp,
		styleHash:       m.customStyle.Hash64(),
		compact:         m.compact,
		autoSeparate:    m.autoSeparateSections,
		hasMinWidth:     m.hasMinWidth,
		minWidth:        gtx.Dp(m.minWidth),
		minWidthPx:      m.minWidthPx,
		hasMaxWidth:     m.hasMaxWidth,
		maxWidthLimit:   gtx.Dp(m.maxWidth),
		emptyText:       m.emptyText,
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
	key.width = m.preferredWidthPx(ctx, gtx)
	key.ready = true
	state.widthCache = key
	return key.width
}

// preferredWidthPx measures the widest menu entry without painting it. A
// separate pass keeps AutoWidth independent from the fixed-width Flex layout
// used for the actual menu.
func (m Widget) preferredWidthPx(ctx *frame.Context, gtx layout.Context) int {
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		return 0
	}
	entries := m.entries()
	tokens := m.themeTokens(ctx)
	width := 0
	if len(entries) == 0 {
		width = m.measureTextWidth(ctx, gtx, m.emptyText, tokens.ItemTextSize, font.Normal)
	}
	for _, entry := range entries {
		if entry.separator {
			continue
		}
		if entry.sectionTitle != "" {
			titleWidth := m.measureTextWidth(ctx, gtx, entry.sectionTitle, tokens.SectionTextSize, font.Medium)
			titleWidth += 2 * gtx.Dp(tokens.SectionPaddingX)
			width = max(width, titleWidth)
			continue
		}
		width = max(width, m.preferredItemWidthPx(ctx, gtx, entry.item))
	}
	if m.beforeContent != nil {
		width = max(width, m.measureWidgetWidth(ctx, gtx, m.beforeContent))
	}
	if m.afterContent != nil {
		width = max(width, m.measureWidgetWidth(ctx, gtx, m.afterContent))
	}
	return width + 2*gtx.Dp(tokens.Padding)
}

func (m Widget) preferredItemWidthPx(ctx *frame.Context, gtx layout.Context, item Item) int {
	tokens := m.themeTokens(ctx)
	type measuredPart struct {
		width    int
		gapAfter int
	}
	parts := make([]measuredPart, 0, 5)
	add := func(width, gapAfter int) {
		if width > 0 {
			parts = append(parts, measuredPart{width: width, gapAfter: gapAfter})
		}
	}
	if item.Kind == ItemCheckbox || item.Kind == ItemRadio || item.Indicator != nil || item.IndicatorType != IndicatorNone {
		add(gtx.Dp(tokens.IndicatorSize), gtx.Dp(tokens.IndicatorContentGap))
	}
	if item.Leading != nil {
		add(m.measureWidgetWidth(ctx, gtx, item.Leading), gtx.Dp(tokens.ItemContentGap))
	}
	labelWidth := m.measureTextWidth(ctx, gtx, item.Label, tokens.ItemTextSize, m.preferredItemFontWeight(ctx, item))
	if item.Description != "" {
		labelWidth = max(labelWidth, m.measureTextWidth(ctx, gtx, item.Description, tokens.ItemDescriptionSize, font.Normal))
	}
	add(labelWidth, gtx.Dp(tokens.ItemContentGap))
	if item.Shortcut != "" {
		shortcut := m.measureTextWidth(ctx, gtx, item.Shortcut, tokens.ShortcutTextSize, font.Medium)
		shortcut += 2 * gtx.Dp(tokens.ShortcutPaddingX)
		add(shortcut, gtx.Dp(tokens.ItemContentGap))
	}
	if item.Trailing != nil {
		add(m.measureWidgetWidth(ctx, gtx, item.Trailing), gtx.Dp(tokens.ItemContentGap))
	}
	if itemHasSubmenu(item) {
		indicatorWidth := gtx.Dp(tokens.SubmenuIndicatorSize)
		if item.SubmenuIndicator != nil {
			indicatorWidth = m.measureWidgetWidth(ctx, gtx, item.SubmenuIndicator)
		}
		add(indicatorWidth, 0)
	}
	content := 0
	for index, part := range parts {
		content += part.width
		if index+1 < len(parts) {
			content += part.gapAfter
		}
	}
	part := m.resolveItemMeasurementStyle(ctx, item)
	paddingLeft := tokens.ItemPaddingX
	paddingRight := tokens.ItemPaddingX
	if part.Box != nil && part.Box.Padding != nil {
		paddingLeft = part.Box.Padding.Left
		paddingRight = part.Box.Padding.Right
	}
	return content + gtx.Dp(paddingLeft) + gtx.Dp(paddingRight)
}

func (m Widget) preferredItemFontWeight(ctx *frame.Context, item Item) font.Weight {
	weight := menuItemStyle(frame.ActiveTheme(ctx), item.Variant).fontWeight
	part := m.resolveItemMeasurementStyle(ctx, item)
	if part.Text != nil && part.Text.FontWeight != nil {
		weight = font.Weight(*part.Text.FontWeight)
	}
	return weight
}

func (m Widget) resolveItemMeasurementStyle(ctx *frame.Context, item Item) flowstyle.ResolvedStyle {
	return styleruntime.ResolvePartStatic(
		ctx,
		flowstyle.PartItem,
		flowstyle.StyleState{},
		menuItemDefaultDeclaration(frame.ActiveTheme(ctx), m.themeTokens(ctx)),
		menuItemVariantDeclaration(frame.ActiveTheme(ctx), item.Variant),
		flowstyle.Style{},
		m.customStyle,
	)
}

func (m Widget) measureTextWidth(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight) int {
	if value == "" {
		return 0
	}
	measureGtx := gtx
	measureGtx.Constraints = layout.Constraints{Max: image.Pt(gtx.Constraints.Max.X, max(gtx.Constraints.Max.Y, gtx.Dp(1024)))}
	return text.Measure(ctx, measureGtx, text.New(value).Size(float32(size)).Weight(weight)).Size.X
}

func (m Widget) measureWidgetWidth(ctx *frame.Context, gtx layout.Context, value frame.Widget) int {
	if value == nil {
		return 0
	}
	measureGtx := gtx
	measureGtx.Constraints = layout.Constraints{Max: image.Pt(gtx.Constraints.Max.X, max(gtx.Constraints.Max.Y, gtx.Dp(1024)))}
	return frame.MeasureWidget(ctx, measureGtx, value).Size.X
}
