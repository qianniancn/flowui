package chart

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
)

// Axis identifies a chart axis.
type Axis uint8

const (
	AxisX Axis = iota
	AxisY
)

// MarkLine draws a labeled reference line on one axis.
type MarkLine struct {
	Axis  Axis
	Value float64
	Label string
	color color.NRGBA
	width unit.Dp
}

func NewMarkLine(axis Axis, value float64) MarkLine {
	validateAxis(axis)
	validateFinite("mark line value", value)
	return MarkLine{Axis: axis, Value: value}
}

func (m MarkLine) Text(label string) MarkLine {
	m.Label = label
	return m
}

func (m MarkLine) Color(value color.NRGBA) MarkLine {
	m.color = value
	return m
}

func (m MarkLine) Width(dp int) MarkLine {
	if dp <= 0 {
		panic("flowui: chart mark line width must be positive")
	}
	m.width = unit.Dp(dp)
	return m
}

func (m MarkLine) ResolvedColor(fallback color.NRGBA) color.NRGBA {
	if m.color.A != 0 {
		return m.color
	}
	return fallback
}

func (m MarkLine) ResolvedWidth(fallback unit.Dp) unit.Dp {
	if m.width > 0 {
		return m.width
	}
	return fallback
}

// MarkArea draws a labeled reference band on one axis.
type MarkArea struct {
	Axis  Axis
	Start float64
	End   float64
	Label string
	color color.NRGBA
}

func NewMarkArea(axis Axis, start, end float64) MarkArea {
	validateAxis(axis)
	validateFinite("mark area start", start)
	validateFinite("mark area end", end)
	if end <= start {
		panic("flowui: chart mark area end must be greater than start")
	}
	return MarkArea{Axis: axis, Start: start, End: end}
}

func (m MarkArea) Text(label string) MarkArea {
	m.Label = label
	return m
}

func (m MarkArea) Color(value color.NRGBA) MarkArea {
	m.color = value
	return m
}

func (m MarkArea) ResolvedColor(fallback color.NRGBA) color.NRGBA {
	if m.color.A != 0 {
		return m.color
	}
	return fallback
}

// MarkPoint draws a labeled point at one Cartesian coordinate. Bar charts
// interpret X as a category index.
type MarkPoint struct {
	X       float64
	Y       float64
	Label   string
	color   color.NRGBA
	size    unit.Dp
	content frame.Widget
}

func NewMarkPoint(x, y float64) MarkPoint {
	validateFinite("mark point X", x)
	validateFinite("mark point Y", y)
	return MarkPoint{X: x, Y: y}
}

func (m MarkPoint) Text(label string) MarkPoint {
	m.Label = label
	return m
}

func (m MarkPoint) Color(value color.NRGBA) MarkPoint {
	m.color = value
	return m
}

func (m MarkPoint) Size(dp int) MarkPoint {
	if dp <= 0 {
		panic("flowui: chart mark point size must be positive")
	}
	m.size = unit.Dp(dp)
	return m
}

// Content replaces the default point with a custom non-interactive FlowUI widget.
func (m MarkPoint) Content(content frame.Widget) MarkPoint {
	m.content = content
	return m
}

func (m MarkPoint) ResolvedColor(fallback color.NRGBA) color.NRGBA {
	if m.color.A != 0 {
		return m.color
	}
	return fallback
}

func (m MarkPoint) ResolvedSize(fallback unit.Dp) unit.Dp {
	if m.size > 0 {
		return m.size
	}
	return fallback
}

func (m MarkPoint) ResolvedContent() frame.Widget {
	return m.content
}

// LayoutMarkPointContent lays out custom point content centered at a chart
// coordinate. It returns the resolved point size and whether custom content
// replaced the default marker.
func LayoutMarkPointContent(ctx *frame.Context, gtx layout.Context, mark MarkPoint, center f32.Point, plot image.Rectangle, fallbackSize unit.Dp, fallbackColor color.NRGBA) (int, bool) {
	size := max(gtx.Dp(mark.ResolvedSize(fallbackSize)), 3)
	content := mark.ResolvedContent()
	if content == nil {
		return size, false
	}

	markerGtx := gtx.Disabled()
	markerGtx.Constraints = layout.Exact(image.Pt(size, size))
	macro := op.Record(gtx.Ops)
	restore := frame.PushColors(ctx, mark.ResolvedColor(fallbackColor), ctx.BackgroundColor())
	dims := content.Layout(ctx, markerGtx)
	restore()
	call := macro.Stop()

	position := image.Pt(
		int(math.Round(float64(center.X)))-dims.Size.X/2,
		int(math.Round(float64(center.Y)))-dims.Size.Y/2,
	)
	area := clip.Rect(plot).Push(gtx.Ops)
	offset := op.Offset(position).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	area.Pop()
	return size, true
}

func ValidateAnnotations(lines []MarkLine, areas []MarkArea, points []MarkPoint) {
	for _, line := range lines {
		validateAxis(line.Axis)
		validateFinite("mark line value", line.Value)
		if line.width < 0 {
			panic("flowui: chart mark line width must not be negative")
		}
	}
	for _, area := range areas {
		validateAxis(area.Axis)
		validateFinite("mark area start", area.Start)
		validateFinite("mark area end", area.End)
		if area.End <= area.Start {
			panic("flowui: chart mark area end must be greater than start")
		}
	}
	for _, point := range points {
		validateFinite("mark point X", point.X)
		validateFinite("mark point Y", point.Y)
		if point.size < 0 {
			panic("flowui: chart mark point size must not be negative")
		}
	}
}

func validateAxis(axis Axis) {
	if axis != AxisX && axis != AxisY {
		panic("flowui: invalid chart axis")
	}
}

func validateFinite(name string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		panic("flowui: chart " + name + " must be finite")
	}
}
