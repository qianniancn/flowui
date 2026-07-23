package runtime

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlot = "style-runtime"

// Resolve cascades component declarations, inherited scopes, and the
// instance style in that order, then resolves theme tokens and transitions.
func Resolve(
	ctx *frame.Context,
	gtx layout.Context,
	key string,
	state style.StyleState,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	resolved := ResolveStatic(ctx, state, defaults, variant, size, custom)
	return ApplyTransitions(ctx, gtx, key, resolved)
}

// ApplyTransitions animates a style that has already been resolved. Styles
// without transitions remain stateless.
func ApplyTransitions(ctx *frame.Context, gtx layout.Context, key string, resolved style.ResolvedStyle) style.ResolvedStyle {
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	animate(ctx, gtx, key, &resolved)
	return resolved
}

// ApplyOpacity applies a resolved Paint opacity to a concrete color.
func ApplyOpacity(value color.NRGBA, opacity *float32) color.NRGBA {
	if opacity == nil {
		return value
	}
	value.A = byte(float32(value.A)*min(max(*opacity, 0), 1) + .5)
	return value
}

// ApplyOutlineOpacity returns a copy whose outline alpha is multiplied by opacity.
func ApplyOutlineOpacity(value style.ResolvedStyle, opacity float32) style.ResolvedStyle {
	if value.Paint == nil || value.Paint.Outline == nil {
		return value
	}
	paintStyle := *value.Paint
	outline := *paintStyle.Outline
	if col, ok := Color(outline.Color); ok {
		col.A = byte(float32(col.A)*min(max(opacity, 0), 1) + .5)
		outline.Color = style.SolidColor{Color: col}
	}
	paintStyle.Outline = &outline
	value.Paint = &paintStyle
	return value
}

// HasTransitions reports whether any resolved style needs runtime state.
func HasTransitions(values ...style.ResolvedStyle) bool {
	for _, value := range values {
		if len(value.Transitions) != 0 {
			return true
		}
	}
	return false
}

// ResolveStatic performs the same cascade and token lookup without retaining
// transition state. It is used for measurement passes.
func ResolveStatic(
	ctx *frame.Context,
	state style.StyleState,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	layers := []style.Style{defaults}
	layers = append(layers, frame.ActiveInheritedStyles(ctx)...)
	layers = append(layers, variant, size)
	layers = append(layers, frame.ActiveStyles(ctx)...)
	layers = append(layers, custom)
	layers = resolveThemeMetrics(layers, frame.ActiveTheme(ctx))
	resolved := style.Cascade(state, layers...)
	resolveThemeColors(&resolved, frame.ActiveTheme(ctx))
	return resolved
}

// ResolvePart cascades and animates a named element inside a compound component.
func ResolvePart(
	ctx *frame.Context,
	gtx layout.Context,
	key string,
	part style.Part,
	state style.StyleState,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	if part == style.PartRoot {
		return Resolve(ctx, gtx, key, state, defaults, variant, size, custom)
	}
	resolved := ResolvePartStatic(ctx, part, state, defaults, variant, size, custom)
	return ApplyPartTransitions(ctx, gtx, key, part, resolved)
}

// ApplyPartTransitions animates an already-resolved named part under its
// component-owned identity.
func ApplyPartTransitions(ctx *frame.Context, gtx layout.Context, key string, part style.Part, resolved style.ResolvedStyle) style.ResolvedStyle {
	if part == style.PartRoot {
		return ApplyTransitions(ctx, gtx, key, resolved)
	}
	return ApplyTransitions(ctx, gtx, frame.DerivedKey(ctx, key, "style-part:"+string(part)), resolved)
}

// ResolvePartStatic resolves a named part without retaining transition state.
func ResolvePartStatic(
	ctx *frame.Context,
	part style.Part,
	state style.StyleState,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	if part == style.PartRoot {
		return ResolveStatic(ctx, state, defaults, variant, size, custom)
	}
	layers := []style.Style{defaults}
	layers = append(layers, frame.ActiveInheritedStyles(ctx)...)
	layers = append(layers, variant, size)
	layers = append(layers, frame.ActiveStyles(ctx)...)
	layers = append(layers, custom)
	layers = resolveThemeMetrics(layers, frame.ActiveTheme(ctx))
	resolved := style.CascadePart(state, part, layers...)
	resolveThemeColors(&resolved, frame.ActiveTheme(ctx))
	return resolved
}

func resolveThemeMetrics(layers []style.Style, activeTheme *theme.Theme) []style.Style {
	if activeTheme == nil {
		fallback := theme.DefaultTheme()
		activeTheme = &fallback
	}
	for index := range layers {
		layers[index] = style.ExpandTokens(layers[index], func(token style.StyleToken) style.Style {
			var declaration style.Style
			switch token {
			case style.TokenBodyFontSize:
				declaration = declaration.FontSize(activeTheme.Typography.BodySize)
			case style.TokenControlFontSize:
				declaration = declaration.FontSize(activeTheme.Typography.ControlSize)
			case style.TokenSmallFontSize:
				declaration = declaration.FontSize(activeTheme.Typography.SmallSize)
			case style.TokenControlRadius:
				declaration = declaration.Radius(activeTheme.Shape.ControlRadius)
			case style.TokenPopoverRadius:
				declaration = declaration.Radius(activeTheme.Shape.PopoverRadius)
			case style.TokenItemRadius:
				declaration = declaration.Radius(activeTheme.Shape.ItemRadius)
			case style.TokenCheckboxRadius:
				declaration = declaration.Radius(activeTheme.Shape.CheckboxRadius)
			case style.TokenControlHeight:
				declaration = declaration.Height(activeTheme.Spacing.ControlHeight)
			case style.TokenSmallControlHeight:
				declaration = declaration.Height(activeTheme.Spacing.SmallControlHeight)
			case style.TokenLargeControlHeight:
				declaration = declaration.Height(activeTheme.Spacing.LargeControlHeight)
			case style.TokenControlPaddingX:
				declaration = declaration.PaddingX(activeTheme.Spacing.ControlPaddingX)
			case style.TokenSmallControlPaddingX:
				declaration = declaration.PaddingX(activeTheme.Spacing.SmallControlPaddingX)
			case style.TokenLargeControlPaddingX:
				declaration = declaration.PaddingX(activeTheme.Spacing.LargeControlPaddingX)
			case style.TokenIconButtonSize:
				declaration = declaration.Width(activeTheme.Spacing.IconButtonSize).Height(activeTheme.Spacing.IconButtonSize)
			case style.TokenPanelPadding:
				declaration = declaration.Padding(activeTheme.Spacing.PanelPadding)
			case style.TokenItemHeight:
				declaration = declaration.Height(activeTheme.Spacing.ItemHeight)
			}
			return declaration
		})
	}
	return layers
}

func resolveThemeColors(resolved *style.ResolvedStyle, activeTheme *theme.Theme) {
	if resolved.Paint != nil {
		resolved.Paint.Background = resolvePaint(resolved.Paint.Background, activeTheme)
		if resolved.Paint.Border != nil && resolved.Paint.Border.Color != nil {
			resolved.Paint.Border.Color = resolveColor(resolved.Paint.Border.Color, activeTheme)
		}
		resolved.Paint.Shadows = resolveShadows(resolved.Paint.Shadows, activeTheme)
		if resolved.Paint.Outline != nil {
			resolved.Paint.Outline.Color = resolveColor(resolved.Paint.Outline.Color, activeTheme)
		}
	}
	if resolved.Text != nil && resolved.Text.Color != nil {
		resolved.Text.Color = resolveColor(resolved.Text.Color, activeTheme)
	}
}

func resolveShadows(shadows []style.Shadow, activeTheme *theme.Theme) []style.Shadow {
	if len(shadows) == 0 {
		return nil
	}
	result := make([]style.Shadow, 0, len(shadows))
	for _, shadow := range shadows {
		if shadow.Profile == nil {
			shadow.Color = resolveColor(shadow.Color, activeTheme)
			result = append(result, shadow)
			continue
		}
		profile, col := themeShadow(*shadow.Profile, activeTheme)
		for _, layer := range profile.Layers {
			layerColor := col
			layerColor.A = scaledAlpha(layerColor.A, layer.Opacity)
			if layer.Blur < 0 || layerColor.A == 0 {
				continue
			}
			result = append(result, style.Shadow{
				OffsetX: layer.OffsetX,
				OffsetY: layer.OffsetY,
				Blur:    layer.Blur,
				Spread:  layer.Spread,
				Color:   style.SolidColor{Color: layerColor},
			})
		}
	}
	return result
}

func themeShadow(profile style.ShadowProfile, activeTheme *theme.Theme) (theme.ShadowTheme, color.NRGBA) {
	if activeTheme == nil {
		fallback := theme.DefaultTheme()
		activeTheme = &fallback
	}
	if profile == style.ShadowOverlay {
		return activeTheme.Shadows.Overlay, activeTheme.Palette.OverlayShadowColor()
	}
	if profile == style.ShadowMenu {
		return activeTheme.Shadows.Menu, theme.ColorOr(activeTheme.Components.Menu.ShadowColor, activeTheme.Palette.OverlayShadowColor())
	}
	return activeTheme.Shadows.Surface, activeTheme.Palette.SurfaceShadow
}

func scaledAlpha(alpha uint8, opacity float32) uint8 {
	if opacity <= 0 {
		return 0
	}
	if opacity >= 1 {
		return alpha
	}
	return uint8(float32(alpha)*opacity + .5)
}

func resolvePaint(source style.PaintSource, activeTheme *theme.Theme) style.PaintSource {
	switch value := source.(type) {
	case style.ColorSource:
		return resolveColor(value, activeTheme)
	case style.StyleGradient:
		for index := range value.Stops {
			value.Stops[index].Color = resolveColor(value.Stops[index].Color, activeTheme)
		}
		return value
	case *style.StyleGradient:
		if value == nil {
			return nil
		}
		copy := *value
		copy.Stops = append([]style.StyleGradientStop(nil), value.Stops...)
		for index := range copy.Stops {
			copy.Stops[index].Color = resolveColor(copy.Stops[index].Color, activeTheme)
		}
		return &copy
	default:
		return source
	}
}

func resolveColor(source style.ColorSource, activeTheme *theme.Theme) style.SolidColor {
	if source == nil {
		return style.SolidColor{}
	}
	switch value := source.(type) {
	case style.SolidColor:
		return value
	case *style.SolidColor:
		if value == nil {
			return style.SolidColor{}
		}
		return *value
	case style.ThemeColor:
		return style.SolidColor{Color: tokenColor(activeTheme, value.Token)}
	case *style.ThemeColor:
		if value == nil {
			return style.SolidColor{}
		}
		return style.SolidColor{Color: tokenColor(activeTheme, value.Token)}
	case style.AlphaColor:
		resolved := resolveColor(value.Source, activeTheme)
		resolved.Color.A = value.Alpha
		return resolved
	case *style.AlphaColor:
		if value == nil {
			return style.SolidColor{}
		}
		resolved := resolveColor(value.Source, activeTheme)
		resolved.Color.A = value.Alpha
		return resolved
	default:
		return style.SolidColor{}
	}
}

func tokenColor(activeTheme *theme.Theme, token style.ColorToken) color.NRGBA {
	if activeTheme == nil {
		fallback := theme.DefaultTheme()
		activeTheme = &fallback
	}
	switch token {
	case style.ColorBackground:
		return activeTheme.Palette.Background
	case style.ColorSurface:
		return activeTheme.Palette.Surface
	case style.ColorSurfaceForeground:
		return activeTheme.Palette.SurfaceForeground
	case style.ColorSurfaceSecondary:
		return activeTheme.Palette.SurfaceSecondary
	case style.ColorSurfaceSecondaryForeground:
		return activeTheme.Palette.SurfaceSecondaryForeground
	case style.ColorSurfaceTertiary:
		return activeTheme.Palette.SurfaceTertiary
	case style.ColorSurfaceTertiaryForeground:
		return activeTheme.Palette.SurfaceTertiaryForeground
	case style.ColorSurfaceHover:
		return activeTheme.Palette.SurfaceHover
	case style.ColorSurfacePressed:
		return activeTheme.Palette.SurfacePressed
	case style.ColorSurfaceRaised:
		return activeTheme.Palette.SurfaceRaised
	case style.ColorOverlay:
		return activeTheme.Palette.OverlayColor()
	case style.ColorOverlayForeground:
		return activeTheme.Palette.OverlayForegroundColor()
	case style.ColorForeground:
		return activeTheme.Palette.Foreground
	case style.ColorMutedForeground:
		return activeTheme.Palette.MutedForeground
	case style.ColorBorder:
		return activeTheme.Palette.Border
	case style.ColorSeparator:
		return activeTheme.Palette.SeparatorColor()
	case style.ColorDefault:
		return activeTheme.Palette.DefaultColor()
	case style.ColorDefaultForeground:
		return activeTheme.Palette.DefaultForegroundColor()
	case style.ColorDefaultHover:
		return activeTheme.Palette.DefaultHoverColor()
	case style.ColorFieldBackground:
		return activeTheme.Palette.FieldBackgroundColor()
	case style.ColorFieldHover:
		return activeTheme.Palette.FieldHoverColor()
	case style.ColorFieldForeground:
		return activeTheme.Palette.FieldForegroundColor()
	case style.ColorFieldPlaceholder:
		return activeTheme.Palette.FieldPlaceholderColor()
	case style.ColorFieldFocus:
		return activeTheme.Palette.FieldFocusColor()
	case style.ColorSegment:
		return activeTheme.Palette.Segment
	case style.ColorSegmentForeground:
		return activeTheme.Palette.SegmentForeground
	case style.ColorAccent:
		return activeTheme.Palette.Accent
	case style.ColorAccentHover:
		return activeTheme.Palette.AccentHover
	case style.ColorAccentPressed:
		return activeTheme.Palette.AccentPressed
	case style.ColorAccentForeground:
		return activeTheme.Palette.AccentForeground
	case style.ColorAccentSoft:
		return activeTheme.Palette.AccentSoft
	case style.ColorAccentSoftHover:
		return activeTheme.Palette.AccentSoftHover
	case style.ColorAccentSoftForeground:
		return activeTheme.Palette.AccentSoftForeground
	case style.ColorSuccess:
		return activeTheme.Palette.Success
	case style.ColorSuccessForeground:
		return activeTheme.Palette.SuccessForeground
	case style.ColorSuccessSoft:
		return activeTheme.Palette.SuccessSoft
	case style.ColorSuccessSoftForeground:
		return activeTheme.Palette.SuccessSoftForegroundColor()
	case style.ColorWarning:
		return activeTheme.Palette.Warning
	case style.ColorWarningForeground:
		return activeTheme.Palette.WarningForeground
	case style.ColorWarningSoft:
		return activeTheme.Palette.WarningSoft
	case style.ColorWarningSoftForeground:
		return activeTheme.Palette.WarningSoftForegroundColor()
	case style.ColorDanger:
		return activeTheme.Palette.Danger
	case style.ColorDangerHover:
		return activeTheme.Palette.DangerHover
	case style.ColorDangerPressed:
		return activeTheme.Palette.DangerPressed
	case style.ColorDangerForeground:
		return activeTheme.Palette.DangerForeground
	case style.ColorDangerSoft:
		return activeTheme.Palette.DangerSoft
	case style.ColorDangerSoftHover:
		return activeTheme.Palette.DangerSoftHover
	case style.ColorDangerSoftForeground:
		return activeTheme.Palette.DangerSoftForeground
	case style.ColorFocus:
		return activeTheme.Palette.Focus
	case style.ColorSelection:
		return activeTheme.Palette.Selection
	case style.ColorSurfaceShadow:
		return activeTheme.Palette.SurfaceShadow
	case style.ColorOverlayShadow:
		return activeTheme.Palette.OverlayShadowColor()
	default:
		return activeTheme.Palette.Foreground
	}
}

type runtimeState struct {
	background colorAnimation
	text       colorAnimation
	border     colorAnimation
	outline    colorAnimation
	opacity    floatAnimation
	radius     floatAnimation
	radiusNW   floatAnimation
	radiusNE   floatAnimation
	radiusSE   floatAnimation
	radiusSW   floatAnimation
	translateX floatAnimation
	translateY floatAnimation
	scaleX     floatAnimation
	scaleY     floatAnimation
	rotate     floatAnimation
}

func animate(ctx *frame.Context, gtx layout.Context, key string, resolved *style.ResolvedStyle) {
	state := frame.UseState[runtimeState](ctx, key, stateSlot)
	motion := frame.ActiveTheme(ctx).Motion

	if resolved.Paint != nil {
		if value, ok := solidColor(resolved.Paint.Background); ok {
			value = state.background.value(gtx, value, transitionFor(*resolved, style.PropBackgroundColor), motion)
			resolved.Paint.Background = style.SolidColor{Color: value}
		}
		if resolved.Paint.Border != nil {
			if value, ok := solidColor(resolved.Paint.Border.Color); ok {
				value = state.border.value(gtx, value, transitionFor(*resolved, style.PropBorderColor), motion)
				resolved.Paint.Border.Color = style.SolidColor{Color: value}
			}
		}
		if resolved.Paint.Outline != nil {
			if value, ok := solidColor(resolved.Paint.Outline.Color); ok {
				value = state.outline.value(gtx, value, transitionFor(*resolved, style.PropOutlineColor), motion)
				resolved.Paint.Outline.Color = style.SolidColor{Color: value}
			}
		}
		if resolved.Paint.Opacity != nil {
			value := state.opacity.value(gtx, *resolved.Paint.Opacity, transitionFor(*resolved, style.PropOpacity), motion)
			resolved.Paint.Opacity = &value
		}
		if resolved.Paint.Radius != nil {
			value := state.radius.value(gtx, float32(*resolved.Paint.Radius), transitionFor(*resolved, style.PropRadius), motion)
			radius := unit.Dp(value)
			resolved.Paint.Radius = &radius
		}
		if resolved.Paint.Radii != nil {
			transition := transitionFor(*resolved, style.PropRadius)
			resolved.Paint.Radii.TopLeft = unit.Dp(state.radiusNW.value(gtx, float32(resolved.Paint.Radii.TopLeft), transition, motion))
			resolved.Paint.Radii.TopRight = unit.Dp(state.radiusNE.value(gtx, float32(resolved.Paint.Radii.TopRight), transition, motion))
			resolved.Paint.Radii.BottomRight = unit.Dp(state.radiusSE.value(gtx, float32(resolved.Paint.Radii.BottomRight), transition, motion))
			resolved.Paint.Radii.BottomLeft = unit.Dp(state.radiusSW.value(gtx, float32(resolved.Paint.Radii.BottomLeft), transition, motion))
		}
	}
	if resolved.Text != nil {
		if value, ok := solidColor(resolved.Text.Color); ok {
			value = state.text.value(gtx, value, transitionFor(*resolved, style.PropTextColor), motion)
			resolved.Text.Color = style.SolidColor{Color: value}
		}
	}
	if resolved.Trans != nil {
		transition := transitionFor(*resolved, style.PropTransform)
		if resolved.Trans.TranslateX != nil {
			value := unit.Dp(state.translateX.value(gtx, float32(*resolved.Trans.TranslateX), transition, motion))
			resolved.Trans.TranslateX = &value
		}
		if resolved.Trans.TranslateY != nil {
			value := unit.Dp(state.translateY.value(gtx, float32(*resolved.Trans.TranslateY), transition, motion))
			resolved.Trans.TranslateY = &value
		}
		if resolved.Trans.ScaleX != nil {
			value := state.scaleX.value(gtx, *resolved.Trans.ScaleX, transition, motion)
			resolved.Trans.ScaleX = &value
		}
		if resolved.Trans.ScaleY != nil {
			value := state.scaleY.value(gtx, *resolved.Trans.ScaleY, transition, motion)
			resolved.Trans.ScaleY = &value
		}
		if resolved.Trans.Rotate != nil {
			value := state.rotate.value(gtx, *resolved.Trans.Rotate, transition, motion)
			resolved.Trans.Rotate = &value
		}
	}
}

func solidColor(source any) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		return value.Color, true
	default:
		return color.NRGBA{}, false
	}
}

type transitionSpec struct {
	value style.Transition
	ok    bool
}

func transitionFor(resolved style.ResolvedStyle, property style.PropertyID) transitionSpec {
	for index := len(resolved.Transitions) - 1; index >= 0; index-- {
		transition := resolved.Transitions[index]
		if transition.Property == property {
			return transitionSpec{value: transition, ok: true}
		}
	}
	return transitionSpec{}
}

type colorAnimation struct {
	transition animation.ColorTransition
	target     color.NRGBA
	changedAt  time.Time
	ready      bool
}

func (a *colorAnimation) value(gtx layout.Context, target color.NRGBA, spec transitionSpec, motion theme.MotionTheme) color.NRGBA {
	if !a.ready {
		a.ready = true
		a.target = target
		return a.transition.Value(gtx, target, 0, nil)
	}
	if target != a.target {
		a.target = target
		a.changedAt = gtx.Now
		if delay := transitionDelay(spec, motion); delay > 0 {
			a.transition.Set(a.transition.Current(), target, gtx.Now.Add(delay))
		}
	}
	if !spec.ok {
		return a.transition.Value(gtx, target, 0, nil)
	}
	if waitForDelay(gtx, a.changedAt, spec.value.Delay, motion) {
		return a.transition.Current()
	}
	return a.transition.Value(gtx, target, spec.value.Duration, animation.Easing(spec.value.Ease), motion)
}

type floatAnimation struct {
	transition animation.FloatTransition
	target     float32
	changedAt  time.Time
	ready      bool
}

func (a *floatAnimation) value(gtx layout.Context, target float32, spec transitionSpec, motion theme.MotionTheme) float32 {
	if !a.ready {
		a.ready = true
		a.target = target
		return a.transition.Value(gtx, target, 0, nil)
	}
	if target != a.target {
		a.target = target
		a.changedAt = gtx.Now
		if delay := transitionDelay(spec, motion); delay > 0 {
			a.transition.Set(a.transition.Current(), target, gtx.Now.Add(delay))
		}
	}
	if !spec.ok {
		return a.transition.Value(gtx, target, 0, nil)
	}
	if waitForDelay(gtx, a.changedAt, spec.value.Delay, motion) {
		return a.transition.Current()
	}
	return a.transition.Value(gtx, target, spec.value.Duration, animation.Easing(spec.value.Ease), motion)
}

func waitForDelay(gtx layout.Context, changedAt time.Time, delay time.Duration, motion theme.MotionTheme) bool {
	delay = theme.ResolveMotionDuration(motion, delay)
	if delay <= 0 || changedAt.IsZero() || gtx.Now.Sub(changedAt) >= delay {
		return false
	}
	gtx.Execute(op.InvalidateCmd{})
	return true
}

func transitionDelay(spec transitionSpec, motion theme.MotionTheme) time.Duration {
	if !spec.ok {
		return 0
	}
	return theme.ResolveMotionDuration(motion, spec.value.Delay)
}
