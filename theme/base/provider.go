// Package base supplies shared component recipes for custom visual providers.
// Its zero-value Provider implements the complete ComponentVisualProvider.
// Override individual methods when a design system needs a different recipe.
package base

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/alertdialogstyle"
	"github.com/qianniancn/flowui/theme/alertstyle"
	"github.com/qianniancn/flowui/theme/avatarstyle"
	"github.com/qianniancn/flowui/theme/badgestyle"
	"github.com/qianniancn/flowui/theme/barchartstyle"
	"github.com/qianniancn/flowui/theme/buttonstyle"
	"github.com/qianniancn/flowui/theme/candlestickstyle"
	"github.com/qianniancn/flowui/theme/cardstyle"
	"github.com/qianniancn/flowui/theme/checkboxstyle"
	"github.com/qianniancn/flowui/theme/chipstyle"
	"github.com/qianniancn/flowui/theme/closebuttonstyle"
	"github.com/qianniancn/flowui/theme/codestyle"
	"github.com/qianniancn/flowui/theme/collapsiblestyle"
	"github.com/qianniancn/flowui/theme/colorpickerstyle"
	"github.com/qianniancn/flowui/theme/combostyle"
	"github.com/qianniancn/flowui/theme/datepickerstyle"
	"github.com/qianniancn/flowui/theme/descriptionstyle"
	"github.com/qianniancn/flowui/theme/dropdownstyle"
	"github.com/qianniancn/flowui/theme/ganttstyle"
	"github.com/qianniancn/flowui/theme/heatmapstyle"
	"github.com/qianniancn/flowui/theme/iconstyle"
	"github.com/qianniancn/flowui/theme/imagestyle"
	"github.com/qianniancn/flowui/theme/inputactionstyle"
	"github.com/qianniancn/flowui/theme/internal/fieldstyle"
	"github.com/qianniancn/flowui/theme/labelstyle"
	"github.com/qianniancn/flowui/theme/linechartstyle"
	"github.com/qianniancn/flowui/theme/listboxstyle"
	"github.com/qianniancn/flowui/theme/menubarstyle"
	"github.com/qianniancn/flowui/theme/menustyle"
	"github.com/qianniancn/flowui/theme/modalstyle"
	"github.com/qianniancn/flowui/theme/nodegraphstyle"
	"github.com/qianniancn/flowui/theme/paginationstyle"
	"github.com/qianniancn/flowui/theme/piechartstyle"
	"github.com/qianniancn/flowui/theme/popoverstyle"
	"github.com/qianniancn/flowui/theme/progressstyle"
	"github.com/qianniancn/flowui/theme/radiostyle"
	"github.com/qianniancn/flowui/theme/segmentedstyle"
	"github.com/qianniancn/flowui/theme/selectstyle"
	"github.com/qianniancn/flowui/theme/sidebarstyle"
	"github.com/qianniancn/flowui/theme/skeletonstyle"
	"github.com/qianniancn/flowui/theme/sliderstyle"
	"github.com/qianniancn/flowui/theme/spinnerstyle"
	"github.com/qianniancn/flowui/theme/statusbarstyle"
	"github.com/qianniancn/flowui/theme/surfacestyle"
	"github.com/qianniancn/flowui/theme/switchstyle"
	"github.com/qianniancn/flowui/theme/tablestyle"
	"github.com/qianniancn/flowui/theme/tabsstyle"
	"github.com/qianniancn/flowui/theme/textstyle"
	"github.com/qianniancn/flowui/theme/timelinestyle"
	"github.com/qianniancn/flowui/theme/titlebarstyle"
	"github.com/qianniancn/flowui/theme/toaststyle"
	"github.com/qianniancn/flowui/theme/togglestyle"
	"github.com/qianniancn/flowui/theme/toolbarstyle"
	"github.com/qianniancn/flowui/theme/tooltipstyle"
	"github.com/qianniancn/flowui/theme/treestyle"
)

// Provider delegates common components to the shared theme recipes. Its zero
// value is ready to embed; methods always read the active theme passed to them.
// Provider deliberately does not implement theme.ImmutableVisualProvider:
// a custom provider must explicitly promise that its overrides are immutable.
type Provider struct{}

var _ theme.ComponentVisualProvider = Provider{}

func (Provider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	config := buttonstyle.DefaultConfig()
	config.Radius = 4
	if request.Size == theme.ButtonSmall {
		config.Radius = 3
	}
	if request.Size == theme.ButtonLarge {
		config.Radius = 6
	}
	config.Height, config.SmallHeight, config.LargeHeight = 32, 24, 40
	config.PaddingX, config.SmallPaddingX, config.LargePaddingX = 12, 8, 16
	config.PressedScaleSmall, config.PressedScaleMedium, config.PressedScaleLarge = 1, 1, 1
	visual := buttonstyle.BuildButton(active, request, config)
	if request.Size == theme.ButtonSmall {
		visual.Size = style.Join(visual.Size, style.New().FontSize(12))
	}
	sem := active.Semantic()
	pair := sem.Tone(theme.TonePrimary)
	if request.Variant == theme.ButtonDangerOutline {
		pair = sem.Tone(theme.ToneDanger)
	}
	switch request.Variant {
	case theme.ButtonSecondary:
		visual.Variant = softButton(sem, pair)
	case theme.ButtonOutline, theme.ButtonDangerOutline:
		edge := pair.Fill
		dark := luminance(sem.Background) < luminance(sem.Foreground)
		if scale, ok := accentScaleFor(pair.Fill, dark); ok {
			edge = rgb(scale.edge)
		}
		visual.Variant = style.New().Background(style.RGBA(0)).
			TextColor(style.Solid(pair.OnSoft)).BorderWidth(1).BorderColor(style.Solid(edge)).
			On(style.Hovered, style.New().Background(style.Solid(pair.Soft))).
			On(style.Pressed, style.New().Background(style.Solid(pair.SoftHover))).
			On(style.Disabled, disabledButton(sem))
	}
	visual.Defaults = style.Join(visual.Defaults, style.New().
		OnNot(style.FocusVisible, style.Disabled, style.New().Outline(2, 2, style.Solid(sem.Ring))))
	visual.LoadingInset = func(_ layout.Context, request theme.ButtonLoadingRequest) layout.Inset {
		return request.Inset
	}
	return visual
}

func softButton(sem theme.Semantic, pair theme.TonePair) style.Style {
	// Pressing steps one rung past the hovered tint; sharing the hover colour
	// made the two states indistinguishable.
	pressed := theme.Mix(pair.SoftHover, pair.OnSoft, 8)
	return style.New().Background(style.Solid(pair.Soft)).TextColor(style.Solid(pair.OnSoft)).
		On(style.Hovered, style.New().Background(style.Solid(pair.SoftHover))).
		On(style.Pressed, style.New().Background(style.Solid(pressed))).
		On(style.Disabled, disabledButton(sem))
}

func disabledButton(sem theme.Semantic) style.Style {
	return style.New().Background(style.Solid(sem.DisabledBackground)).
		TextColor(style.Solid(sem.DisabledForeground)).BorderColor(style.Solid(sem.DisabledBorder)).
		BoxShadowNone().Opacity(1)
}

func (Provider) Chip(active *theme.Theme, request theme.ChipRequest) theme.ChipVisual {
	return chipstyle.BuildChip(active, request)
}

func (Provider) CloseButton(active *theme.Theme, request theme.CloseButtonRequest) theme.CloseButtonVisual {
	return closebuttonstyle.BuildCloseButton(active, request)
}

func (Provider) Badge(active *theme.Theme, request theme.BadgeRequest) theme.BadgeVisual {
	return badgestyle.BuildBadge(active, request)
}

func (Provider) Avatar(active *theme.Theme, request theme.AvatarRequest) theme.AvatarVisual {
	return avatarstyle.BuildAvatar(active, request)
}

func (Provider) Description(active *theme.Theme, request theme.DescriptionRequest) theme.DescriptionVisual {
	return descriptionstyle.BuildDescription(active, request)
}

func (Provider) Label(active *theme.Theme, request theme.LabelRequest) theme.LabelVisual {
	return labelstyle.BuildLabel(active, request)
}

func (Provider) Text(active *theme.Theme, request theme.TextRequest) theme.TextVisual {
	return textstyle.BuildText(active, request)
}

func (Provider) Icon(active *theme.Theme, request theme.IconRequest) theme.IconVisual {
	return iconstyle.BuildIcon(active, request)
}

func (Provider) Surface(active *theme.Theme, request theme.SurfaceRequest) theme.SurfaceVisual {
	return surfacestyle.BuildSurface(active, request)
}

func (Provider) Card(active *theme.Theme, request theme.CardRequest) theme.CardVisual {
	visual := cardstyle.BuildCard(active, request)
	visual.Defaults = style.Join(visual.Defaults, style.New().BorderColor(style.Solid(active.Semantic().Separator)))
	return visual
}

func (Provider) StatusBar(active *theme.Theme, request theme.StatusBarRequest) theme.StatusBarVisual {
	return statusbarstyle.BuildStatusBar(active, request, statusbarstyle.DefaultConfig())
}

func (Provider) TitleBar(active *theme.Theme, request theme.TitleBarRequest) theme.TitleBarVisual {
	return titlebarstyle.BuildTitleBar(active, request)
}

func (Provider) Toolbar(active *theme.Theme, request theme.ToolbarRequest) theme.ToolbarVisual {
	return toolbarstyle.BuildToolbar(active, request)
}

func (Provider) Input(active *theme.Theme, request theme.InputRequest) theme.InputVisual {
	return fieldstyle.BuildInput(active, request)
}

func (Provider) DatePicker(active *theme.Theme, request theme.DatePickerRequest) theme.DatePickerVisual {
	return datepickerstyle.BuildDatePicker(active, request)
}

func (Provider) InputAction(active *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	return inputactionstyle.BuildInputAction(active, request)
}

func (Provider) Select(active *theme.Theme, request theme.SelectRequest) theme.SelectVisual {
	return selectstyle.BuildSelect(active, request)
}

func (Provider) Combo(active *theme.Theme, request theme.ComboRequest) theme.ComboVisual {
	return combostyle.BuildCombo(active, request)
}

func (Provider) ListBox(active *theme.Theme, request theme.ListBoxRequest) theme.ListBoxVisual {
	visual := listboxstyle.BuildListBox(active, request)
	item := visual.Item
	sem := active.Semantic()
	visual.Item = func(request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		value := item(request)
		if request.State.Selected {
			background := sem.Tone(theme.TonePrimary).Soft
			if request.State.Hovered {
				background = sem.Tone(theme.TonePrimary).SoftHover
			}
			value.Variant = style.Join(value.Variant, style.New().Part(style.PartItem,
				style.New().Background(style.Solid(background))))
		}
		return value
	}
	return visual
}

func (Provider) Sidebar(active *theme.Theme, request theme.SidebarRequest) theme.SidebarVisual {
	return sidebarstyle.BuildSidebar(active, request)
}

func (Provider) Switch(active *theme.Theme, request theme.SwitchRequest) theme.SwitchVisual {
	visual := switchstyle.BuildSwitch(active, request)
	visual.SelectionDuration = active.Motion.ControlDuration
	white := style.Solid(rgb(0xffffff))
	visual.Defaults = style.Join(visual.Defaults, style.New().
		Part(style.PartTrack, style.New().BorderWidth(1).BorderColor(style.Solid(active.Semantic().Separator)).
			On(style.Checked, style.New().BorderColor(style.RGBA(0)))).
		Part(style.PartThumb, style.New().Background(white).
			BoxShadow(0, 1, 2, 0, style.WithAlpha(rgb(0), .16)).
			On(style.Checked, style.New().Background(white))))
	return visual
}

func (Provider) SwitchGroup(active *theme.Theme, request theme.SwitchGroupRequest) theme.SwitchGroupVisual {
	return switchstyle.BuildSwitchGroup(active, request)
}

func (Provider) ToggleButton(active *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	return togglestyle.BuildToggleButton(active, request)
}

func (Provider) Checkbox(active *theme.Theme, request theme.CheckboxRequest) theme.CheckboxVisual {
	return checkboxstyle.BuildCheckbox(active, request)
}

func (Provider) RadioGroup(active *theme.Theme, request theme.RadioGroupRequest) theme.RadioGroupVisual {
	return radiostyle.BuildRadioGroup(active, request)
}

func (Provider) Slider(active *theme.Theme, request theme.SliderRequest) theme.SliderVisual {
	visual := sliderstyle.BuildSlider(active, request)
	margin := max((active.Components.Slider.ThumbSize-active.Components.Slider.TrackThickness)/2, 0)
	track := style.New().MarginY(margin)
	if request.Orientation == theme.SliderVertical {
		track = style.New().MarginX(margin)
	}
	sem := active.Semantic()
	thumb := style.New().Background(style.Solid(rgb(0xffffff))).BorderColor(style.Solid(sem.Border)).
		BoxShadowNone().BoxShadow(0, 1, 3, 0, style.WithAlpha(rgb(0), .16)).
		On(style.Disabled, style.New().
			Background(style.Solid(active.DisabledColor(rgb(0xffffff)))).
			BorderColor(style.Solid(active.DisabledColor(sem.Border))).BoxShadowNone())
	visual.Defaults = style.Join(visual.Defaults, style.New().
		Part(style.PartTrack, track).Part(style.PartThumb, thumb))
	return visual
}

func (Provider) Menu(active *theme.Theme, request theme.MenuRequest) theme.MenuVisual {
	return menustyle.BuildMenu(active, request)
}

func (Provider) Menubar(active *theme.Theme, request theme.MenubarRequest) theme.MenubarVisual {
	return menubarstyle.BuildMenubar(active, request)
}

func (Provider) Dropdown(active *theme.Theme, request theme.DropdownRequest) theme.DropdownVisual {
	return dropdownstyle.BuildDropdown(active, request)
}

func (Provider) Popover(active *theme.Theme, request theme.PopoverRequest) theme.PopoverVisual {
	return popoverstyle.BuildPopover(active, request)
}

func (Provider) Tooltip(active *theme.Theme, request theme.TooltipRequest) theme.TooltipVisual {
	return tooltipstyle.BuildTooltip(active, request)
}

func (Provider) Toast(active *theme.Theme, request theme.ToastRequest) theme.ToastVisual {
	return toaststyle.BuildToast(active, request)
}

func (Provider) Tabs(active *theme.Theme, request theme.TabsRequest) theme.TabsVisual {
	visual := tabsstyle.BuildTabs(active, request)
	sem := active.Semantic()
	visual.Colors.ListBorder = sem.Separator
	if request.Variant == 1 {
		visual.Metrics.SmallTabHeight, visual.Metrics.TabHeight, visual.Metrics.LargeTabHeight = 32, 40, 48
	} else {
		visual.Colors.ListBackground = sem.SurfaceHover
	}
	return visual
}

func (Provider) Pagination(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	return paginationstyle.BuildPagination(active, request)
}

func (Provider) Modal(active *theme.Theme, request theme.ModalRequest) theme.ModalVisual {
	return modalstyle.BuildModal(active, request)
}

func (Provider) Alert(active *theme.Theme, request theme.AlertRequest) theme.AlertVisual {
	return alertstyle.BuildAlert(active, request)
}

func (Provider) AlertDialog(active *theme.Theme, request theme.AlertDialogRequest) theme.AlertDialogVisual {
	return alertdialogstyle.BuildAlertDialog(active, request)
}

func (Provider) Table(active *theme.Theme, request theme.TableRequest) theme.TableVisual {
	visual := tablestyle.BuildTable(active, request)
	row := visual.Row
	sem := active.Semantic()
	visual.Colors.Root, visual.Colors.Body = sem.Surface, sem.Surface
	visual.Colors.Header = theme.Mix(sem.Surface, sem.SurfaceHover, 50)
	visual.Colors.Border = sem.Separator
	visual.Colors.RowSeparator, visual.Colors.HeaderSeparator = sem.Separator, sem.Separator
	visual.Colors.ColumnSeparator = sem.Separator
	visual.Row = func(request theme.TableRowRequest) theme.TableRowVisual {
		value := row(request)
		if request.Selected {
			value.Background = sem.Tone(theme.TonePrimary).Soft
			if request.Hovered || request.Pressed {
				value.Background = sem.Tone(theme.TonePrimary).SoftHover
			}
		} else if request.Hovered || request.Pressed {
			value.Background = sem.SurfaceHover
		}
		return value
	}
	return visual
}

func (Provider) Heatmap(active *theme.Theme, request theme.HeatmapRequest) theme.HeatmapVisual {
	return heatmapstyle.BuildHeatmap(active, request)
}

func (Provider) LineChart(active *theme.Theme, request theme.LineChartRequest) theme.LineChartVisual {
	return linechartstyle.BuildLineChart(active, request)
}

func (Provider) PieChart(active *theme.Theme, request theme.PieChartRequest) theme.PieChartVisual {
	return piechartstyle.BuildPieChart(active, request)
}

func (Provider) Candlestick(active *theme.Theme, request theme.CandlestickRequest) theme.CandlestickVisual {
	return candlestickstyle.BuildCandlestick(active, request)
}

func (Provider) BarChart(active *theme.Theme, request theme.BarChartRequest) theme.BarChartVisual {
	return barchartstyle.BuildBarChart(active, request)
}

func (Provider) Gantt(active *theme.Theme, request theme.GanttRequest) theme.GanttVisual {
	return ganttstyle.BuildGantt(active, request)
}

func (Provider) Timeline(active *theme.Theme, request theme.TimelineRequest) theme.TimelineVisual {
	return timelinestyle.BuildTimeline(active, request)
}

func (Provider) NodeGraph(active *theme.Theme, request theme.NodeGraphRequest) theme.NodeGraphVisual {
	return nodegraphstyle.BuildNodeGraph(active, request)
}

func (Provider) Tree(active *theme.Theme, request theme.TreeRequest) theme.TreeVisual {
	return treestyle.BuildTree(active, request)
}

func (Provider) ColorPicker(active *theme.Theme, request theme.ColorPickerRequest) theme.ColorPickerVisual {
	return colorpickerstyle.BuildColorPicker(active, request)
}

func (Provider) ProgressBar(active *theme.Theme, request theme.ProgressBarRequest) theme.ProgressBarVisual {
	return progressstyle.BuildProgressBar(active, request)
}

func (Provider) ProgressCircle(active *theme.Theme, request theme.ProgressCircleRequest) theme.ProgressCircleVisual {
	return progressstyle.BuildProgressCircle(active, request)
}

func (Provider) Spinner(active *theme.Theme, request theme.SpinnerRequest) theme.SpinnerVisual {
	return spinnerstyle.BuildSpinner(active, request)
}

func (Provider) Collapsible(active *theme.Theme, request theme.CollapsibleRequest) theme.CollapsibleVisual {
	return collapsiblestyle.BuildCollapsible(active, request)
}

// Image returns the image shape recipe for the active theme.
func (Provider) Image(active *theme.Theme, request theme.ImageRequest) theme.ImageVisual {
	return imagestyle.BuildImage(active, request)
}

// Code returns the code-block surface recipe for the active theme.
func (Provider) Code(active *theme.Theme, request theme.CodeRequest) theme.CodeVisual {
	return codestyle.BuildCode(active, request)
}

// Skeleton returns the placeholder recipe for the active theme.
func (Provider) Skeleton(active *theme.Theme, request theme.SkeletonRequest) theme.SkeletonVisual {
	return skeletonstyle.BuildSkeleton(active, request)
}

// Segmented returns the segmented-track recipe for the active theme.
func (Provider) Segmented(active *theme.Theme, request theme.SegmentedRequest) theme.SegmentedVisual {
	return segmentedstyle.BuildSegmented(active, request)
}
