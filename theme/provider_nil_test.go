package theme_test

import (
	"errors"
	"github.com/qianniancn/flowui/theme/codestyle"
	"github.com/qianniancn/flowui/theme/imagestyle"
	"github.com/qianniancn/flowui/theme/inputactionstyle"
	"github.com/qianniancn/flowui/theme/segmentedstyle"
	"github.com/qianniancn/flowui/theme/skeletonstyle"
	"testing"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/badgestyle"
	"github.com/qianniancn/flowui/theme/buttonstyle"
	"github.com/qianniancn/flowui/theme/cardstyle"
	"github.com/qianniancn/flowui/theme/checkboxstyle"
	"github.com/qianniancn/flowui/theme/combostyle"
	"github.com/qianniancn/flowui/theme/datepickerstyle"
	"github.com/qianniancn/flowui/theme/descriptionstyle"
	"github.com/qianniancn/flowui/theme/dropdownstyle"
	"github.com/qianniancn/flowui/theme/internal/fieldstyle"
	"github.com/qianniancn/flowui/theme/labelstyle"
	"github.com/qianniancn/flowui/theme/listboxstyle"
	"github.com/qianniancn/flowui/theme/menustyle"
	"github.com/qianniancn/flowui/theme/modalstyle"
	"github.com/qianniancn/flowui/theme/popoverstyle"
	"github.com/qianniancn/flowui/theme/radiostyle"
	"github.com/qianniancn/flowui/theme/selectstyle"
	"github.com/qianniancn/flowui/theme/sidebarstyle"
	"github.com/qianniancn/flowui/theme/sliderstyle"
	"github.com/qianniancn/flowui/theme/statusbarstyle"
	"github.com/qianniancn/flowui/theme/surfacestyle"
	"github.com/qianniancn/flowui/theme/switchstyle"
	"github.com/qianniancn/flowui/theme/tablestyle"
	"github.com/qianniancn/flowui/theme/togglestyle"
	"github.com/qianniancn/flowui/theme/tooltipstyle"
)

func TestSharedVisualRecipesRejectMissingTheme(t *testing.T) {
	tests := map[string]func(){
		"button":       func() { buttonstyle.BuildButton(nil, theme.ButtonRequest{}, buttonstyle.Config{}) },
		"badge":        func() { badgestyle.BuildBadge(nil, theme.BadgeRequest{}) },
		"checkbox":     func() { checkboxstyle.BuildCheckbox(nil, theme.CheckboxRequest{}) },
		"combo":        func() { combostyle.BuildCombo(nil, theme.ComboRequest{}) },
		"menu":         func() { menustyle.BuildMenu(nil, theme.MenuRequest{}) },
		"dropdown":     func() { dropdownstyle.BuildDropdown(nil, theme.DropdownRequest{}) },
		"popover":      func() { popoverstyle.BuildPopover(nil, theme.PopoverRequest{}) },
		"tooltip":      func() { tooltipstyle.BuildTooltip(nil, theme.TooltipRequest{}) },
		"modal":        func() { modalstyle.BuildModal(nil, theme.ModalRequest{}) },
		"listbox":      func() { listboxstyle.BuildListBox(nil, theme.ListBoxRequest{}) },
		"sidebar":      func() { sidebarstyle.BuildSidebar(nil, theme.SidebarRequest{}) },
		"table":        func() { tablestyle.BuildTable(nil, theme.TableRequest{}) },
		"datepicker":   func() { datepickerstyle.BuildDatePicker(nil, theme.DatePickerRequest{}) },
		"description":  func() { descriptionstyle.BuildDescription(nil, theme.DescriptionRequest{}) },
		"input":        func() { fieldstyle.BuildInput(nil, theme.InputRequest{}) },
		"input-action": func() { inputactionstyle.BuildInputAction(nil, theme.InputActionRequest{}) },
		"label":        func() { labelstyle.BuildLabel(nil, theme.LabelRequest{}) },
		"surface":      func() { surfacestyle.BuildSurface(nil, theme.SurfaceRequest{}) },
		"card":         func() { cardstyle.BuildCard(nil, theme.CardRequest{}) },
		"statusbar":    func() { statusbarstyle.BuildStatusBar(nil, theme.StatusBarRequest{}, statusbarstyle.DefaultConfig()) },
		"radio":        func() { radiostyle.BuildRadioGroup(nil, theme.RadioGroupRequest{}) },
		"select":       func() { selectstyle.BuildSelect(nil, theme.SelectRequest{}) },
		"slider":       func() { sliderstyle.BuildSlider(nil, theme.SliderRequest{}) },
		"switch":       func() { switchstyle.BuildSwitch(nil, theme.SwitchRequest{}) },
		"switch-group": func() { switchstyle.BuildSwitchGroup(nil, theme.SwitchGroupRequest{}) },
		"togglebutton": func() { togglestyle.BuildToggleButton(nil, theme.ToggleRequest{}) },
		"code":         func() { codestyle.BuildCode(nil, theme.CodeRequest{}) },
		"image":        func() { imagestyle.BuildImage(nil, theme.ImageRequest{}) },
		"skeleton":     func() { skeletonstyle.BuildSkeleton(nil, theme.SkeletonRequest{}) },
		"segmented":    func() { segmentedstyle.BuildSegmented(nil, theme.SegmentedRequest{}) },
	}
	for name, build := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				err, ok := recovered.(error)
				if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
					t.Fatalf("panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
				}
			}()
			build()
		})
	}
}
