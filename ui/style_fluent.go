package ui

import (
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	giotext "gioui.org/text"
	"gioui.org/unit"
)

func styleStart() Style {
	var declaration Style
	return declaration
}

func Use(tokens ...StyleToken) Style {
	return styleStart().Use(tokens...)
}

func Width(value unit.Dp) Style {
	return styleStart().Width(value)
}

func Height(value unit.Dp) Style {
	return styleStart().Height(value)
}

func MinWidth(value unit.Dp) Style {
	return styleStart().MinWidth(value)
}

func MaxWidth(value unit.Dp) Style {
	return styleStart().MaxWidth(value)
}

func MinHeight(value unit.Dp) Style {
	return styleStart().MinHeight(value)
}

func MaxHeight(value unit.Dp) Style {
	return styleStart().MaxHeight(value)
}

func FillWidth() Style {
	return styleStart().FillWidth()
}

func FillHeight() Style {
	return styleStart().FillHeight()
}

func Padding(value unit.Dp) Style {
	return styleStart().Padding(value)
}

func PaddingX(value unit.Dp) Style {
	return styleStart().PaddingX(value)
}

func PaddingY(value unit.Dp) Style {
	return styleStart().PaddingY(value)
}

func PaddingTop(value unit.Dp) Style {
	return styleStart().PaddingTop(value)
}

func PaddingRight(value unit.Dp) Style {
	return styleStart().PaddingRight(value)
}

func PaddingBottom(value unit.Dp) Style {
	return styleStart().PaddingBottom(value)
}

func PaddingLeft(value unit.Dp) Style {
	return styleStart().PaddingLeft(value)
}

func Margin(value unit.Dp) Style {
	return styleStart().Margin(value)
}

func MarginX(value unit.Dp) Style {
	return styleStart().MarginX(value)
}

func MarginY(value unit.Dp) Style {
	return styleStart().MarginY(value)
}

func MarginTop(value unit.Dp) Style {
	return styleStart().MarginTop(value)
}

func MarginRight(value unit.Dp) Style {
	return styleStart().MarginRight(value)
}

func MarginBottom(value unit.Dp) Style {
	return styleStart().MarginBottom(value)
}

func MarginLeft(value unit.Dp) Style {
	return styleStart().MarginLeft(value)
}

func Overflow(value StyleOverflow) Style {
	return styleStart().Overflow(value)
}

func Cursor(value pointer.Cursor) Style {
	return styleStart().Cursor(value)
}

func Background(source PaintSource) Style {
	return styleStart().Background(source)
}

func BackgroundNone() Style {
	return styleStart().BackgroundNone()
}

func BorderColor(value ColorSource) Style {
	return styleStart().BorderColor(value)
}

func BorderWidth(value unit.Dp) Style {
	return styleStart().BorderWidth(value)
}

func Radius(value unit.Dp) Style {
	return styleStart().Radius(value)
}

func RadiusTopLeft(value unit.Dp) Style {
	return styleStart().RadiusTopLeft(value)
}

func RadiusTopRight(value unit.Dp) Style {
	return styleStart().RadiusTopRight(value)
}

func RadiusBottomRight(value unit.Dp) Style {
	return styleStart().RadiusBottomRight(value)
}

func RadiusBottomLeft(value unit.Dp) Style {
	return styleStart().RadiusBottomLeft(value)
}

func BoxShadow(offsetX, offsetY, blur, spread unit.Dp, col ColorSource) Style {
	return styleStart().BoxShadow(offsetX, offsetY, blur, spread, col)
}

func Shadow(profile ShadowProfile) Style {
	return styleStart().Shadow(profile)
}

func BoxShadowNone() Style {
	return styleStart().BoxShadowNone()
}

func Outline(width, offset unit.Dp, col ColorSource) Style {
	return styleStart().Outline(width, offset, col)
}

func Opacity(value float32) Style {
	return styleStart().Opacity(value)
}

func TextColor(value ColorSource) Style {
	return styleStart().TextColor(value)
}

func FontSize(value unit.Sp) Style {
	return styleStart().FontSize(value)
}

func FontWeight(value int) Style {
	return styleStart().FontWeight(value)
}

func Typeface(value font.Typeface) Style {
	return styleStart().Typeface(value)
}

func FontStyle(value font.Style) Style {
	return styleStart().FontStyle(value)
}

func LineHeight(value unit.Sp) Style {
	return styleStart().LineHeight(value)
}

func LineHeightScale(value float32) Style {
	return styleStart().LineHeightScale(value)
}

func MaxLines(value int) Style {
	return styleStart().MaxLines(value)
}

func TextAlign(value StyleTextAlign) Style {
	return styleStart().TextAlign(value)
}

// TextWrap starts a text style declaration with a wrapping policy. Wrap is
// already the public wrapping layout constructor.
func TextWrap(value giotext.WrapPolicy) Style {
	return styleStart().Wrap(value)
}

func Truncator(value string) Style {
	return styleStart().Truncator(value)
}

func Translate(x, y unit.Dp) Style {
	return styleStart().Translate(x, y)
}

func Scale(x, y float32) Style {
	return styleStart().Scale(x, y)
}

func Rotate(radians float32) Style {
	return styleStart().Rotate(radians)
}

func Transition(property PropertyID, duration time.Duration, options ...TransitionOption) Style {
	return styleStart().Transition(property, duration, options...)
}

func When(predicate Condition, override Style) Style {
	return styleStart().When(predicate, override)
}

func Part(part StylePart, override Style) Style {
	return styleStart().Part(part, override)
}
