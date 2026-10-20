package selectfield

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type selectResolvedStyle struct {
	content     style.ResolvedStyle
	placeholder style.ResolvedStyle
	selection   style.ResolvedStyle
	indicator   style.ResolvedStyle
	icon        style.ResolvedStyle
	description style.ResolvedStyle
	visual      theme.SelectVisual
}

func (s *selectState) partTransitionKey(ctx *core.Context, owner string, part style.Part) string {
	if s.styleKeyOwner != owner {
		s.styleKeyOwner = owner
		s.stylePartKeys = [6]string{}
	}
	index := -1
	switch part {
	case theme.SelectPartTrigger:
		index = 0
	case theme.SelectPartPlaceholder:
		index = 1
	case theme.SelectPartSelection:
		index = 2
	case theme.SelectPartIndicator:
		index = 3
	case theme.SelectPartIcon:
		index = 4
	case theme.SelectPartDescription:
		index = 5
	}
	if index < 0 {
		return core.DerivedKey(ctx, owner, "select-part:"+string(part))
	}
	if s.stylePartKeys[index] == "" {
		s.stylePartKeys[index] = core.DerivedKey(ctx, owner, "select-part:"+string(part))
	}
	return s.stylePartKeys[index]
}

func selectVisualBundle(activeTheme *theme.Theme, variant SelectVariant, override *theme.SelectOverride) selectDeclarationBundle {
	visual := theme.ResolveVisual(activeTheme, theme.SelectRequest{
		Variant:  theme.InputVariant(variant),
		Override: override,
	}, theme.ComponentVisualProvider.Select, theme.ValidateSelectVisual)
	return selectDeclarationBundle{
		visual: visual,
	}
}

type selectDeclarationBundle struct {
	visual theme.SelectVisual
}

// selectStyleDeclarationsCached resolves the provider-owned base bundle. An
// instance override deliberately bypasses this cache at the call site so it
// can never pollute the unstyled declaration used by a later frame.
func selectStyleDeclarationsCached(
	activeTheme *theme.Theme,
	variant SelectVariant,
	cache ...*computed.ProviderCache[SelectVariant, selectDeclarationBundle],
) selectDeclarationBundle {
	if len(cache) > 0 && cache[0] != nil {
		return cache[0].Resolve(activeTheme, variant, func() selectDeclarationBundle {
			return selectVisualBundle(activeTheme, variant, nil)
		})
	}
	return selectVisualBundle(activeTheme, variant, nil)
}

func (s Widget) resolveTriggerStyle(ctx *core.Context, gtx layout.Context, state *selectState, open bool, instanceStyle style.Style, bundles ...selectDeclarationBundle) selectResolvedStyle {
	activeTheme := core.ActiveTheme(ctx)
	key := s.resolvedKey(ctx, state)
	focused := gtx.Focused(&state.trigger)
	styleState := style.State{
		Hovered: state.trigger.Hovered(), Pressed: state.trigger.Pressed(),
		Focused: focused, FocusVisible: core.FocusVisible(ctx, &state.trigger, focused),
		Disabled: s.disabled || !gtx.Enabled(), Invalid: s.invalid,
		Selected: s.selectedKey != "" || len(s.selectedKeys) > 0, Open: open,
	}
	var declarations selectDeclarationBundle
	if len(bundles) > 0 {
		declarations = bundles[0]
	} else {
		declarations = selectStyleDeclarationsCached(activeTheme, s.variant, &state.styleCache)
	}
	defaults := declarations.visual.Defaults
	resolvePart := func(part style.Part) style.ResolvedStyle {
		// Keep one animation identity per semantic Part. ResolvePartWithKey can
		// share settled transition-free branches instead of cloning every
		// cached declaration on each open frame.
		partKey := state.partTransitionKey(ctx, key, part)
		return computed.ResolvePartWithKey(ctx, gtx, partKey, part, styleState, defaults, style.Style{}, style.Style{}, instanceStyle)
	}
	return selectResolvedStyle{
		visual:      declarations.visual,
		content:     resolvePart(theme.SelectPartTrigger),
		placeholder: resolvePart(theme.SelectPartPlaceholder),
		selection:   resolvePart(theme.SelectPartSelection),
		indicator:   resolvePart(theme.SelectPartIndicator),
		icon:        resolvePart(theme.SelectPartIcon),
		description: resolvePart(theme.SelectPartDescription),
	}
}
