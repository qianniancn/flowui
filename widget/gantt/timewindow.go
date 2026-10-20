package gantt

import "time"

// TimeWindow is a visible time range. It is a plain value type so an
// application can keep the window it last observed and feed it back through
// Widget.TimeWindow when the chart is controlled.
type TimeWindow struct {
	// Start marks the beginning of the visible time range.
	Start time.Time
	// End is the end of the visible time range.
	End time.Time
}
