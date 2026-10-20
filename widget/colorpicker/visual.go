package colorpicker

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type colorPickerVisualCache struct {
	cache computed.ProviderCache[struct{}, theme.ColorPickerVisual]
}

const (
	standaloneColorPickerVisualKey  = "colorpicker:standalone"
	standaloneColorPickerVisualSlot = "visual"
)

func (cache *colorPickerVisualCache) resolve(
	active *theme.Theme,
	request theme.ColorPickerRequest,
) theme.ColorPickerVisual {
	if request.Override != nil || cache == nil {
		return resolveColorPickerVisual(active, request)
	}
	return cache.cache.Resolve(active, struct{}{}, func() theme.ColorPickerVisual {
		return resolveColorPickerVisual(active, request)
	})
}

func resolveColorPickerVisual(
	active *theme.Theme,
	request theme.ColorPickerRequest,
) theme.ColorPickerVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.ColorPicker, theme.ValidateColorPickerVisual)
}

func resolveStandaloneColorPickerVisual(ctx *core.Context, override *theme.ColorPickerOverride) theme.ColorPickerVisual {
	cache := core.UseState[colorPickerVisualCache](ctx, standaloneColorPickerVisualKey, standaloneColorPickerVisualSlot)
	return cache.resolve(core.ActiveTheme(ctx), colorPickerRequest(override))
}

func colorPickerRequest(override *theme.ColorPickerOverride) theme.ColorPickerRequest {
	if override == nil {
		return theme.ColorPickerRequest{}
	}
	return theme.ColorPickerRequest{Override: override}
}

func resolveColorPickerVisualWithOverride(ctx *core.Context, state *colorPickerState, override *theme.ColorPickerOverride) theme.ColorPickerVisual {
	return state.visualCache.resolve(core.ActiveTheme(ctx), colorPickerRequest(override))
}

func colorPickerHSV(value hsvColor) theme.ColorPickerHSV {
	return theme.ColorPickerHSV{
		Hue:        value.h,
		Saturation: value.s,
		Value:      value.v,
		Alpha:      value.a,
	}
}
