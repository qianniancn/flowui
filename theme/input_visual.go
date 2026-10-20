package theme

import (
	"image/color"
	"math"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// InputVariant is the neutral field-surface vocabulary shared by input-like
// behavior widgets and style providers.
type InputVariant uint8

const (
	InputPrimary InputVariant = iota
	InputSecondary
	InputFilled
	InputBorderless
	InputUnderlined
)

// InputSize identifies the field geometry tier exposed to visual providers.
// Medium remains zero so existing input widgets keep their zero-value behavior.
type InputSize uint8

const (
	InputMedium InputSize = iota
	InputSmall
	InputLarge
)

// InputRequest describes the geometry and semantic field face needed by an
// input visual provider. Height is supplied for dynamic text-area sizing.
type InputRequest struct {
	Variant   InputVariant
	Mode      InputStyleMode
	Size      InputSize
	FullWidth bool
	Height    unit.Dp
	Override  *InputOverride
}

// InputOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type InputOverride struct {
	Declaration style.Style
}

// InputVisual contains the declaration layers consumed by input behavior.
// Variant and size are request axes resolved by the provider before this
// bundle is returned; unlike components with independent cascade layers,
// Input has no second declaration layer for either axis.
type InputVisual struct {
	Defaults    style.Style
	Override    style.Style
	EditorStyle InputEditorStyleResolver
}

// InputEditorStyle is the provider-owned bridge to Gio's stateful editor.
// Optional typography fields retain the distinction between an inherited
// material default and an explicit declaration.
type InputEditorStyle struct {
	TextSize           unit.Sp
	LineHeight         unit.Sp
	LineHeightScale    float32
	HasLineHeightScale bool
	Typeface           font.Typeface
	HasTypeface        bool
	FontStyle          font.Style
	HasFontStyle       bool
	FontWeight         font.Weight
	HasFontWeight      bool
	Foreground         color.NRGBA
	Placeholder        color.NRGBA
	Selection          color.NRGBA
}

// InputEditorStyleResolver resolves provider-owned editor colors and
// typography from the already-cascaded field parts.
type InputEditorStyleResolver func(content, placeholder, selection style.ResolvedStyle) InputEditorStyle

// InputOverrideStyle extracts the provider-owned instance layer from a
// request. Theme packages use this when assembling their visual bundle.
func InputOverrideStyle(request InputRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// InputProvider supplies input-like field visuals.
type InputProvider interface {
	Input(active *Theme, request InputRequest) InputVisual
}

// ValidateInputVisual checks the complete field declaration and editor bridge
// before a widget stores a provider result in its declaration cache. Input
// behavior cannot recover missing field parts or editor typography/colors
// without reintroducing visual interpretation.
func ValidateInputVisual(visual InputVisual) {
	if visual.EditorStyle == nil {
		invalidVisual("Input", "EditorStyle", "must be provided")
	}

	resolve := func(state style.State, part style.Part) style.ResolvedStyle {
		return style.CascadePart(state, part, visual.Defaults, visual.Override)
	}
	for _, state := range []style.State{
		{},
		{Hovered: true},
		{Focused: true},
		{FocusVisible: true},
		{Invalid: true},
		{Warning: true},
		{Disabled: true},
		{ReadOnly: true},
		{Focused: true, Invalid: true},
		{Focused: true, Warning: true},
		{Disabled: true, Invalid: true},
	} {
		root := resolve(state, style.PartRoot)
		if root.Box == nil {
			invalidVisual("Input", "Root.Box", "must be provided")
		}
		if !hasPositiveHeight(root.Box.Height, root.Box.MinHeight) {
			invalidVisual("Input", "Root.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if root.Paint == nil {
			invalidVisual("Input", "Root.Paint", "must be provided")
		}
		if !hasPaint(root.Paint.Background) {
			invalidVisual("Input", "Root.Paint.Background", "must provide a solid color or nonempty gradient")
		}
		if root.Text == nil {
			invalidVisual("Input", "Root.Text", "must be provided")
		}
		if !hasColor(root.Text.Color) {
			invalidVisual("Input", "Root.Text.Color", "must resolve to a solid color")
		}
		if root.Text.FontSize == nil {
			invalidVisual("Input", "Root.Text.FontSize", "must be provided")
		}
		if *root.Text.FontSize <= 0 {
			invalidVisual("Input", "Root.Text.FontSize", "must be greater than 0")
		}
		if root.Text.LineHeight == nil {
			invalidVisual("Input", "Root.Text.LineHeight", "must be provided")
		}
		if *root.Text.LineHeight <= 0 {
			invalidVisual("Input", "Root.Text.LineHeight", "must be greater than 0")
		}

		placeholder := resolve(state, style.PartPlaceholder)
		if placeholder.Text == nil {
			invalidVisual("Input", "Placeholder.Text", "must be provided")
		}
		if !hasColor(placeholder.Text.Color) {
			invalidVisual("Input", "Placeholder.Text.Color", "must resolve to a solid color")
		}

		selection := resolve(state, style.PartSelection)
		if selection.Paint == nil {
			invalidVisual("Input", "Selection.Paint", "must be provided")
		}
		if !hasPaint(selection.Paint.Background) {
			invalidVisual("Input", "Selection.Paint.Background", "must provide a solid color or nonempty gradient")
		}

		content := root
		groupContent := resolve(state, style.PartContent)
		if groupContent.Box != nil || groupContent.Paint != nil {
			content = groupContent
		}
		editor := visual.EditorStyle(content, placeholder, selection)
		if editor.Foreground.A == 0 {
			invalidVisual("Input", "Editor.Foreground.A", "must be nonzero")
		}
		if editor.Placeholder.A == 0 {
			invalidVisual("Input", "Editor.Placeholder.A", "must be nonzero")
		}
		if editor.Selection.A == 0 {
			invalidVisual("Input", "Editor.Selection.A", "must be nonzero")
		}
		if editor.TextSize <= 0 {
			invalidVisual("Input", "Editor.TextSize", "must be greater than 0")
		}
		if editor.LineHeight <= 0 {
			invalidVisual("Input", "Editor.LineHeight", "must be greater than 0")
		}
		if math.IsNaN(float64(editor.TextSize)) {
			invalidVisual("Input", "Editor.TextSize", "must be finite")
		}
		if math.IsInf(float64(editor.TextSize), 0) {
			invalidVisual("Input", "Editor.TextSize", "must be finite")
		}
		if math.IsNaN(float64(editor.LineHeight)) {
			invalidVisual("Input", "Editor.LineHeight", "must be finite")
		}
		if math.IsInf(float64(editor.LineHeight), 0) {
			invalidVisual("Input", "Editor.LineHeight", "must be finite")
		}
		if editor.HasLineHeightScale && (editor.LineHeightScale <= 0 ||
			math.IsNaN(float64(editor.LineHeightScale)) || math.IsInf(float64(editor.LineHeightScale), 0)) {
			invalidVisual("Input", "EditorStyle.LineHeightScale", "must be finite and greater than 0 when enabled")
		}
	}
}
