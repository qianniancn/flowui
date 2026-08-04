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

// Use applies theme tokens to the style declaration.
func Use(tokens ...StyleToken) Style {
	return styleStart().Use(tokens...)
}

// Width sets the preferred width of the styled box.
func Width(value unit.Dp) Style {
	return styleStart().Width(value)
}

// Height sets the preferred height of the styled box.
func Height(value unit.Dp) Style {
	return styleStart().Height(value)
}

// MinWidth sets the minimum width constraint.
func MinWidth(value unit.Dp) Style {
	return styleStart().MinWidth(value)
}

// MaxWidth sets the maximum width constraint.
func MaxWidth(value unit.Dp) Style {
	return styleStart().MaxWidth(value)
}

// MinHeight sets the minimum height constraint.
func MinHeight(value unit.Dp) Style {
	return styleStart().MinHeight(value)
}

// MaxHeight sets the maximum height constraint.
func MaxHeight(value unit.Dp) Style {
	return styleStart().MaxHeight(value)
}

// FillWidth makes the styled child consume the available width.
func FillWidth() Style {
	return styleStart().FillWidth()
}

// FillHeight makes the styled child consume the available height.
func FillHeight() Style {
	return styleStart().FillHeight()
}

// Padding sets equal padding on all four sides.
func Padding(value unit.Dp) Style {
	return styleStart().Padding(value)
}

// PaddingX sets the horizontal padding.
func PaddingX(value unit.Dp) Style {
	return styleStart().PaddingX(value)
}

// PaddingY sets the vertical padding.
func PaddingY(value unit.Dp) Style {
	return styleStart().PaddingY(value)
}

// PaddingTop sets the top padding.
func PaddingTop(value unit.Dp) Style {
	return styleStart().PaddingTop(value)
}

// PaddingRight sets the right padding.
func PaddingRight(value unit.Dp) Style {
	return styleStart().PaddingRight(value)
}

// PaddingBottom sets the bottom padding.
func PaddingBottom(value unit.Dp) Style {
	return styleStart().PaddingBottom(value)
}

// PaddingLeft sets the left padding.
func PaddingLeft(value unit.Dp) Style {
	return styleStart().PaddingLeft(value)
}

// Margin sets equal margin on all four sides.
func Margin(value unit.Dp) Style {
	return styleStart().Margin(value)
}

// MarginX sets the horizontal margin.
func MarginX(value unit.Dp) Style {
	return styleStart().MarginX(value)
}

// MarginY sets the vertical margin.
func MarginY(value unit.Dp) Style {
	return styleStart().MarginY(value)
}

// MarginTop sets the top margin.
func MarginTop(value unit.Dp) Style {
	return styleStart().MarginTop(value)
}

// MarginRight sets the right margin.
func MarginRight(value unit.Dp) Style {
	return styleStart().MarginRight(value)
}

// MarginBottom sets the bottom margin.
func MarginBottom(value unit.Dp) Style {
	return styleStart().MarginBottom(value)
}

// MarginLeft sets the left margin.
func MarginLeft(value unit.Dp) Style {
	return styleStart().MarginLeft(value)
}

// Overflow controls whether child content may paint outside the styled box.
func Overflow(value StyleOverflow) Style {
	return styleStart().Overflow(value)
}

// Cursor sets the pointer cursor while it is over the styled widget.
func Cursor(value pointer.Cursor) Style {
	return styleStart().Cursor(value)
}

// Background sets a solid color or gradient paint source.
func Background(source PaintSource) Style {
	return styleStart().Background(source)
}

// BackgroundNone clears the background paint source.
func BackgroundNone() Style {
	return styleStart().BackgroundNone()
}

// BorderColor sets the border color.
func BorderColor(value ColorSource) Style {
	return styleStart().BorderColor(value)
}

// BorderWidth sets the border width.
func BorderWidth(value unit.Dp) Style {
	return styleStart().BorderWidth(value)
}

// BorderBottomColor sets the bottom border color without changing the other
// sides.
func BorderBottomColor(value ColorSource) Style {
	return styleStart().BorderBottomColor(value)
}

// BorderBottomWidth sets the bottom border width without changing the other
// sides.
func BorderBottomWidth(value unit.Dp) Style {
	return styleStart().BorderBottomWidth(value)
}

// Radius sets equal corner radii on all four corners.
func Radius(value unit.Dp) Style {
	return styleStart().Radius(value)
}

// RadiusTopLeft sets the top-left corner radius.
func RadiusTopLeft(value unit.Dp) Style {
	return styleStart().RadiusTopLeft(value)
}

// RadiusTopRight sets the top-right corner radius.
func RadiusTopRight(value unit.Dp) Style {
	return styleStart().RadiusTopRight(value)
}

// RadiusBottomRight sets the bottom-right corner radius.
func RadiusBottomRight(value unit.Dp) Style {
	return styleStart().RadiusBottomRight(value)
}

// RadiusBottomLeft sets the bottom-left corner radius.
func RadiusBottomLeft(value unit.Dp) Style {
	return styleStart().RadiusBottomLeft(value)
}

// BoxShadow adds a custom box shadow to the styled box.
func BoxShadow(offsetX, offsetY, blur, spread unit.Dp, col ColorSource) Style {
	return styleStart().BoxShadow(offsetX, offsetY, blur, spread, col)
}

// Shadow adds one of the theme shadow profiles to the styled box.
func Shadow(profile ShadowProfile) Style {
	return styleStart().Shadow(profile)
}

// BoxShadowNone removes box shadows from the style declaration.
func BoxShadowNone() Style {
	return styleStart().BoxShadowNone()
}

// Outline sets the focus or emphasis outline around the styled box.
func Outline(width, offset unit.Dp, col ColorSource) Style {
	return styleStart().Outline(width, offset, col)
}

// Opacity sets the paint opacity, normally in the range [0, 1].
func Opacity(value float32) Style {
	return styleStart().Opacity(value)
}

// TextColor sets the text color source.
func TextColor(value ColorSource) Style {
	return styleStart().TextColor(value)
}

// FontSize sets the text size.
func FontSize(value unit.Sp) Style {
	return styleStart().FontSize(value)
}

// FontWeight sets the numeric font weight.
func FontWeight(value int) Style {
	return styleStart().FontWeight(value)
}

// Typeface sets the font family.
func Typeface(value font.Typeface) Style {
	return styleStart().Typeface(value)
}

// FontStyle sets the font style, such as italic.
func FontStyle(value font.Style) Style {
	return styleStart().FontStyle(value)
}

// LineHeight sets the absolute line height.
func LineHeight(value unit.Sp) Style {
	return styleStart().LineHeight(value)
}

// LineHeightScale sets line height relative to the font size.
func LineHeightScale(value float32) Style {
	return styleStart().LineHeightScale(value)
}

// MaxLines limits the number of rendered text lines.
func MaxLines(value int) Style {
	return styleStart().MaxLines(value)
}

// TextAlign sets the alignment of text within its layout width.
func TextAlign(value StyleTextAlign) Style {
	return styleStart().TextAlign(value)
}

// TextWrap sets the wrapping policy for text in a style declaration.
func TextWrap(value giotext.WrapPolicy) Style {
	return styleStart().Wrap(value)
}

// Truncator sets the string used when text is truncated.
func Truncator(value string) Style {
	return styleStart().Truncator(value)
}

// Translate offsets the styled content in dp.
func Translate(x, y unit.Dp) Style {
	return styleStart().Translate(x, y)
}

// Scale scales the styled content around its origin.
func Scale(x, y float32) Style {
	return styleStart().Scale(x, y)
}

// Rotate rotates the styled content by radians.
func Rotate(radians float32) Style {
	return styleStart().Rotate(radians)
}

// Transition enables animated changes for a style property.
func Transition(property PropertyID, duration time.Duration, options ...TransitionOption) Style {
	return styleStart().Transition(property, duration, options...)
}

// When conditionally applies override when predicate matches the style state.
func When(predicate Condition, override Style) Style {
	return styleStart().When(predicate, override)
}

// Part applies override to a named part of a compound component.
func Part(part StylePart, override Style) Style {
	return styleStart().Part(part, override)
}
