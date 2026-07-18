package selects

import (
	"strings"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/listbox"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

type SelectItem = listbox.ListBoxItem
type SelectSection = listbox.ListBoxSection
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

type SelectWidget struct {
	key               string
	selectedKey       string
	selectedKeys      []string
	items             []SelectItem
	sections          []SelectSection
	dataVersion       uint64
	hasDataVersion    bool
	placeholder       string
	emptyText         string
	label             string
	description       string
	errorMessage      string
	valueText         string
	indicator         frame.Widget
	onChange          func(string)
	onSelectionChange func([]string)
	onOpenChange      func(bool)
	selectionMode     SelectSelectionMode
	disabledKeys      []string
	variant           SelectVariant
	open              bool
	hasOpen           bool
	defaultOpen       bool
	hasDefaultOpen    bool
	placement         overlay.PopoverPlacement
	shouldFlip        bool
	hasShouldFlip     bool
	avoidOverflow     bool
	hasAvoidOverflow  bool
	disabled          bool
	invalid           bool
	required          bool
	fullWidth         bool
}

const (
	selectEnterDuration     = 150 * time.Millisecond
	selectExitDuration      = 100 * time.Millisecond
	selectIndicatorDuration = 150 * time.Millisecond
)

func Select(key, selectedKey string, items []SelectItem) SelectWidget {
	return SelectWidget{
		key:           key,
		selectedKey:   selectedKey,
		items:         items,
		placeholder:   "Select an item",
		emptyText:     "No items",
		selectionMode: SelectSelectionSingle,
	}
}

func SelectMultiple(key string, selectedKeys []string, items []SelectItem) SelectWidget {
	return SelectWidget{
		key:           key,
		selectedKeys:  selectedKeys,
		items:         items,
		placeholder:   "Select items",
		emptyText:     "No items",
		selectionMode: SelectSelectionMultiple,
	}
}

func SelectSections(key, selectedKey string, sections []SelectSection) SelectWidget {
	return Select(key, selectedKey, nil).Sections(sections)
}

func SelectMultipleSections(key string, selectedKeys []string, sections []SelectSection) SelectWidget {
	return SelectMultiple(key, selectedKeys, nil).Sections(sections)
}

func (s SelectWidget) Placeholder(placeholder string) SelectWidget {
	s.placeholder = placeholder
	return s
}

// DataVersion enables item validation and flattened-data reuse. Increase
// version whenever the item data or section structure changes.
func (s SelectWidget) DataVersion(version uint64) SelectWidget {
	s.dataVersion = version
	s.hasDataVersion = true
	return s
}

func (s SelectWidget) EmptyText(text string) SelectWidget {
	s.emptyText = text
	return s
}

func (s SelectWidget) Label(label string) SelectWidget {
	s.label = label
	return s
}

func (s SelectWidget) Description(description string) SelectWidget {
	s.description = description
	return s
}

func (s SelectWidget) ErrorMessage(message string) SelectWidget {
	s.errorMessage = message
	return s
}

func (s SelectWidget) ValueText(text string) SelectWidget {
	s.valueText = text
	return s
}

func (s SelectWidget) Indicator(indicator frame.Widget) SelectWidget {
	s.indicator = indicator
	return s
}

func (s SelectWidget) Sections(sections []SelectSection) SelectWidget {
	s.sections = sections
	return s
}

func (s SelectWidget) OnChange(fn func(string)) SelectWidget {
	s.onChange = fn
	return s
}

func (s SelectWidget) OnSelectionChange(fn func([]string)) SelectWidget {
	s.onSelectionChange = fn
	return s
}

func (s SelectWidget) OnOpenChange(fn func(bool)) SelectWidget {
	s.onOpenChange = fn
	return s
}

func (s SelectWidget) Open(open bool) SelectWidget {
	s.open = open
	s.hasOpen = true
	return s
}

func (s SelectWidget) DefaultOpen(open bool) SelectWidget {
	s.defaultOpen = open
	s.hasDefaultOpen = true
	return s
}

func (s SelectWidget) Placement(placement overlay.PopoverPlacement) SelectWidget {
	s.placement = placement
	return s
}

func (s SelectWidget) ShouldFlip(shouldFlip bool) SelectWidget {
	s.shouldFlip = shouldFlip
	s.hasShouldFlip = true
	return s
}

func (s SelectWidget) AvoidOverflow(avoidOverflow bool) SelectWidget {
	s.avoidOverflow = avoidOverflow
	s.hasAvoidOverflow = true
	return s
}

func (s SelectWidget) DisabledKeys(keys []string) SelectWidget {
	s.disabledKeys = keys
	return s
}

func (s SelectWidget) Variant(variant SelectVariant) SelectWidget {
	s.variant = variant
	return s
}

func (s SelectWidget) Disabled(disabled bool) SelectWidget {
	s.disabled = disabled
	return s
}

func (s SelectWidget) Invalid(invalid bool) SelectWidget {
	s.invalid = invalid
	return s
}

func (s SelectWidget) Required(required bool) SelectWidget {
	s.required = required
	return s
}

func (s SelectWidget) FullWidth() SelectWidget {
	s.fullWidth = true
	return s
}

func (s SelectWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := selectStateFor(ctx, s.key)
	interactive := frame.OverlayInteractive(ctx, frame.OverlayLayerPopup, state.key)
	naturallyDisabled := frame.OverlayNaturallyDisabled(gtx)
	if s.label != "" {
		frame.PrepareFieldLabel(ctx, state.key, s.label)
	}
	if message, isError := s.supportMessage(); message != "" && !isError {
		frame.PrepareFieldDescription(ctx, state.key, message)
	}
	frame.RegisterFieldFocus(ctx, state.key, &state.trigger, gtx.Enabled() && !s.disabled)
	state.bind(s)
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
		frame.AfterOverlays(ctx, func() {
			if frame.OverlayTopmost(ctx, frame.OverlayLayerPopup, state.key) || !frame.HasTopOverlay(ctx) {
				frame.RequestFocus(ctx, &state.trigger)
			}
		})
	}
	eventGtx := gtx
	if s.disabled {
		eventGtx = eventGtx.Disabled()
	}

	progress := state.progress(gtx, open && !s.disabled, frame.ActiveTheme(ctx).Motion)
	dims := s.layout(ctx, eventGtx, state, open)
	if progress == 0 {
		return dims
	}
	s.layoutPopover(ctx, state, state.triggerRect, open, progress, naturallyDisabled)
	return dims
}

func (s SelectWidget) flipEnabled() bool {
	if !s.hasShouldFlip {
		return true
	}
	return s.shouldFlip
}

func (s SelectWidget) overflowAvoidanceEnabled() bool {
	if !s.hasAvoidOverflow {
		return true
	}
	return s.avoidOverflow
}

func (s SelectWidget) displayValue() (string, bool) {
	return s.displayValueItems(s.allItems())
}

func (s SelectWidget) displayValueCached(state *selectState) (string, bool) {
	return s.displayValueItems(state.itemsFor(s))
}

func (s SelectWidget) displayValueItems(items []SelectItem) (string, bool) {
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

func (s SelectWidget) selectedLabelFrom(items []SelectItem) (string, bool) {
	for _, item := range items {
		if item.Key == s.selectedKey {
			return item.Label, true
		}
	}
	return "", false
}

func (s SelectWidget) selectedLabelsFrom(items []SelectItem) []string {
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

func (s SelectWidget) allItems() []SelectItem {
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
