package chart

import "image/color"

// Datum describes one visible series value in a chart selection.
type Datum struct {
	SeriesKey   string
	SeriesLabel string
	X           float64
	Y           float64
	Color       color.NRGBA
}

// Selection describes the values selected at one X position or category.
type Selection struct {
	Label string
	// Index is the category index, or -1 for non-category Cartesian data.
	Index int
	X     float64
	Items []Datum
}
