package selectfield

import (
	"strings"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/listbox"
)

type SelectItem = listbox.Item
type SelectSection = listbox.Section
type SelectVariant = field.Variant

const (
	SelectPrimary   SelectVariant = field.Primary
	SelectSecondary SelectVariant = field.Secondary
)

type SelectSelectionMode int

const (
	SelectSelectionSingle SelectSelectionMode = iota
	SelectSelectionMultiple
)

type Widget struct {
	override            *theme.SelectOverride
	key                 string
	selectedKey         string
	hasSelectedKey      bool
	defaultSelectedKey  string
	hasDefaultSelected  bool
	selectedKeys        []string
	hasSelectedKeys     bool
	defaultSelectedKeys []string
	hasDefaultSelecteds bool
	items               []SelectItem
	sections            []SelectSection
	dataVersion         uint64
	hasDataVersion      bool
	placeholder         string
	hasPlaceholder      bool
	emptyText           string
	hasEmptyText        bool
	label               string
	description         string
	errorMessage        string
	valueText           string
	indicator           core.Widget
	onChange            func(string)
	onSelectionChange   func([]string)
	onOpenChange        func(bool)
	selectionMode       SelectSelectionMode
	disabledKeys        []string
	variant             SelectVariant
	open                bool
	hasOpen             bool
	defaultOpen         bool
	hasDefaultOpen      bool
	placement           core.PopoverPlacement
	shouldFlip          bool
	hasShouldFlip       bool
	avoidOverflow       bool
	hasAvoidOverflow    bool
	disabled            bool
	invalid             bool
	required            bool
	fullWidth           bool
}

// New creates a select field. It starts uncontrolled with no selection;
// SelectedKey sets the controlled value and DefaultSelectedKey the initial one.
func New(items []SelectItem) Widget {
	return Widget{
		key:           core.AutoKey(),
		items:         items,
		selectionMode: SelectSelectionSingle,
	}
}

// NewSections creates a select field from grouped sections.
func NewSections(sections []SelectSection) Widget {
	return New(nil).Sections(sections)
}

// Key sets an explicit identity for the select state.
func (s Widget) Key(key string) Widget {
	s.key = key
	return s
}

// Uncontrolled uses the current selection only as the initial internal value.
func (s Widget) Uncontrolled() Widget {
	if s.selectionMode == SelectSelectionMultiple {
		s.defaultSelectedKeys = append([]string(nil), s.selectedKeys...)
		s.hasDefaultSelecteds = true
		s.hasSelectedKeys = false
	} else {
		s.defaultSelectedKey = s.selectedKey
		s.hasDefaultSelected = true
		s.hasSelectedKey = false
	}
	return s
}

func (s Widget) Placeholder(placeholder string) Widget {
	s.placeholder = placeholder
	s.hasPlaceholder = true
	return s
}

// DataVersion enables item validation and flattened-data reuse. Increase
// version whenever the item data or section structure changes.
func (s Widget) DataVersion(version uint64) Widget {
	s.dataVersion = version
	s.hasDataVersion = true
	return s
}

func (s Widget) EmptyText(text string) Widget {
	s.emptyText = text
	s.hasEmptyText = true
	return s
}

func (s Widget) resolvedPlaceholder(ctx *core.Context) string {
	if s.hasPlaceholder {
		return s.placeholder
	}
	if s.selectionMode == SelectSelectionMultiple {
		return ctx.Text(i18n.SelectPlaceholderItems)
	}
	return ctx.Text(i18n.SelectPlaceholderItem)
}

func (s Widget) Label(label string) Widget {
	s.label = label
	return s
}

func (s Widget) Description(description string) Widget {
	s.description = description
	return s
}

func (s Widget) ErrorMessage(message string) Widget {
	s.errorMessage = message
	return s
}

func (s Widget) ValueText(text string) Widget {
	s.valueText = text
	return s
}

func (s Widget) Indicator(indicator core.Widget) Widget {
	s.indicator = indicator
	return s
}

func (s Widget) Sections(sections []SelectSection) Widget {
	s.sections = sections
	return s
}

func (s Widget) OnChange(fn func(string)) Widget {
	s.onChange = fn
	return s
}

func (s Widget) SelectedKey(key string) Widget {
	s.selectedKey = key
	s.hasSelectedKey = true
	return s
}

func (s Widget) DefaultSelectedKey(key string) Widget {
	s.defaultSelectedKey = key
	s.hasDefaultSelected = true
	s.hasSelectedKey = false
	return s
}

// SelectedKeys sets the controlled selection and switches to multiple selection.
func (s Widget) SelectedKeys(keys []string) Widget {
	s.selectedKeys = keys
	s.hasSelectedKeys = true
	s.selectionMode = SelectSelectionMultiple
	return s
}

func (s Widget) DefaultSelectedKeys(keys []string) Widget {
	s.defaultSelectedKeys = keys
	s.hasDefaultSelecteds = true
	s.hasSelectedKeys = false
	return s
}

func (s Widget) OnSelectionChange(fn func([]string)) Widget {
	s.onSelectionChange = fn
	return s
}

func (s Widget) OnOpenChange(fn func(bool)) Widget {
	s.onOpenChange = fn
	return s
}

func (s Widget) Open(open bool) Widget {
	s.open = open
	s.hasOpen = true
	return s
}

func (s Widget) DefaultOpen(open bool) Widget {
	s.defaultOpen = open
	s.hasDefaultOpen = true
	return s
}

func (s Widget) Placement(placement core.PopoverPlacement) Widget {
	s.placement = placement
	return s
}

func (s Widget) ShouldFlip(shouldFlip bool) Widget {
	s.shouldFlip = shouldFlip
	s.hasShouldFlip = true
	return s
}

func (s Widget) AvoidOverflow(avoidOverflow bool) Widget {
	s.avoidOverflow = avoidOverflow
	s.hasAvoidOverflow = true
	return s
}

func (s Widget) DisabledKeys(keys []string) Widget {
	s.disabledKeys = keys
	return s
}

func (s Widget) Variant(variant SelectVariant) Widget {
	s.variant = variant
	return s
}

func (s Widget) Disabled(disabled bool) Widget {
	s.disabled = disabled
	return s
}

func (s Widget) Invalid(invalid bool) Widget {
	s.invalid = invalid
	return s
}

func (s Widget) Required(required bool) Widget {
	s.required = required
	return s
}

func (s Widget) FullWidth() Widget {
	s.fullWidth = true
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return s.layoutWithOverride(ctx, gtx, s.override)
}

func (s Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.SelectOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	state := selectStateFor(ctx, s.key)
	interactive := core.OverlayInteractive(ctx, core.OverlayLayerPopup, state.key)
	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)
	if s.label != "" {
		core.PrepareFieldLabel(ctx, state.key, s.label)
	}
	if message, isError := s.supportMessage(); message != "" && !isError {
		core.PrepareFieldDescription(ctx, state.key, message)
	}
	core.RegisterFieldFocus(ctx, state.key, &state.trigger, gtx.Enabled() && !s.disabled)

	// Bind disclosure state for open/close
	state.bind(s)

	// Bind and get current selection state
	state.bindSelected(s)
	if s.selectionMode == SelectSelectionSingle {
		s.selectedKey = state.currentSelectedKey(s)
	} else {
		s.selectedKeys = state.currentSelectedKeys(s)
	}

	open := state.isOpen(s)
	if s.disabled {
		state.open = false
		open = false
	}
	if !open || interactive {
		open = state.handleTrigger(ctx, gtx, s, open)
	}
	if open && !state.wasOpen {
		activateSelect(ctx, state)
	} else if !open {
		releaseSelect(ctx, state)
	}
	restoreFocus := !open && state.wasOpen && !s.disabled && !naturallyDisabled && !state.skipRestore
	state.observeOpen(ctx, open, false)
	if restoreFocus {
		core.AfterOverlays(ctx, func() {
			if core.OverlayTopmost(ctx, core.OverlayLayerPopup, state.key) || !core.HasTopOverlay(ctx) {
				core.RequestFocus(ctx, &state.trigger)
			}
		})
	}
	eventGtx := gtx
	if s.disabled {
		eventGtx = eventGtx.Disabled()
	}

	activeTheme := core.ActiveTheme(ctx)
	var declarations selectDeclarationBundle
	if instanceOverride == nil {
		declarations = selectStyleDeclarationsCached(activeTheme, s.variant, &state.styleCache)
	} else {
		declarations = selectVisualBundle(activeTheme, s.variant, instanceOverride)
	}
	progress := state.progress(gtx, open && !s.disabled, declarations.visual.Metrics, declarations.visual.Motion)
	focused := eventGtx.Focused(&state.trigger)
	override := declarations.visual.Override
	dims := layoutui.Styled(ctx, eventGtx, state.key, style.State{
		Hovered:      state.trigger.Hovered(),
		Pressed:      state.trigger.Pressed(),
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &state.trigger, focused),
		Disabled:     s.disabled || !gtx.Enabled(),
		Selected:     s.selectedKey != "" || len(s.selectedKeys) > 0,
		Invalid:      s.invalid,
		Open:         open,
	}, override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return s.layoutWithBundle(ctx, gtx, state, open, override, declarations)
	}))
	if progress == 0 {
		return dims
	}
	s.layoutPopoverWithBundle(ctx, state, state.triggerRect, open, progress, naturallyDisabled, declarations)
	return dims
}

func (s Widget) flipEnabled() bool {
	if !s.hasShouldFlip {
		return true
	}
	return s.shouldFlip
}

func (s Widget) overflowAvoidanceEnabled() bool {
	if !s.hasAvoidOverflow {
		return true
	}
	return s.avoidOverflow
}

func (s Widget) displayValue() (string, bool) {
	return s.displayValueItems(s.allItems())
}

func (s Widget) displayValueCached(state *selectState) (string, bool) {
	return s.displayValueItems(state.itemsFor(s))
}

func (s Widget) displayValueItems(items []SelectItem) (string, bool) {
	if s.valueText != "" {
		return s.valueText, true
	}
	if s.selectionMode == SelectSelectionMultiple {
		labels := s.selectedLabelsFrom(items)
		if len(labels) == 0 {
			return s.placeholder, false
		}
		return strings.Join(labels, ", "), true
	}
	if label, ok := s.selectedLabelFrom(items); ok {
		return label, true
	}
	return s.placeholder, false
}

func (s Widget) selectedLabelFrom(items []SelectItem) (string, bool) {
	for _, item := range items {
		if item.Key == s.selectedKey {
			return item.Label, true
		}
	}
	return "", false
}

func (s Widget) selectedLabelsFrom(items []SelectItem) []string {
	labels := make([]string, 0, len(s.selectedKeys))
	seen := make(map[string]struct{}, len(s.selectedKeys))
	for _, selectedKey := range s.selectedKeys {
		if _, ok := seen[selectedKey]; ok {
			continue
		}
		seen[selectedKey] = struct{}{}
		for _, item := range items {
			if item.Key == selectedKey {
				labels = append(labels, item.Label)
				break
			}
		}
	}
	return labels
}

func (s Widget) allItems() []SelectItem {
	if len(s.sections) == 0 {
		return s.items
	}
	count := 0
	for _, section := range s.sections {
		count += len(section.Items)
	}
	items := make([]SelectItem, 0, count)
	for _, section := range s.sections {
		items = append(items, section.Items...)
	}
	return items
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.SelectOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
