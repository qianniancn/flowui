package linechart

import (
	"math"
	"strconv"
	"strings"
)

func applyLineStacks(data *chartData, stackByX bool) {
	groups := make(map[string][]int)
	groupOrder := make([]string, 0)
	for index := range data.series {
		name := data.series[index].stack
		if name == "" {
			continue
		}
		if _, exists := groups[name]; !exists {
			groupOrder = append(groupOrder, name)
		}
		groups[name] = append(groups[name], index)
	}

	for _, name := range groupOrder {
		indices := append([]int(nil), groups[name]...)
		if data.series[indices[0]].stackOrder == StackSeriesDescending {
			for left, right := 0, len(indices)-1; left < right; left, right = left+1, right-1 {
				indices[left], indices[right] = indices[right], indices[left]
			}
		}
		lookups := lineStackLookups(data.series, indices, stackByX)
		for stackIndex, seriesIndex := range indices {
			series := &data.series[seriesIndex]
			if stackIndex > 0 {
				series.stackedOnSmooth = data.series[indices[stackIndex-1]].smooth
			}
			for pointIndex := range series.points {
				point := &series.points[pointIndex]
				if !point.valid {
					continue
				}
				for previousIndex := stackIndex - 1; previousIndex >= 0; previousIndex-- {
					previousSeriesIndex := indices[previousIndex]
					previousPoint, exists := lineStackPoint(data.series[previousSeriesIndex], lookups[previousSeriesIndex], pointIndex, point.X, stackByX)
					if !exists {
						continue
					}
					previousValue := math.NaN()
					if previousPoint.valid {
						previousValue = previousPoint.Y
					}
					if !lineStackCompatible(series.stackStrategy, point.rawY, previousValue) {
						continue
					}
					point.stackBase = previousValue
					point.hasStackBase = true
					point.Y = addLineStackValue(point.rawY, previousValue)
					point.valid = finite(point.Y)
					break
				}
			}
		}
	}
}

func lineStackLookups(series []resolvedSeries, indices []int, stackByX bool) map[int]map[uint64]int {
	lookups := make(map[int]map[uint64]int)
	if !stackByX {
		return lookups
	}
	for _, seriesIndex := range indices {
		lookup := make(map[uint64]int, len(series[seriesIndex].points))
		for pointIndex, point := range series[seriesIndex].points {
			if finite(point.X) {
				lookup[lineStackXKey(point.X)] = pointIndex
			}
		}
		lookups[seriesIndex] = lookup
	}
	return lookups
}

func lineStackPoint(series resolvedSeries, lookup map[uint64]int, pointIndex int, x float64, stackByX bool) (resolvedPoint, bool) {
	if stackByX {
		index, exists := lookup[lineStackXKey(x)]
		if !exists {
			return resolvedPoint{}, false
		}
		return series.points[index], true
	}
	if pointIndex < 0 || pointIndex >= len(series.points) {
		return resolvedPoint{}, false
	}
	return series.points[pointIndex], true
}

func lineStackXKey(value float64) uint64 {
	if value == 0 {
		return 0
	}
	return math.Float64bits(value)
}

func lineStackCompatible(strategy StackStrategy, value, previous float64) bool {
	switch strategy {
	case StackAll:
		return true
	case StackPositive:
		return previous > 0
	case StackNegative:
		return previous < 0
	default:
		return value >= 0 && previous > 0 || value <= 0 && previous < 0
	}
}

func addLineStackValue(first, second float64) float64 {
	sum := first + second
	precision := max(lineDecimalPrecision(first), lineDecimalPrecision(second))
	if precision > 20 {
		return sum
	}
	power := math.Pow10(precision)
	if !finite(power) || math.Abs(sum) > math.MaxFloat64/power {
		return sum
	}
	return math.Round(sum*power) / power
}

func lineDecimalPrecision(value float64) int {
	text := strings.ToLower(strconv.FormatFloat(value, 'g', -1, 64))
	exponentIndex := strings.IndexByte(text, 'e')
	exponent := 0
	significandEnd := len(text)
	if exponentIndex >= 0 {
		significandEnd = exponentIndex
		exponent, _ = strconv.Atoi(text[exponentIndex+1:])
	}
	dotIndex := strings.IndexByte(text[:significandEnd], '.')
	decimalLength := 0
	if dotIndex >= 0 {
		decimalLength = significandEnd - dotIndex - 1
	}
	return max(decimalLength-exponent, 0)
}
