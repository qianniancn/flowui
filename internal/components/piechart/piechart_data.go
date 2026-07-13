package piechart

import (
	"fmt"
	"image/color"
	"math"

	"github.com/qianniancn/FlowUI/internal/theme"
)

const fullCircle = 2 * math.Pi

type resolvedSlice struct {
	index      int
	key        string
	label      string
	value      float64
	percent    float64
	color      color.NRGBA
	hidden     bool
	startAngle float32
	endAngle   float32
	rawAngle   float32
}

func (s resolvedSlice) sweep() float32 {
	return float32(math.Abs(float64(s.endAngle - s.startAngle)))
}

type chartData struct {
	slices []resolvedSlice
	legend []resolvedSlice
	total  float64
	start  float32
	dir    float32
}

func resolveChartData(widget Widget, activeTheme *theme.Theme) chartData {
	widget.validateRadii()
	seen := make(map[string]struct{}, len(widget.data))
	result := chartData{start: -widget.startAngle * math.Pi / 180, dir: 1}
	if !widget.clockwise {
		result.dir = -1
	}

	for index, source := range widget.data {
		if source.key == "" {
			panic("flowui: empty pie chart data key")
		}
		if _, exists := seen[source.key]; exists {
			panic(fmt.Sprintf("flowui: duplicate pie chart data key %q", source.key))
		}
		seen[source.key] = struct{}{}
		label := source.label
		if label == "" {
			label = source.key
		}
		itemColor := source.color
		if !source.hasColor {
			colors := activeTheme.Components.PieChart.SeriesColors
			itemColor = colors[index%len(colors)]
		}
		item := resolvedSlice{index: index, key: source.key, label: label, value: source.value, color: itemColor, hidden: source.hidden}
		result.legend = append(result.legend, item)
		if source.hidden || !finiteNonnegative(source.value) {
			continue
		}
		result.total += source.value
		result.slices = append(result.slices, item)
	}

	allocateAngles(result.slices, result.total, result.start, result.dir, widget.padAngle*math.Pi/180, widget.minAngle*math.Pi/180, widget.stillShowZeroSum)
	allocatePercents(result.slices, result.total)
	return result
}

func allocatePercents(slices []resolvedSlice, sum float64) {
	if sum == 0 || len(slices) == 0 {
		return
	}
	const digits = 100
	const target = 100 * digits
	seats := make([]int, len(slices))
	remainders := make([]float64, len(slices))
	current := 0
	for index, slice := range slices {
		votes := slice.value / sum * target
		seats[index] = int(math.Floor(votes))
		remainders[index] = votes - float64(seats[index])
		current += seats[index]
	}
	for current < target {
		largest := 0
		for index := 1; index < len(remainders); index++ {
			if remainders[index] > remainders[largest] {
				largest = index
			}
		}
		seats[largest]++
		remainders[largest] = 0
		current++
	}
	for index := range slices {
		slices[index].percent = float64(seats[index]) / digits
	}
}

func allocateAngles(slices []resolvedSlice, sum float64, start, dir, padAngle, minAngle float32, stillShowZeroSum bool) {
	if len(slices) == 0 {
		return
	}
	angleRange := float32(fullCircle)
	minAndPad := minAngle + padAngle
	restAngle := angleRange
	valueSumLargerThanMin := float64(0)
	unitAngle := angleRange / float32(len(slices))
	if sum != 0 {
		unitAngle = angleRange / float32(sum)
	}
	halfPad := dir * padAngle / 2
	current := start

	for index := range slices {
		angle := float32(slices[index].value) * unitAngle
		if sum == 0 && stillShowZeroSum {
			angle = unitAngle
		}
		if angle < minAndPad {
			angle = minAndPad
			restAngle -= minAndPad
		} else {
			valueSumLargerThanMin += slices[index].value
		}
		slices[index].rawAngle = angle
		slices[index].startAngle, slices[index].endAngle = displayedAngles(current, angle, dir, halfPad, padAngle)
		current += dir * angle
	}

	if restAngle >= angleRange || len(slices) == 0 {
		return
	}
	if restAngle <= 1e-3 || valueSumLargerThanMin == 0 {
		angle := angleRange / float32(len(slices))
		for index := range slices {
			slices[index].rawAngle = angle
			slices[index].startAngle, slices[index].endAngle = displayedAngles(start+dir*float32(index)*angle, angle, dir, halfPad, padAngle)
		}
		return
	}

	unitAngle = restAngle / float32(valueSumLargerThanMin)
	current = start
	for index := range slices {
		angle := slices[index].rawAngle
		if angle != minAndPad {
			angle = float32(slices[index].value) * unitAngle
		}
		slices[index].rawAngle = angle
		slices[index].startAngle, slices[index].endAngle = displayedAngles(current, angle, dir, halfPad, padAngle)
		current += dir * angle
	}
}

func displayedAngles(current, angle, dir, halfPad, padAngle float32) (float32, float32) {
	if padAngle > angle {
		middle := current + dir*angle/2
		return middle, middle
	}
	return current + halfPad, current + dir*angle - halfPad
}

func finiteNonnegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
