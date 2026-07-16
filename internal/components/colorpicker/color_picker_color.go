package colorpicker

import (
	"image/color"
	"strconv"
	"strings"
)

func formatHexColor(value color.NRGBA, alpha bool) string {
	const digits = "0123456789ABCDEF"
	var result [9]byte
	result[0] = '#'
	channels := [...]byte{value.R, value.G, value.B, value.A}
	count := 3
	if alpha {
		count = 4
	}
	for index, channel := range channels[:count] {
		result[index*2+1] = digits[channel>>4]
		result[index*2+2] = digits[channel&0x0f]
	}
	return string(result[:count*2+1])
}

func parseHexColor(value string, fallbackAlpha uint8) (color.NRGBA, bool) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "#"))
	switch len(value) {
	case 3, 4:
		var expanded strings.Builder
		for _, channel := range value {
			expanded.WriteRune(channel)
			expanded.WriteRune(channel)
		}
		value = expanded.String()
	case 6, 8:
	default:
		return color.NRGBA{}, false
	}

	parsed, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return color.NRGBA{}, false
	}
	if len(value) == 6 {
		return color.NRGBA{
			R: uint8(parsed >> 16),
			G: uint8(parsed >> 8),
			B: uint8(parsed),
			A: fallbackAlpha,
		}, true
	}
	return color.NRGBA{
		R: uint8(parsed >> 24),
		G: uint8(parsed >> 16),
		B: uint8(parsed >> 8),
		A: uint8(parsed),
	}, true
}
