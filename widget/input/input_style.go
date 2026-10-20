package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

var inputEditorStyleParts = [...]style.Part{
	style.PartRoot,
	style.PartPlaceholder,
	style.PartSelection,
}

var inputEditorRootStyleParts = [...]style.Part{style.PartRoot}

// Text areas and multiline groups resolve a provisional zero-height visual
// before requesting the row-derived height. Keep both plans so the normal
// frame does not rebuild and revalidate the provider bundle twice.
const inputVisualCacheCapacity = 8

type inputEditorStyleResolution struct {
	ctx              *core.Context
	gtx              layout.Context
	key              string
	state            style.State
	defaults, custom style.Style
}

func resolveInputEditorPartsInto(request inputEditorStyleResolution, parts []style.Part, destination []style.ResolvedStyle) {
	computed.ResolvePartsInto(
		request.ctx,
		request.gtx,
		request.key,
		request.state,
		parts,
		destination,
		request.defaults,
		style.Style{},
		style.Style{},
		request.custom,
	)
}

func resolveInputEditorStyles(request inputEditorStyleResolution) (root, placeholder, selection style.ResolvedStyle) {
	var resolved [len(inputEditorStyleParts)]style.ResolvedStyle
	resolveInputEditorPartsInto(request, inputEditorStyleParts[:], resolved[:])
	return resolved[0], resolved[1], resolved[2]
}

func resolveInputEditorRootStyle(request inputEditorStyleResolution) style.ResolvedStyle {
	var resolved [len(inputEditorRootStyleParts)]style.ResolvedStyle
	resolveInputEditorPartsInto(request, inputEditorRootStyleParts[:], resolved[:])
	return resolved[0]
}

// inputVisualKey is the complete request identity for one InputVisual bundle.
// The provider cache owns theme identity, hit/miss accounting, and commit
// semantics; this key contains only the widget-owned request dimensions.
type inputVisualKey struct {
	variant       field.Variant
	sizeTier      Size
	fullWidth     bool
	dynamicHeight unit.Dp
	overrideHash  uint64
	overrideSet   bool
}

func inputOverrideKey(override *theme.InputOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	hash := override.Declaration.Hash64()
	if hash == (style.Style{}).Hash64() {
		return 0, false
	}
	return hash, true
}

func (s *inputState) visualDeclarations(i Widget, activeTheme *theme.Theme, instanceOverride *theme.InputOverride) theme.InputVisual {
	overrideHash, overrideSet := inputOverrideKey(instanceOverride)
	key := inputVisualKey{
		variant: i.variant, sizeTier: i.size, fullWidth: i.fullWidth,
		overrideHash: overrideHash, overrideSet: overrideSet,
	}
	return s.styleCache.Resolve(activeTheme, key, func() theme.InputVisual {
		return inputVisualDeclarations(activeTheme, i.variant, i.fullWidth, theme.InputModeSingleLine, i.size, instanceOverride)
	})
}

func inputVisualDeclarations(activeTheme *theme.Theme, variant Variant, fullWidth bool, mode theme.InputStyleMode, sizeTier Size, override *theme.InputOverride, heights ...unit.Dp) theme.InputVisual {
	var height unit.Dp
	if len(heights) > 0 {
		height = heights[0]
	}
	return theme.ResolveVisual(activeTheme, theme.InputRequest{
		Variant:   theme.InputVariant(variant),
		Mode:      mode,
		Size:      theme.InputSize(sizeTier),
		FullWidth: fullWidth,
		Height:    height,
		Override:  override,
	}, theme.ComponentVisualProvider.Input, theme.ValidateInputVisual)
}

func (s *textAreaState) visualDeclarations(t TextAreaWidget, activeTheme *theme.Theme, height unit.Dp, instanceOverride *theme.InputOverride) theme.InputVisual {
	s.styleCache.SetCapacity(inputVisualCacheCapacity)
	overrideHash, overrideSet := inputOverrideKey(instanceOverride)
	key := inputVisualKey{
		variant: t.variant, fullWidth: t.fullWidth, dynamicHeight: height,
		overrideHash: overrideHash, overrideSet: overrideSet,
	}
	return s.styleCache.Resolve(activeTheme, key, func() theme.InputVisual {
		return inputVisualDeclarations(activeTheme, Variant(t.variant), t.fullWidth, theme.InputModeTextArea, Medium, instanceOverride, height)
	})
}

func (s *inputGroupState) visualDeclarations(g GroupWidget, activeTheme *theme.Theme, minHeight unit.Dp, instanceOverride *theme.InputOverride) theme.InputVisual {
	s.styleCache.SetCapacity(inputVisualCacheCapacity)
	overrideHash, overrideSet := inputOverrideKey(instanceOverride)
	key := inputVisualKey{
		variant: g.variant, fullWidth: g.fullWidth, dynamicHeight: minHeight,
		overrideHash: overrideHash, overrideSet: overrideSet,
	}
	return s.styleCache.Resolve(activeTheme, key, func() theme.InputVisual {
		return inputVisualDeclarations(activeTheme, g.variant, g.fullWidth, theme.InputModeGroup, Medium, instanceOverride, minHeight)
	})
}
