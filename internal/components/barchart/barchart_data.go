package barchart

import (
	"fmt"
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type resolvedBar struct {
	value float64
	start float64
	end   float64
	color color.NRGBA
	valid bool
}

type resolvedSeries struct {
	key           string
	label         string
	hidden        bool
	columnID      string
	columnIndex   int
	values        []resolvedBar
	color         color.NRGBA
	minHeight     int
	radius        int
	hasRadius     bool
	showLabels    bool
	labelPosition LabelPosition
	formatLabel   func(float64) string
}

type barColumn struct {
	id             string
	width          float32
	maxWidth       float32
	showBackground bool
}

type dataExtent = chart.Extent

type chartData struct {
	series     []resolvedSeries
	legend     []resolvedSeries
	columns    []barColumn
	categories int
	yExtent    dataExtent
	generation uint64
}

type chartDataCache struct {
	cache chart.DataCache[chartData]
}

func (c *chartDataCache) resolve(widget Widget, activeTheme *theme.Theme, metric unit.Metric) chartData {
	data, generation := c.cache.Resolve(widget.hasDataVersion, widget.dataVersion, activeTheme, metric, func() chartData {
		return resolveChartData(widget, activeTheme, metric.Dp)
	})
	data.generation = generation
	return data
}

func resolveChartData(widget Widget, activeTheme *theme.Theme, dp func(unit.Dp) int) chartData {
	seen := make(map[string]struct{}, len(widget.series))
	columnIndexes := make(map[string]int, len(widget.series))
	data := chartData{categories: len(widget.categories)}
	for _, source := range widget.series {
		if !source.hidden {
			data.categories = max(data.categories, len(source.values))
		}
	}

	for index, source := range widget.series {
		if source.key == "" {
			panic("flowui: empty bar chart series key")
		}
		if _, exists := seen[source.key]; exists {
			panic(fmt.Sprintf("flowui: duplicate bar chart series key %q", source.key))
		}
		seen[source.key] = struct{}{}
		label := source.label
		if label == "" {
			label = source.key
		}
		barColor := source.color
		if !source.hasColor {
			colors := activeTheme.Components.BarChart.SeriesColors
			barColor = colors[index%len(colors)]
		}
		data.legend = append(data.legend, resolvedSeries{
			key:    source.key,
			label:  label,
			hidden: source.hidden,
			color:  barColor,
		})
		if source.hidden {
			continue
		}
		columnID := source.stack
		if columnID == "" {
			columnID = "\x00" + source.key
		} else {
			columnID = "\x01" + columnID
		}
		columnIndex, exists := columnIndexes[columnID]
		if !exists {
			columnIndex = len(data.columns)
			columnIndexes[columnID] = columnIndex
			data.columns = append(data.columns, barColumn{id: columnID})
		}
		column := &data.columns[columnIndex]
		if source.width > 0 && column.width == 0 {
			column.width = float32(dp(source.width))
		}
		if source.maxWidth > 0 {
			column.maxWidth = float32(dp(source.maxWidth))
		}
		column.showBackground = column.showBackground || source.showBackground

		resolved := resolvedSeries{
			key:           source.key,
			label:         label,
			columnID:      columnID,
			columnIndex:   columnIndex,
			values:        make([]resolvedBar, data.categories),
			color:         barColor,
			minHeight:     dp(source.minHeight),
			radius:        dp(source.radius),
			hasRadius:     source.hasRadius,
			showLabels:    source.showLabels,
			labelPosition: source.labelPosition,
			formatLabel:   source.formatLabel,
		}
		for valueIndex, value := range source.values {
			if chart.Finite(value) {
				itemColor := barColor
				if valueIndex < len(source.itemColors) {
					itemColor = source.itemColors[valueIndex]
				}
				resolved.values[valueIndex] = resolvedBar{value: value, color: itemColor, valid: true}
			}
		}
		data.series = append(data.series, resolved)
	}

	positive := make([]float64, len(data.columns))
	negative := make([]float64, len(data.columns))
	for category := 0; category < data.categories; category++ {
		clear(positive)
		clear(negative)
		for seriesIndex := range data.series {
			bar := &data.series[seriesIndex].values[category]
			if !bar.valid {
				continue
			}
			columnIndex := data.series[seriesIndex].columnIndex
			if bar.value >= 0 {
				bar.start = positive[columnIndex]
				bar.end = bar.start + bar.value
				positive[columnIndex] = bar.end
			} else {
				bar.start = negative[columnIndex]
				bar.end = bar.start + bar.value
				negative[columnIndex] = bar.end
			}
			if widget.includeZero || bar.start != 0 {
				data.yExtent.Include(bar.start)
			}
			data.yExtent.Include(bar.end)
		}
	}
	return data
}
