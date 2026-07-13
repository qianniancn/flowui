package chart

import "image/color"

// Datum describes one visible value in a chart selection.
type Datum struct {
	SeriesKey   string
	SeriesLabel string
	X           float64
	Y           float64
	// Percent is populated by proportional charts such as PieChart.
	Percent float64
	// Open, Close, Low, and High are populated by CandlestickChart.
	Open  float64
	Close float64
	Low   float64
	High  float64
	Color color.NRGBA
}

// Selection describes the current chart selection.
type Selection struct {
	Label string
	// Index is the source data or category index, or -1 when not applicable.
	Index int
	X     float64
	Items []Datum
}
