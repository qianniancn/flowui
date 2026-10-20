// Package themetest provides reusable visual-contract acceptance tests for
// built-in and third-party themes. It tests declarations and representative
// states without a window or graphics device; it does not compare screenshots.
package themetest

import (
	"fmt"
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Check validates every component supplied by active, then exercises common
// request variants. Call it inside a subtest for each supported theme mode.
// Validators check required renderers, geometry, paint, text and motion;
// provider and validator panics are reported as failures of that named case.
func Check(t *testing.T, active theme.Theme) {
	t.Helper()
	provider, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range componentCases(&active, provider) {
		run(t, test.name, test.check)
	}
	checkRequests(t, &active, provider)
}

type contractCase struct {
	name  string
	check func()
}

func componentCases(active *theme.Theme, provider theme.ComponentVisualProvider) []contractCase {
	return []contractCase{
		{"Button", func() { theme.ValidateButtonVisual(provider.Button(active, theme.ButtonRequest{})) }},
		{"Chip", func() { theme.ValidateChipVisual(provider.Chip(active, theme.ChipRequest{})) }},
		{"CloseButton", func() { theme.ValidateCloseButtonVisual(provider.CloseButton(active, theme.CloseButtonRequest{})) }},
		{"Badge", func() { theme.ValidateBadgeVisual(provider.Badge(active, theme.BadgeRequest{})) }},
		{"Avatar", func() { theme.ValidateAvatarVisual(provider.Avatar(active, theme.AvatarRequest{})) }},
		{"Description", func() { theme.ValidateDescriptionVisual(provider.Description(active, theme.DescriptionRequest{})) }},
		{"Label", func() { theme.ValidateLabelVisual(provider.Label(active, theme.LabelRequest{})) }},
		{"Text", func() { theme.ValidateTextVisual(provider.Text(active, theme.TextRequest{})) }},
		{"Icon", func() {
			theme.ValidateIconVisual(provider.Icon(active, theme.IconRequest{Foreground: active.Semantic().Foreground}))
		}},
		{"Surface", func() { theme.ValidateSurfaceVisual(provider.Surface(active, theme.SurfaceRequest{})) }},
		{"Card", func() { theme.ValidateCardVisual(provider.Card(active, theme.CardRequest{})) }},
		{"StatusBar", func() { theme.ValidateStatusBarVisual(provider.StatusBar(active, theme.StatusBarRequest{})) }},
		{"TitleBar", func() { theme.ValidateTitleBarVisual(provider.TitleBar(active, theme.TitleBarRequest{})) }},
		{"Toolbar", func() { theme.ValidateToolbarVisual(provider.Toolbar(active, theme.ToolbarRequest{})) }},
		{"Input", func() { theme.ValidateInputVisual(provider.Input(active, theme.InputRequest{})) }},
		{"DatePicker", func() { theme.ValidateDatePickerVisual(provider.DatePicker(active, theme.DatePickerRequest{})) }},
		{"InputAction", func() { theme.ValidateInputActionVisual(provider.InputAction(active, theme.InputActionRequest{})) }},
		{"Select", func() { theme.ValidateSelectVisual(provider.Select(active, theme.SelectRequest{})) }},
		{"Combo", func() { theme.ValidateComboVisual(provider.Combo(active, theme.ComboRequest{})) }},
		{"ListBox", func() { theme.ValidateListBoxVisual(provider.ListBox(active, theme.ListBoxRequest{})) }},
		{"Sidebar", func() { theme.ValidateSidebarVisual(provider.Sidebar(active, theme.SidebarRequest{})) }},
		{"Switch", func() { theme.ValidateSwitchVisual(provider.Switch(active, theme.SwitchRequest{})) }},
		{"SwitchGroup", func() { theme.ValidateSwitchGroupVisual(provider.SwitchGroup(active, theme.SwitchGroupRequest{})) }},
		{"ToggleButton", func() { theme.ValidateToggleVisual(provider.ToggleButton(active, theme.ToggleRequest{})) }},
		{"Checkbox", func() { theme.ValidateCheckboxVisual(provider.Checkbox(active, theme.CheckboxRequest{})) }},
		{"RadioGroup", func() { theme.ValidateRadioGroupVisual(provider.RadioGroup(active, theme.RadioGroupRequest{})) }},
		{"Slider", func() { theme.ValidateSliderVisual(provider.Slider(active, theme.SliderRequest{})) }},
		{"Menu", func() { theme.ValidateMenuVisual(provider.Menu(active, theme.MenuRequest{})) }},
		{"Menubar", func() { theme.ValidateMenubarVisual(provider.Menubar(active, theme.MenubarRequest{})) }},
		{"Dropdown", func() { theme.ValidateDropdownVisual(provider.Dropdown(active, theme.DropdownRequest{})) }},
		{"Popover", func() { theme.ValidatePopoverVisual(provider.Popover(active, theme.PopoverRequest{})) }},
		{"Tooltip", func() { theme.ValidateTooltipVisual(provider.Tooltip(active, theme.TooltipRequest{})) }},
		{"Toast", func() { theme.ValidateToastVisual(provider.Toast(active, theme.ToastRequest{})) }},
		{"Tabs", func() { theme.ValidateTabsVisual(provider.Tabs(active, theme.TabsRequest{})) }},
		{"Pagination", func() { theme.ValidatePaginationVisual(provider.Pagination(active, theme.PaginationRequest{})) }},
		{"Modal", func() { theme.ValidateModalVisual(provider.Modal(active, theme.ModalRequest{})) }},
		{"Alert", func() { theme.ValidateAlertVisual(provider.Alert(active, theme.AlertRequest{})) }},
		{"AlertDialog", func() { theme.ValidateAlertDialogVisual(provider.AlertDialog(active, theme.AlertDialogRequest{})) }},
		{"Table", func() { theme.ValidateTableVisual(provider.Table(active, theme.TableRequest{})) }},
		{"Heatmap", func() { theme.ValidateHeatmapVisual(provider.Heatmap(active, theme.HeatmapRequest{})) }},
		{"LineChart", func() { theme.ValidateLineChartVisual(provider.LineChart(active, theme.LineChartRequest{})) }},
		{"PieChart", func() { theme.ValidatePieChartVisual(provider.PieChart(active, theme.PieChartRequest{})) }},
		{"Candlestick", func() { theme.ValidateCandlestickVisual(provider.Candlestick(active, theme.CandlestickRequest{})) }},
		{"BarChart", func() { theme.ValidateBarChartVisual(provider.BarChart(active, theme.BarChartRequest{})) }},
		{"Gantt", func() { theme.ValidateGanttVisual(provider.Gantt(active, theme.GanttRequest{})) }},
		{"Timeline", func() { theme.ValidateTimelineVisual(provider.Timeline(active, theme.TimelineRequest{})) }},
		{"NodeGraph", func() { theme.ValidateNodeGraphVisual(provider.NodeGraph(active, theme.NodeGraphRequest{})) }},
		{"Tree", func() { theme.ValidateTreeVisual(provider.Tree(active, theme.TreeRequest{})) }},
		{"ColorPicker", func() { theme.ValidateColorPickerVisual(provider.ColorPicker(active, theme.ColorPickerRequest{})) }},
		{"ProgressBar", func() { theme.ValidateProgressBarVisual(provider.ProgressBar(active, theme.ProgressBarRequest{})) }},
		{"ProgressCircle", func() {
			theme.ValidateProgressCircleVisual(provider.ProgressCircle(active, theme.ProgressCircleRequest{}))
		}},
		{"Spinner", func() { theme.ValidateSpinnerVisual(provider.Spinner(active, theme.SpinnerRequest{})) }},
		{"Collapsible", func() { theme.ValidateCollapsibleVisual(provider.Collapsible(active, theme.CollapsibleRequest{})) }},
		{"Image", func() { theme.ValidateImageVisual(provider.Image(active, theme.ImageRequest{})) }},
		{"Code", func() { theme.ValidateCodeVisual(provider.Code(active, theme.CodeRequest{})) }},
		{"Skeleton", func() { theme.ValidateSkeletonVisual(provider.Skeleton(active, theme.SkeletonRequest{})) }},
		{"Segmented", func() { theme.ValidateSegmentedVisual(provider.Segmented(active, theme.SegmentedRequest{})) }},
	}
}

func run(t *testing.T, name string, check func()) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		defer func() {
			if value := recover(); value != nil {
				t.Fatalf("visual contract panic: %v", value)
			}
		}()
		check()
	})
}

func checkRequests(t *testing.T, active *theme.Theme, provider theme.ComponentVisualProvider) {
	t.Helper()
	for _, variant := range []theme.ButtonVariant{
		theme.ButtonPrimary, theme.ButtonSecondary, theme.ButtonTertiary,
		theme.ButtonGhost, theme.ButtonOutline, theme.ButtonDanger,
		theme.ButtonDangerSoft, theme.ButtonDangerOutline,
	} {
		for _, size := range []theme.ButtonSize{theme.ButtonSmall, theme.ButtonMedium, theme.ButtonLarge} {
			for _, iconOnly := range []bool{false, true} {
				run(t, fmt.Sprintf("Button/variant=%d/size=%d/icon=%t", variant, size, iconOnly), func() {
					theme.ValidateButtonVisual(provider.Button(active, theme.ButtonRequest{Variant: variant, Size: size, IconOnly: iconOnly}))
				})
			}
		}
	}

	run(t, "Button/grouped", func() {
		theme.ValidateButtonVisual(provider.Button(active, theme.ButtonRequest{Grouped: true, GroupCorners: theme.ButtonGroupRequest{NorthWest: true, SouthWest: true}}))
	})
	run(t, "Button/override", func() {
		visual := provider.Button(active, theme.ButtonRequest{Override: &theme.ButtonOverride{Declaration: style.Style{}.Height(47).Radius(0)}})
		theme.ValidateButtonVisual(visual)
		root := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if root.Box.Height == nil || *root.Box.Height != 47 || root.Paint.Radius == nil || *root.Paint.Radius != 0 {
			panic("button instance override must retain height 47 and radius 0")
		}
	})
	for _, variant := range []theme.InputVariant{theme.InputPrimary, theme.InputSecondary, theme.InputFilled, theme.InputBorderless, theme.InputUnderlined} {
		for _, size := range []theme.InputSize{theme.InputSmall, theme.InputMedium, theme.InputLarge} {
			for _, mode := range []theme.InputStyleMode{theme.InputModeSingleLine, theme.InputModeTextArea, theme.InputModeGroup} {
				run(t, fmt.Sprintf("Input/variant=%d/size=%d/mode=%d", variant, size, mode), func() {
					theme.ValidateInputVisual(provider.Input(active, theme.InputRequest{Variant: variant, Size: size, Mode: mode}))
				})
			}
		}
		run(t, fmt.Sprintf("Select/variant=%d", variant), func() {
			theme.ValidateSelectVisual(provider.Select(active, theme.SelectRequest{Variant: variant}))
		})
		run(t, fmt.Sprintf("Combo/variant=%d", variant), func() {
			theme.ValidateComboVisual(provider.Combo(active, theme.ComboRequest{Variant: variant}))
		})
		run(t, fmt.Sprintf("DatePicker/variant=%d", variant), func() {
			theme.ValidateDatePickerVisual(provider.DatePicker(active, theme.DatePickerRequest{Variant: variant}))
		})
	}
	for _, variant := range []theme.TreeVariant{theme.TreeVariantDefault, theme.TreeVariantSurface} {
		for _, size := range []theme.TreeSize{theme.TreeSizeSmall, theme.TreeSizeMedium} {
			run(t, fmt.Sprintf("Tree/variant=%d/size=%d", variant, size), func() {
				theme.ValidateTreeVisual(provider.Tree(active, theme.TreeRequest{Variant: variant, Size: size}))
			})
		}
	}
}
