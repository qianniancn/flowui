package theme

import (
	"image/color"
	"math"
	"time"

	"gioui.org/font"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Theme contains FlowUI design tokens.
//
// The Gio material theme is an unexported substrate bridge for text/editor
// helpers. Applications theme through Palette, Typography, Shape, Spacing,
// Motion, Components, and Fonts — never through material.Theme.
type Theme struct {
	Palette         Palette
	Typography      Typography
	Fonts           FontConfig
	Shape           Shape
	Spacing         Spacing
	Shadows         ShadowsTheme
	Motion          MotionTheme
	Components      ComponentsTheme
	DisabledOpacity float32
	material        *material.Theme
	managedShaper   *text.Shaper
	managedFace     font.Typeface
}

type MotionTheme struct {
	// Enabled controls all framework animations.
	Enabled bool
	// DefaultDuration is used by Tween when no duration is specified.
	DefaultDuration time.Duration
	// DurationScale scales animation time globally. Zero disables motion.
	DurationScale float32
}

// ResolveMotionDuration applies the active motion policy to a transition.
// A non-positive result means the transition should snap to its target.
func ResolveMotionDuration(motion MotionTheme, duration time.Duration) time.Duration {
	if !MotionEnabled(motion) || duration <= 0 {
		return 0
	}
	scaled := time.Duration(float64(duration) * float64(motion.DurationScale))
	if scaled <= 0 {
		return 0
	}
	return scaled
}

// MotionEnabled reports whether the global motion policy permits animation.
func MotionEnabled(motion MotionTheme) bool {
	return motion.Enabled && motion.DurationScale > 0 &&
		!math.IsNaN(float64(motion.DurationScale)) && !math.IsInf(float64(motion.DurationScale), 0)
}

type Palette struct {
	Background                 color.NRGBA
	Surface                    color.NRGBA
	SurfaceForeground          color.NRGBA
	SurfaceSecondary           color.NRGBA
	SurfaceSecondaryForeground color.NRGBA
	SurfaceTertiary            color.NRGBA
	SurfaceTertiaryForeground  color.NRGBA
	SurfaceHover               color.NRGBA
	SurfacePressed             color.NRGBA
	SurfaceRaised              color.NRGBA
	Overlay                    color.NRGBA
	OverlayForeground          color.NRGBA
	Foreground                 color.NRGBA
	MutedForeground            color.NRGBA
	Border                     color.NRGBA
	Separator                  color.NRGBA
	Default                    color.NRGBA
	DefaultForeground          color.NRGBA
	DefaultHover               color.NRGBA
	FieldBackground            color.NRGBA
	FieldHover                 color.NRGBA
	FieldForeground            color.NRGBA
	FieldPlaceholder           color.NRGBA
	FieldFocus                 color.NRGBA
	Segment                    color.NRGBA
	SegmentForeground          color.NRGBA
	Accent                     color.NRGBA
	AccentHover                color.NRGBA
	AccentPressed              color.NRGBA
	AccentForeground           color.NRGBA
	AccentSoft                 color.NRGBA
	AccentSoftHover            color.NRGBA
	AccentSoftForeground       color.NRGBA
	Success                    color.NRGBA
	SuccessForeground          color.NRGBA
	SuccessSoft                color.NRGBA
	SuccessSoftForeground      color.NRGBA
	Warning                    color.NRGBA
	WarningForeground          color.NRGBA
	WarningSoft                color.NRGBA
	WarningSoftForeground      color.NRGBA
	Danger                     color.NRGBA
	DangerHover                color.NRGBA
	DangerPressed              color.NRGBA
	DangerForeground           color.NRGBA
	DangerSoft                 color.NRGBA
	DangerSoftHover            color.NRGBA
	DangerSoftForeground       color.NRGBA
	Focus                      color.NRGBA
	Selection                  color.NRGBA
	SurfaceShadow              color.NRGBA
	OverlayShadow              color.NRGBA
}

type Typography struct {
	// Typeface is the default family list used by framework text.
	Typeface font.Typeface
	// MonoTypeface is the recommended family list for code and tabular text.
	MonoTypeface font.Typeface
	BodySize     unit.Sp
	ControlSize  unit.Sp
	SmallSize    unit.Sp
}

// FontConfig controls the font faces available to each window's text shaper.
// Collection can contain faces parsed from TTF, OTF, or TTC data. SystemFonts
// keeps the platform font fallback enabled when true.
type FontConfig struct {
	Collection  []font.FontFace
	SystemFonts bool
}

type Shape struct {
	ControlRadius  unit.Dp
	PopoverRadius  unit.Dp
	ItemRadius     unit.Dp
	CheckboxRadius unit.Dp
}

type Spacing struct {
	ControlHeight        unit.Dp
	SmallControlHeight   unit.Dp
	LargeControlHeight   unit.Dp
	ControlPaddingX      unit.Dp
	SmallControlPaddingX unit.Dp
	LargeControlPaddingX unit.Dp
	IconButtonSize       unit.Dp
	PanelGap             unit.Dp
	PanelPadding         unit.Dp
	ItemHeight           unit.Dp
}

const ShadowLayerCount = 3

// ShadowLayerTheme describes one theme-controlled soft shadow pass.
type ShadowLayerTheme struct {
	OffsetX unit.Dp
	OffsetY unit.Dp
	// Blur is capped at 96 physical pixels during rasterization.
	Blur   unit.Dp
	Spread unit.Dp
	// Opacity is clamped to [0, 1]. Zero disables the layer.
	Opacity float32
}

// ShadowTheme stores shadow layers from tightest to broadest.
type ShadowTheme struct {
	Layers [ShadowLayerCount]ShadowLayerTheme
}

// ShadowsTheme groups the shadow profiles used by framework surfaces.
type ShadowsTheme struct {
	Surface     ShadowTheme
	Overlay     ShadowTheme
	Menu        ShadowTheme
	Control     ShadowTheme
	Checkbox    ShadowTheme
	SwitchThumb ShadowTheme
}

type ComponentsTheme struct {
	Button            ButtonTheme
	ButtonGroup       ButtonGroupTheme
	ToggleButton      ToggleButtonTheme
	CloseButton       CloseButtonTheme
	Chip              ChipTheme
	Avatar            AvatarTheme
	Badge             BadgeTheme
	Surface           SurfaceTheme
	Card              CardTheme
	Alert             AlertTheme
	AlertDialog       AlertDialogTheme
	Description       DescriptionTheme
	Label             LabelTheme
	Input             InputTheme
	TextArea          TextAreaTheme
	InputGroup        InputGroupTheme
	Checkbox          CheckboxTheme
	Switch            SwitchTheme
	SwitchGroup       SwitchGroupTheme
	RadioGroup        RadioGroupTheme
	ProgressBar       ProgressBarTheme
	ProgressCircle    ProgressCircleTheme
	Spinner           SpinnerTheme
	Slider            SliderTheme
	ListBox           ListBoxTheme
	Tree              TreeTheme
	Sidebar           SidebarTheme
	Scrollbar         ScrollbarTheme
	SplitPane         SplitPaneTheme
	TitleBar          TitleBarTheme
	Toolbar           ToolbarTheme
	Table             TableTheme
	Pagination        PaginationTheme
	Menu              MenuTheme
	Dropdown          DropdownTheme
	Menubar           MenubarTheme
	LineChart         LineChartTheme
	BarChart          BarChartTheme
	PieChart          PieChartTheme
	CandlestickChart  CandlestickChartTheme
	Heatmap           HeatmapTheme
	GanttChart        GanttChartTheme
	NodeGraph         NodeGraphTheme
	Tabs              TabsTheme
	Workbench         WorkbenchTheme
	Collapsible       CollapsibleTheme
	Select            SelectTheme
	Popover           PopoverTheme
	Tooltip           TooltipTheme
	Toast             ToastTheme
	Modal             ModalTheme
	ComboBox          ComboBoxTheme
	DatePicker        DatePickerTheme
	ColorPicker       ColorPickerTheme
	ColorWheel        ColorWheelTheme
	ColorArea         ColorAreaTheme
	ColorField        ColorFieldTheme
	ColorSlider       ColorSliderTheme
	ColorSwatch       ColorSwatchTheme
	ColorSwatchPicker ColorSwatchPickerTheme
}

type SurfaceTheme struct {
	BorderWidth unit.Dp
}

type CardTheme struct {
	Padding unit.Dp
	Gap     unit.Dp
	Radius  unit.Dp
}

type AlertTheme struct {
	PaddingX              unit.Dp
	PaddingY              unit.Dp
	Gap                   unit.Dp
	Radius                unit.Dp
	IndicatorPadding      unit.Dp
	IconSize              unit.Dp
	TitleSize             unit.Sp
	TitleLineHeight       unit.Sp
	DescriptionSize       unit.Sp
	DescriptionLineHeight unit.Sp
}

type AlertDialogTheme struct {
	IconSize      unit.Dp
	IconGlyphSize unit.Dp
	HeaderGap     unit.Dp
	TitleSize     unit.Sp
}

type DescriptionTheme struct {
	TextSize unit.Sp
}

type LabelTheme struct {
	TextSize           unit.Sp
	RequiredMarkOffset unit.Dp
}

type ButtonTheme struct {
	Radius             unit.Dp
	BorderWidth        unit.Dp
	ContentGap         unit.Dp
	FocusRingWidth     unit.Dp
	SpinnerStrokeWidth unit.Dp
	SpinnerSmall       unit.Dp
	SpinnerMedium      unit.Dp
	SpinnerLarge       unit.Dp
	PressedScaleSmall  float32
	PressedScaleMedium float32
	PressedScaleLarge  float32
}

type ButtonGroupTheme struct {
	SeparatorWidth   unit.Dp
	SeparatorLength  float32
	SeparatorOpacity float32
}

type ToggleButtonTheme struct {
	SmallHeight        unit.Dp
	MediumHeight       unit.Dp
	LargeHeight        unit.Dp
	SmallPaddingX      unit.Dp
	MediumPaddingX     unit.Dp
	LargePaddingX      unit.Dp
	Radius             unit.Dp
	ContentGap         unit.Dp
	SmallTextSize      unit.Sp
	MediumTextSize     unit.Sp
	LargeTextSize      unit.Sp
	FocusRingWidth     unit.Dp
	FocusRingOffset    unit.Dp
	PressedScaleSmall  float32
	PressedScaleMedium float32
	PressedScaleLarge  float32
}

type CloseButtonTheme struct {
	Size           unit.Dp
	Radius         unit.Dp
	Padding        unit.Dp
	IconSize       unit.Dp
	FocusRingWidth unit.Dp
	PressedScale   float32
}

type ChipTheme struct {
	SmallHeight    unit.Dp
	MediumHeight   unit.Dp
	LargeHeight    unit.Dp
	SmallPaddingX  unit.Dp
	MediumPaddingX unit.Dp
	LargePaddingX  unit.Dp
	SmallPaddingY  unit.Dp
	MediumPaddingY unit.Dp
	LargePaddingY  unit.Dp
	LabelPaddingX  unit.Dp
	ContentGap     unit.Dp
	Radius         unit.Dp
	SmallTextSize  unit.Sp
	MediumTextSize unit.Sp
	LargeTextSize  unit.Sp
	LineHeight     unit.Sp
}

type AvatarTheme struct {
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	SmallRadius    unit.Dp
	MediumRadius   unit.Dp
	LargeRadius    unit.Dp
	SmallTextSize  unit.Sp
	MediumTextSize unit.Sp
	LargeTextSize  unit.Sp
	SmallIconSize  unit.Dp
	MediumIconSize unit.Dp
	LargeIconSize  unit.Dp
}

type BadgeTheme struct {
	SmallMinSize         unit.Dp
	MediumMinSize        unit.Dp
	LargeMinSize         unit.Dp
	SmallRadius          unit.Dp
	MediumRadius         unit.Dp
	LargeRadius          unit.Dp
	SmallTextSize        unit.Sp
	MediumTextSize       unit.Sp
	LargeTextSize        unit.Sp
	SmallLineHeight      unit.Sp
	MediumLineHeight     unit.Sp
	LargeLineHeight      unit.Sp
	LabelPaddingX        unit.Dp
	BorderWidth          unit.Dp
	PlacementOffsetRatio float32
}

type InputTheme struct {
	Height              unit.Dp
	Radius              unit.Dp
	PaddingX            unit.Dp
	TextSize            unit.Sp
	LineHeight          unit.Sp
	FocusRingWidth      unit.Dp
	InvalidOutlineWidth unit.Dp
	ShadowColor         color.NRGBA
	ShadowOpacity       float32
	ShadowStrength      float32
}

type TextAreaTheme struct {
	MinHeight           unit.Dp
	Radius              unit.Dp
	PaddingX            unit.Dp
	PaddingY            unit.Dp
	TextSize            unit.Sp
	LineHeight          unit.Sp
	FocusRingWidth      unit.Dp
	InvalidOutlineWidth unit.Dp
	ShadowColor         color.NRGBA
	ShadowOpacity       float32
	ShadowStrength      float32
}

type InputGroupTheme struct {
	MinHeight           unit.Dp
	Radius              unit.Dp
	PaddingX            unit.Dp
	TextAreaMinHeight   unit.Dp
	TextAreaPaddingY    unit.Dp
	DividerWidth        unit.Dp
	TextSize            unit.Sp
	LineHeight          unit.Sp
	FocusRingWidth      unit.Dp
	InvalidOutlineWidth unit.Dp
	ShadowColor         color.NRGBA
	ShadowOpacity       float32
	ShadowStrength      float32
}
type CheckboxTheme struct {
	Size                unit.Dp
	FocusSpace          unit.Dp
	FocusRingWidth      unit.Dp
	BorderWidth         unit.Dp
	CheckStroke         unit.Dp
	IndeterminateStroke unit.Dp
	IndicatorSize       unit.Dp
	LabelGap            unit.Dp
	DescriptionGap      unit.Dp
	DescriptionIndent   unit.Dp
	ShadowOpacity       float32
}

type SwitchTheme struct {
	SmallTrackWidth   unit.Dp
	SmallTrackHeight  unit.Dp
	SmallThumbWidth   unit.Dp
	SmallThumbHeight  unit.Dp
	MediumTrackWidth  unit.Dp
	MediumTrackHeight unit.Dp
	MediumThumbWidth  unit.Dp
	MediumThumbHeight unit.Dp
	LargeTrackWidth   unit.Dp
	LargeTrackHeight  unit.Dp
	LargeThumbWidth   unit.Dp
	LargeThumbHeight  unit.Dp
	FocusSpace        unit.Dp
	FocusRingWidth    unit.Dp
	ContentGap        unit.Dp
	DescriptionGap    unit.Dp
	TextSize          unit.Sp
}

type SwitchGroupTheme struct {
	VerticalGap   unit.Dp
	HorizontalGap unit.Dp
}

type RadioGroupTheme struct {
	Size            unit.Dp
	FocusSpace      unit.Dp
	BorderWidth     unit.Dp
	ContentGap      unit.Dp
	VerticalGap     unit.Dp
	HorizontalGap   unit.Dp
	DescriptionGap  unit.Dp
	TextSize        unit.Sp
	DescriptionSize unit.Sp
	DotScale        float32
	DotPressedScale float32
	PressedScale    float32
}

type ProgressBarTheme struct {
	SmallHeight  unit.Dp
	MediumHeight unit.Dp
	LargeHeight  unit.Dp
	SmallRadius  unit.Dp
	MediumRadius unit.Dp
	LargeRadius  unit.Dp
	HeaderGap    unit.Dp
	TextSize     unit.Sp
}

type ProgressCircleTheme struct {
	SmallSize   unit.Dp
	MediumSize  unit.Dp
	LargeSize   unit.Dp
	StrokeRatio float32
}

type SpinnerTheme struct {
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	ExtraLargeSize unit.Dp
	StrokeRatio    float32
	InsetRatio     float32
}

type SliderTheme struct {
	TrackThickness  unit.Dp
	TrackRadius     unit.Dp
	EdgeInset       unit.Dp
	ThumbLength     unit.Dp
	ThumbExtra      unit.Dp
	HeaderGap       unit.Dp
	VerticalGap     unit.Dp
	TextSize        unit.Sp
	FocusRingWidth  unit.Dp
	FocusRingOffset unit.Dp
	DraggingScale   float32
}

type ListBoxTheme struct {
	Padding               unit.Dp
	Gap                   unit.Dp
	MaxHeight             unit.Dp
	SectionHeaderTextSize unit.Sp
	SectionHeaderPaddingX unit.Dp
	SectionHeaderPaddingY unit.Dp
	ItemMinHeight         unit.Dp
	ItemRadius            unit.Dp
	ItemPaddingX          unit.Dp
	ItemPaddingY          unit.Dp
	ItemContentGap        unit.Dp
	ItemTextSize          unit.Sp
	ItemDescriptionSize   unit.Sp
	ItemIndicatorSize     unit.Dp
	ItemIndicatorInset    unit.Dp
	ItemIndicatorStroke   unit.Dp
	FocusRingWidth        unit.Dp
	PressedScale          float32
}

type TreeTheme struct {
	Padding                   unit.Dp
	Gap                       unit.Dp
	MaxHeight                 unit.Dp
	RowHeight                 unit.Dp
	DescriptionRowHeight      unit.Dp
	RowRadius                 unit.Dp
	RowPaddingX               unit.Dp
	RowPaddingY               unit.Dp
	Indent                    unit.Dp
	ChevronSlotSize           unit.Dp
	ChevronIconSize           unit.Dp
	ContentGap                unit.Dp
	ItemTextSize              unit.Sp
	ItemDescriptionSize       unit.Sp
	FocusRingWidth            unit.Dp
	SurfaceRadius             unit.Dp
	DragPreviewOffset         unit.Dp
	DragPreviewMaxWidth       unit.Dp
	DragPreviewPaddingX       unit.Dp
	DragPreviewPaddingY       unit.Dp
	DragPreviewRadius         unit.Dp
	SmallPadding              unit.Dp
	SmallGap                  unit.Dp
	SmallRowHeight            unit.Dp
	SmallDescriptionRowHeight unit.Dp
	SmallRowRadius            unit.Dp
	SmallRowPaddingX          unit.Dp
	SmallRowPaddingY          unit.Dp
	SmallIndent               unit.Dp
	SmallChevronSlotSize      unit.Dp
	SmallChevronIconSize      unit.Dp
	SmallContentGap           unit.Dp
	SmallItemTextSize         unit.Sp
	SmallItemDescriptionSize  unit.Sp
}

type SidebarTheme struct {
	Width                 unit.Dp
	CollapsedWidth        unit.Dp
	Padding               unit.Dp
	ContentGap            unit.Dp
	ItemGap               unit.Dp
	ItemHeight            unit.Dp
	ItemRadius            unit.Dp
	ItemPaddingX          unit.Dp
	ItemContentGap        unit.Dp
	ItemTextSize          unit.Sp
	SectionTextSize       unit.Sp
	SectionHeight         unit.Dp
	SectionSeparatorInset unit.Dp
	BorderWidth           unit.Dp
	FocusRingWidth        unit.Dp
}

type ScrollbarTheme struct {
	TrackWidth     unit.Dp
	ThumbWidth     unit.Dp
	ContentGap     unit.Dp
	MinThumbLength unit.Dp
	MajorPadding   unit.Dp
	Radius         unit.Dp
	ThumbOpacity   float32
	HoverOpacity   float32
}

type SplitPaneTheme struct {
	DividerWidth unit.Dp
	HitSize      unit.Dp
	ActiveWidth  unit.Dp
	HandleLength unit.Dp
}

type TitleBarTheme struct {
	Height          unit.Dp
	PaddingX        unit.Dp
	LeadingGap      unit.Dp
	ControlWidth    unit.Dp
	IconSize        unit.Dp
	IconStrokeWidth unit.Dp
	TitleTextSize   unit.Sp
	BorderWidth     unit.Dp
	FocusRingWidth  unit.Dp
	ControlPressed  color.NRGBA
	CloseHover      color.NRGBA
	ClosePressed    color.NRGBA
}

type ToolbarTheme struct {
	Gap             unit.Dp
	Padding         unit.Dp
	Radius          unit.Dp
	SeparatorLength unit.Dp
	SeparatorWidth  unit.Dp
}

type TableTheme struct {
	RootPadding           unit.Dp
	RootRadius            unit.Dp
	HeaderRadius          unit.Dp
	BodyRadius            unit.Dp
	HeaderHeight          unit.Dp
	RowMinHeight          unit.Dp
	StripeBackground      color.NRGBA
	EmptyHeight           unit.Dp
	MaxHeight             unit.Dp
	MinColumnWidth        unit.Dp
	CellPaddingX          unit.Dp
	CellPaddingY          unit.Dp
	HeaderTextSize        unit.Sp
	CellTextSize          unit.Sp
	SeparatorWidth        unit.Dp
	ColumnSeparatorHeight unit.Dp
	ColumnResizerHitSize  unit.Dp
	ColumnResizerWidth    unit.Dp
	ColumnResizeStep      unit.Dp
	FocusRingWidth        unit.Dp
	FocusRadius           unit.Dp
	SelectionColumnWidth  unit.Dp
	SortIconSize          unit.Dp
	SortGap               unit.Dp
	FooterPaddingX        unit.Dp
	FooterPaddingY        unit.Dp
	LoadMoreHeight        unit.Dp
}

type PaginationTheme struct {
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	Radius         unit.Dp
	SmallTextSize  unit.Sp
	MediumTextSize unit.Sp
	LargeTextSize  unit.Sp
	SmallPaddingX  unit.Dp
	MediumPaddingX unit.Dp
	LargePaddingX  unit.Dp
	IconSize       unit.Dp
	ItemGap        unit.Dp
	ContentGap     unit.Dp
	NavGap         unit.Dp
	FocusRingWidth unit.Dp
	CompactWidth   unit.Dp
}

type MenuTheme struct {
	BackgroundColor color.NRGBA
	// HoverColor overrides the default item hover fill. An alpha of zero
	// falls back to Palette.DefaultHover.
	HoverColor                 color.NRGBA
	ForegroundColor            color.NRGBA
	MutedColor                 color.NRGBA
	IndicatorColor             color.NRGBA
	DangerColor                color.NRGBA
	FocusColor                 color.NRGBA
	BorderColor                color.NRGBA
	ShadowColor                color.NRGBA
	Width                      unit.Dp
	MaxHeight                  unit.Dp
	MaxWidthFraction           float32
	Padding                    unit.Dp
	Radius                     unit.Dp
	BorderWidth                unit.Dp
	ItemGap                    unit.Dp
	ItemMinHeight              unit.Dp
	ItemRadius                 unit.Dp
	ItemPaddingX               unit.Dp
	ItemPaddingY               unit.Dp
	ItemContentGap             unit.Dp
	ItemTextSize               unit.Sp
	ItemDescriptionSize        unit.Sp
	ShortcutTextSize           unit.Sp
	ShortcutHeight             unit.Dp
	ShortcutPaddingX           unit.Dp
	IndicatorSize              unit.Dp
	IndicatorContentGap        unit.Dp
	CheckmarkSize              unit.Dp
	RadioDotSize               unit.Dp
	IndicatorOffsetY           unit.Dp
	SubmenuIndicatorSize       unit.Dp
	FocusRingWidth             unit.Dp
	FocusRingOffset            unit.Dp
	PressedScale               float32
	SectionTextSize            unit.Sp
	SectionPaddingX            unit.Dp
	SectionPaddingTop          unit.Dp
	SectionPaddingBottom       unit.Dp
	SeparatorMarginX           unit.Dp
	SeparatorMarginY           unit.Dp
	SeparatorWidth             unit.Dp
	DescriptionLeadingHeight   unit.Dp
	DescriptionLeadingInsetTop unit.Dp
	SubmenuGap                 unit.Dp
	ContextMenuOffset          unit.Dp
	EnterScale                 float32
	ExitScale                  float32
	AnimationDistance          unit.Dp
	ShadowOpacity              float32
}

type DropdownTheme struct {
	FocusColor             color.NRGBA
	TriggerFocusRingWidth  unit.Dp
	TriggerFocusRingOffset unit.Dp
	TriggerFocusRadius     unit.Dp
	TriggerPressedScale    float32
	PanelGap               unit.Dp
	ArrowSize              unit.Dp
}

type MenubarTheme struct {
	TriggerHeight          unit.Dp
	TriggerPaddingX        unit.Dp
	TriggerRadius          unit.Dp
	TriggerTextSize        unit.Sp
	TriggerFocusRingWidth  unit.Dp
	TriggerFocusRingOffset unit.Dp
	Gap                    unit.Dp
	PanelGap               unit.Dp
}

type LineChartTheme struct {
	Height            unit.Dp
	PlotPaddingTop    unit.Dp
	PlotPaddingRight  unit.Dp
	PlotPaddingBottom unit.Dp
	PlotPaddingLeft   unit.Dp
	AxisNameGap       unit.Dp
	TickLabelGap      unit.Dp
	AxisTextSize      unit.Sp
	LegendTextSize    unit.Sp
	LegendMarkerWidth unit.Dp
	LegendMarkerSize  unit.Dp
	LegendItemGap     unit.Dp
	LegendLineGap     unit.Dp
	LegendGap         unit.Dp
	GridWidth         unit.Dp
	AxisWidth         unit.Dp
	LineWidth         unit.Dp
	PointSize         unit.Dp
	HoverPointSize    unit.Dp
	CrosshairWidth    unit.Dp
	TooltipGap        unit.Dp
	TooltipRowGap     unit.Dp
	TooltipMarkerSize unit.Dp
	SeriesColors      [9]color.NRGBA
}

type BarChartTheme struct {
	Height             unit.Dp
	PlotPaddingTop     unit.Dp
	PlotPaddingRight   unit.Dp
	PlotPaddingBottom  unit.Dp
	PlotPaddingLeft    unit.Dp
	AxisNameGap        unit.Dp
	TickLabelGap       unit.Dp
	AxisTextSize       unit.Sp
	LegendTextSize     unit.Sp
	LegendMarkerSize   unit.Dp
	LegendMarkerGap    unit.Dp
	LegendMarkerRadius unit.Dp
	LegendItemGap      unit.Dp
	LegendLineGap      unit.Dp
	LegendGap          unit.Dp
	GridWidth          unit.Dp
	AxisWidth          unit.Dp
	BarRadius          unit.Dp
	MinBarHeight       unit.Dp
	BackgroundRadius   unit.Dp
	TooltipGap         unit.Dp
	TooltipRowGap      unit.Dp
	TooltipMarkerSize  unit.Dp
	SeriesColors       [9]color.NRGBA
}

type PieChartTheme struct {
	Height             unit.Dp
	PlotPaddingTop     unit.Dp
	PlotPaddingRight   unit.Dp
	PlotPaddingBottom  unit.Dp
	PlotPaddingLeft    unit.Dp
	LegendTextSize     unit.Sp
	LegendMarkerSize   unit.Dp
	LegendMarkerGap    unit.Dp
	LegendMarkerRadius unit.Dp
	LegendItemGap      unit.Dp
	LegendLineGap      unit.Dp
	LegendGap          unit.Dp
	LabelTextSize      unit.Sp
	LabelLineLength    unit.Dp
	LabelLineLength2   unit.Dp
	LabelLineWidth     unit.Dp
	LabelGap           unit.Dp
	EmphasisSize       unit.Dp
	TooltipGap         unit.Dp
	TooltipRowGap      unit.Dp
	TooltipMarkerSize  unit.Dp
	SeriesColors       [9]color.NRGBA
}

type CandlestickChartTheme struct {
	Height                   unit.Dp
	PlotPaddingTop           unit.Dp
	PlotPaddingRight         unit.Dp
	PlotPaddingBottom        unit.Dp
	PlotPaddingLeft          unit.Dp
	AxisNameGap              unit.Dp
	TickLabelGap             unit.Dp
	AxisTextSize             unit.Sp
	GridWidth                unit.Dp
	AxisWidth                unit.Dp
	CrosshairWidth           unit.Dp
	CrosshairLabelPadding    unit.Dp
	CrosshairLabelBackground color.NRGBA
	CrosshairLabelForeground color.NRGBA
	WickWidth                unit.Dp
	BorderWidth              unit.Dp
	EmphasisBorderWidth      unit.Dp
	TooltipGap               unit.Dp
	TooltipRowGap            unit.Dp
	UpColor                  color.NRGBA
	DownColor                color.NRGBA
	DojiColor                color.NRGBA
}

// HeatmapTheme holds Heatmap defaults.
type HeatmapTheme struct {
	Height            unit.Dp
	PlotPaddingTop    unit.Dp
	PlotPaddingRight  unit.Dp
	PlotPaddingBottom unit.Dp
	PlotPaddingLeft   unit.Dp
	AxisTextSize      unit.Sp
	TickLabelGap      unit.Dp
	CellSize          unit.Dp
	CellGap           unit.Dp
	CellRadius        unit.Dp
	MinColor          color.NRGBA
	MaxColor          color.NRGBA
	EmptyColor        color.NRGBA
	TooltipGap        unit.Dp
}

// GanttChartTheme holds GanttChart defaults.
type GanttChartTheme struct {
	Height            unit.Dp
	PlotPaddingTop    unit.Dp
	PlotPaddingRight  unit.Dp
	PlotPaddingBottom unit.Dp
	PlotPaddingLeft   unit.Dp
	AxisNameGap       unit.Dp
	TickLabelGap      unit.Dp
	AxisTextSize      unit.Sp
	GridWidth         unit.Dp
	AxisWidth         unit.Dp
	RowHeight         unit.Dp
	BarHeight         unit.Dp
	BarRadius         unit.Dp
	TaskIndent        unit.Dp
	TaskToggleSize    unit.Dp
	TaskToggleGap     unit.Dp
	BaselineHeight    unit.Dp
	TaskLabelPaddingX unit.Dp
	DependencyWidth   unit.Dp
	DependencyDash    unit.Dp
	MarkerWidth       unit.Dp
	MarkerLabelGap    unit.Dp
	LegendGap         unit.Dp
	LegendTextSize    unit.Sp
	LegendMarkerSize  unit.Dp
	LegendMarkerGap   unit.Dp
	LegendItemGap     unit.Dp
	LegendLineGap     unit.Dp
	TooltipGap        unit.Dp
	SeriesColors      [9]color.NRGBA
}

type TabsTheme struct {
	RootGap             unit.Dp
	ListPadding         unit.Dp
	ListRadius          unit.Dp
	TabHeight           unit.Dp
	SmallTabHeight      unit.Dp
	LargeTabHeight      unit.Dp
	TabMinWidth         unit.Dp
	TabPaddingX         unit.Dp
	SmallTabPaddingX    unit.Dp
	LargeTabPaddingX    unit.Dp
	TabGap              unit.Dp
	TextSize            unit.Sp
	LargeTextSize       unit.Sp
	IconSize            unit.Dp
	IconGap             unit.Dp
	CloseButtonSize     unit.Dp
	CloseButtonGap      unit.Dp
	ExtraContentGap     unit.Dp
	IndicatorRadius     unit.Dp
	IndicatorLineWidth  unit.Dp
	IndicatorWidth      unit.Dp
	IndicatorMinWidth   unit.Dp
	IndicatorInset      unit.Dp
	FocusRingWidth      unit.Dp
	SeparatorWidth      unit.Dp
	PanelPadding        unit.Dp
	PanelGap            unit.Dp
	ColorDuration       time.Duration
	IndicatorDuration   time.Duration
	PanelDuration       time.Duration
	ScrollButtonSize    unit.Dp
	ScrollButtonInset   unit.Dp
	ScrollShadowSize    unit.Dp
	ScrollChevronSize   unit.Dp
	ScrollChevronStroke unit.Dp
}

// NodeGraphTheme controls canvas, node, edge, and selection colors for
// node-based graph editors.
type NodeGraphTheme struct {
	CanvasBackground    color.NRGBA
	CanvasBorder        color.NRGBA
	CanvasRadius        unit.Dp
	CanvasBorderWidth   unit.Dp
	GridColor           color.NRGBA
	GridOpacity         float32
	NodeBackground      color.NRGBA
	NodeBorder          color.NRGBA
	NodeForeground      color.NRGBA
	NodeMutedForeground color.NRGBA
	PortColor           color.NRGBA
	PortBorder          color.NRGBA
	EdgeColor           color.NRGBA
	SelectedEdgeColor   color.NRGBA
	SelectedNodeBorder  color.NRGBA
	SelectionFill       color.NRGBA
	SelectionBorder     color.NRGBA
}

// WorkbenchTheme contains shell-level tokens shared by editor workbenches.
// It is intentionally separate from TabsTheme so an application can tune its
// chrome without changing ordinary tabs elsewhere in the UI.
type WorkbenchTheme struct {
	SidebarWidth              unit.Dp
	SidebarMinWidth           unit.Dp
	SidebarBackground         color.NRGBA
	SidebarForeground         color.NRGBA
	SidebarHoverBackground    color.NRGBA
	SidebarActiveBackground   color.NRGBA
	SidebarActiveForeground   color.NRGBA
	SidebarBorder             color.NRGBA
	EditorBackground          color.NRGBA
	EditorForeground          color.NRGBA
	EditorTabBackground       color.NRGBA
	EditorTabHoverBackground  color.NRGBA
	EditorTabActiveBackground color.NRGBA
	EditorTabActiveForeground color.NRGBA
	BottomPanelBackground     color.NRGBA
	BottomPanelForeground     color.NRGBA
	BottomPanelBorder         color.NRGBA
	StatusBarBackground       color.NRGBA
	StatusBarForeground       color.NRGBA
	DividerColor              color.NRGBA
	DividerHoverColor         color.NRGBA
	DividerWidth              unit.Dp
	DividerHandleSize         unit.Dp
	TabHeight                 unit.Dp
	TabPaddingX               unit.Dp
	TabGap                    unit.Dp
	GroupGap                  unit.Dp
	Density                   float32
}

// EffectiveDensity returns a usable density multiplier for shell dimensions.
// Invalid or non-positive custom values fall back to the default density.
func (t WorkbenchTheme) EffectiveDensity() float32 {
	if math.IsNaN(float64(t.Density)) || math.IsInf(float64(t.Density), 0) || t.Density <= 0 {
		return 1
	}
	return t.Density
}

// Scale applies the Workbench density to a dimension token.
func (t WorkbenchTheme) Scale(value unit.Dp) unit.Dp {
	return unit.Dp(float32(value) * t.EffectiveDensity())
}

type CollapsibleTheme struct {
	BodyPadding       unit.Dp
	IndicatorSize     unit.Dp
	IndicatorStroke   unit.Dp
	ContentDuration   time.Duration
	IndicatorDuration time.Duration
}

type SelectTheme struct {
	Height            unit.Dp
	Radius            unit.Dp
	TextSize          unit.Sp
	ContentGap        unit.Dp
	TriggerPaddingX   unit.Dp
	TriggerPaddingY   unit.Dp
	IndicatorWidth    unit.Dp
	IndicatorSize     unit.Dp
	IndicatorStroke   unit.Dp
	PanelGap          unit.Dp
	PanelRadius       unit.Dp
	PanelMaxHeight    unit.Dp
	PanelPadding      unit.Dp
	AnimationScale    float32
	AnimationDistance unit.Dp
}

type PopoverTheme struct {
	Offset            unit.Dp
	Padding           unit.Dp
	Radius            unit.Dp
	MaxWidth          unit.Dp
	ArrowWidth        unit.Dp
	ArrowHeight       unit.Dp
	HeadingSize       unit.Sp
	BodyTextSize      unit.Sp
	SectionGap        unit.Dp
	AnimationScale    float32
	AnimationDistance unit.Dp
}

type TooltipTheme struct {
	Offset            unit.Dp
	ArrowOffset       unit.Dp
	Padding           unit.Dp
	Radius            unit.Dp
	BorderWidth       unit.Dp
	MaxWidth          unit.Dp
	ArrowSize         unit.Dp
	TextSize          unit.Sp
	AnimationScale    float32
	ExitScale         float32
	AnimationDistance unit.Dp
	Delay             time.Duration
	CloseDelay        time.Duration
}

type ToastTheme struct {
	Width             unit.Dp
	Inset             unit.Dp
	Gap               unit.Dp
	PaddingX          unit.Dp
	PaddingY          unit.Dp
	Radius            unit.Dp
	ContentGap        unit.Dp
	IndicatorPadding  unit.Dp
	IndicatorSize     unit.Dp
	CloseInset        unit.Dp
	FocusRingWidth    unit.Dp
	TitleSize         unit.Sp
	DescriptionSize   unit.Sp
	MaxVisible        int
	ScaleFactor       float32
	AnimationDuration time.Duration
	DefaultTimeout    time.Duration
}

type ModalTheme struct {
	XSmallWidth          unit.Dp
	SmallWidth           unit.Dp
	MediumWidth          unit.Dp
	LargeWidth           unit.Dp
	Margin               unit.Dp
	DesktopMargin        unit.Dp
	DesktopBreakpoint    unit.Dp
	Radius               unit.Dp
	Padding              unit.Dp
	HeaderGap            unit.Dp
	BodyGap              unit.Dp
	FooterGap            unit.Dp
	SectionGap           unit.Dp
	IconSize             unit.Dp
	CloseInset           unit.Dp
	TitleSize            unit.Sp
	BodyTextSize         unit.Sp
	Backdrop             color.NRGBA
	BlurBackdrop         color.NRGBA
	AnimationScale       float32
	AnimationDistance    unit.Dp
	AnimationBounceScale float32
}

type ComboBoxTheme struct {
	Height              unit.Dp
	Radius              unit.Dp
	TextSize            unit.Sp
	TriggerWidth        unit.Dp
	PanelGap            unit.Dp
	PanelPadding        unit.Dp
	PanelRadius         unit.Dp
	PanelMaxHeight      unit.Dp
	ItemHeight          unit.Dp
	ItemRadius          unit.Dp
	ItemPaddingX        unit.Dp
	ItemPaddingY        unit.Dp
	ItemTextSize        unit.Sp
	ItemDescriptionSize unit.Sp
	ItemCheckSize       unit.Dp
	ItemCheckInset      unit.Dp
	ItemCheckStroke     unit.Dp
	ChevronStroke       unit.Dp
}

type DatePickerTheme struct {
	Height             unit.Dp
	Radius             unit.Dp
	RangeRadius        unit.Dp
	TextSize           unit.Sp
	FieldGap           unit.Dp
	SegmentHeight      unit.Dp
	SegmentRadius      unit.Dp
	YearSegmentWidth   unit.Dp
	SegmentWidth       unit.Dp
	SeparatorWidth     unit.Dp
	RangeSeparatorSize unit.Dp
	TriggerWidth       unit.Dp
	IconSize           unit.Dp
	IconRadius         unit.Dp
	IconStroke         unit.Dp
	PopoverGap         unit.Dp
	PopoverPadding     unit.Dp
	PopoverRadius      unit.Dp
	PopoverMaxHeight   unit.Dp
	CalendarWidth      unit.Dp
	CellSize           unit.Dp
	MonthCellHeight    unit.Dp
	YearCellHeight     unit.Dp
	YearGridGap        unit.Dp
	NavButtonSize      unit.Dp
	NavChevronStroke   unit.Dp
	HeaderGap          unit.Dp
	HeaderIconSize     unit.Dp
	HeaderTextSize     unit.Sp
	WeekdayTextSize    unit.Sp
	CellTextSize       unit.Sp
	CellStrikeWidth    unit.Dp
	CellStrikeHalfSize unit.Dp
}

type ColorPickerTheme struct {
	TriggerGap        unit.Dp
	TriggerRadius     unit.Dp
	TriggerTextSize   unit.Sp
	FocusRingWidth    unit.Dp
	PanelGap          unit.Dp
	PanelWidth        unit.Dp
	PanelPadding      unit.Dp
	PanelRadius       unit.Dp
	PanelMaxHeight    unit.Dp
	ContentGap        unit.Dp
	CompactContentGap unit.Dp
}

type ColorWheelTheme struct {
	Size             unit.Dp
	ThumbSize        unit.Dp
	ThumbBorderWidth unit.Dp
	FocusRingWidth   unit.Dp
}

type ColorAreaTheme struct {
	Size              unit.Dp
	Radius            unit.Dp
	ThumbSize         unit.Dp
	DraggingThumbSize unit.Dp
	ThumbBorderWidth  unit.Dp
	FocusRingWidth    unit.Dp
	DotSize           unit.Dp
	DotGap            unit.Dp
}

type ColorFieldTheme struct {
	Gap unit.Dp
}

type ColorSliderTheme struct {
	TextSize         unit.Sp
	HeaderGap        unit.Dp
	TrackHeight      unit.Dp
	ThumbSize        unit.Dp
	ThumbBorderWidth unit.Dp
	FocusRingWidth   unit.Dp
}

type ColorSwatchTheme struct {
	ExtraSmallSize unit.Dp
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	ExtraLargeSize unit.Dp
	SquareRadius   unit.Dp
}

type ColorSwatchPickerTheme struct {
	Gap                          unit.Dp
	ExtraSmallBorderWidth        unit.Dp
	BorderWidth                  unit.Dp
	LargeBorderWidth             unit.Dp
	FocusRingWidth               unit.Dp
	FocusRingGap                 unit.Dp
	CheckStroke                  unit.Dp
	SquareItemRadiusExtraSmall   unit.Dp
	SquareItemRadiusSmall        unit.Dp
	SquareItemRadius             unit.Dp
	SquareSwatchRadiusExtraSmall unit.Dp
	SquareSwatchRadiusSmall      unit.Dp
	SquareSwatchRadius           unit.Dp
	SquareSelectedSmallRadius    unit.Dp
	ShadowOpacity                float32
	SelectedScale                float32
}

func DefaultTheme() Theme {
	theme := Theme{
		Palette: Palette{
			Background:                 color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			Surface:                    color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			SurfaceForeground:          color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			SurfaceSecondary:           color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff},
			SurfaceSecondaryForeground: color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			SurfaceTertiary:            color.NRGBA{R: 0xec, G: 0xec, B: 0xee, A: 0xff},
			SurfaceTertiaryForeground:  color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			SurfaceHover:               color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff},
			SurfacePressed:             color.NRGBA{R: 0xec, G: 0xec, B: 0xee, A: 0xff},
			SurfaceRaised:              color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff},
			Overlay:                    color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			OverlayForeground:          color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			Foreground:                 color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			MutedForeground:            color.NRGBA{R: 0x76, G: 0x76, B: 0x7a, A: 0xff},
			Border:                     color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff},
			Separator:                  color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff},
			Default:                    color.NRGBA{R: 0xeb, G: 0xeb, B: 0xec, A: 0xff},
			DefaultForeground:          color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
			DefaultHover:               color.NRGBA{R: 0xe1, G: 0xe1, B: 0xe2, A: 0xff},
			FieldBackground:            color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			FieldHover:                 color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf9, A: 0xff},
			FieldForeground:            color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			FieldPlaceholder:           color.NRGBA{R: 0x76, G: 0x76, B: 0x7a, A: 0xff},
			FieldFocus:                 color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			Segment:                    color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			SegmentForeground:          color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			Accent:                     color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff},
			AccentHover:                color.NRGBA{R: 0x1a, G: 0x7f, B: 0xf0, A: 0xff},
			AccentPressed:              color.NRGBA{R: 0x00, G: 0x5f, B: 0xce, A: 0xff},
			AccentForeground:           color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
			AccentSoft:                 color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x22},
			AccentSoftHover:            color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x33},
			AccentSoftForeground:       color.NRGBA{R: 0x00, G: 0x56, B: 0xbd, A: 0xff},
			Success:                    color.NRGBA{R: 0x17, G: 0xc9, B: 0x64, A: 0xff},
			SuccessForeground:          color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			SuccessSoft:                color.NRGBA{R: 0x17, G: 0xc9, B: 0x64, A: 0x26},
			SuccessSoftForeground:      color.NRGBA{R: 0x2b, G: 0x77, B: 0x45, A: 0xff},
			Warning:                    color.NRGBA{R: 0xf5, G: 0xa5, B: 0x24, A: 0xff},
			WarningForeground:          color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			WarningSoft:                color.NRGBA{R: 0xf5, G: 0xa5, B: 0x24, A: 0x26},
			WarningSoftForeground:      color.NRGBA{R: 0x85, G: 0x5f, B: 0x2e, A: 0xff},
			Danger:                     color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff},
			DangerHover:                color.NRGBA{R: 0xf5, G: 0x3a, B: 0x79, A: 0xff},
			DangerPressed:              color.NRGBA{R: 0xcf, G: 0x0b, B: 0x4f, A: 0xff},
			DangerForeground:           color.NRGBA{R: 0xff, G: 0xf7, B: 0xfb, A: 0xff},
			DangerSoft:                 color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0x26},
			DangerSoftHover:            color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0x33},
			DangerSoftForeground:       color.NRGBA{R: 0xba, G: 0x0f, B: 0x49, A: 0xff},
			Focus:                      color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff},
			Selection:                  color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x50},
			SurfaceShadow:              color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x34},
			OverlayShadow:              color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x68},
		},
		Typography: Typography{
			Typeface:     "sans-serif",
			MonoTypeface: "monospace",
			BodySize:     14,
			ControlSize:  14,
			SmallSize:    12,
		},
		Fonts: FontConfig{
			SystemFonts: true,
		},
		Shape: Shape{
			ControlRadius:  12,
			PopoverRadius:  20,
			ItemRadius:     16,
			CheckboxRadius: 4,
		},
		Spacing: Spacing{
			ControlHeight:        40,
			SmallControlHeight:   36,
			LargeControlHeight:   44,
			ControlPaddingX:      16,
			SmallControlPaddingX: 12,
			LargeControlPaddingX: 18,
			IconButtonSize:       36,
			PanelGap:             6,
			PanelPadding:         6,
			ItemHeight:           36,
		},
		Shadows: DefaultShadows(),
		Components: ComponentsTheme{
			Button: ButtonTheme{
				Radius:             24,
				BorderWidth:        1,
				ContentGap:         8,
				FocusRingWidth:     2,
				SpinnerStrokeWidth: 2,
				SpinnerSmall:       14,
				SpinnerMedium:      16,
				SpinnerLarge:       18,
				PressedScaleSmall:  0.98,
				PressedScaleMedium: 0.97,
				PressedScaleLarge:  0.96,
			},
			ButtonGroup: ButtonGroupTheme{
				SeparatorWidth:   1,
				SeparatorLength:  0.5,
				SeparatorOpacity: 0.15,
			},
			ToggleButton: ToggleButtonTheme{
				SmallHeight:        32,
				MediumHeight:       36,
				LargeHeight:        40,
				SmallPaddingX:      12,
				MediumPaddingX:     16,
				LargePaddingX:      16,
				Radius:             24,
				ContentGap:         8,
				SmallTextSize:      14,
				MediumTextSize:     14,
				LargeTextSize:      16,
				FocusRingWidth:     2,
				FocusRingOffset:    2,
				PressedScaleSmall:  0.98,
				PressedScaleMedium: 0.97,
				PressedScaleLarge:  0.96,
			},
			CloseButton: CloseButtonTheme{
				Size:           24,
				Radius:         12,
				Padding:        4,
				IconSize:       16,
				FocusRingWidth: 2,
				PressedScale:   0.93,
			},
			Chip: ChipTheme{
				SmallHeight:    20,
				MediumHeight:   24,
				LargeHeight:    28,
				SmallPaddingX:  4,
				MediumPaddingX: 8,
				LargePaddingX:  12,
				SmallPaddingY:  0,
				MediumPaddingY: 2,
				LargePaddingY:  4,
				LabelPaddingX:  2,
				ContentGap:     2,
				Radius:         16,
				SmallTextSize:  12,
				MediumTextSize: 12,
				LargeTextSize:  14,
				LineHeight:     20,
			},
			Avatar: AvatarTheme{
				SmallSize:      32,
				MediumSize:     40,
				LargeSize:      48,
				SmallRadius:    16,
				MediumRadius:   24,
				LargeRadius:    24,
				SmallTextSize:  14,
				MediumTextSize: 14,
				LargeTextSize:  16,
				SmallIconSize:  16,
				MediumIconSize: 20,
				LargeIconSize:  24,
			},
			Badge: BadgeTheme{
				SmallMinSize:         16,
				MediumMinSize:        28,
				LargeMinSize:         32,
				SmallRadius:          12,
				MediumRadius:         24,
				LargeRadius:          16,
				SmallTextSize:        10,
				MediumTextSize:       12,
				LargeTextSize:        14,
				SmallLineHeight:      14,
				MediumLineHeight:     16,
				LargeLineHeight:      20,
				LabelPaddingX:        2,
				BorderWidth:          1,
				PlacementOffsetRatio: 0.25,
			},
			Card: CardTheme{
				Padding: 16,
				Gap:     12,
				Radius:  24,
			},
			Alert: AlertTheme{
				PaddingX:              16,
				PaddingY:              12,
				Gap:                   16,
				Radius:                24,
				IndicatorPadding:      4,
				IconSize:              16,
				TitleSize:             14,
				TitleLineHeight:       24,
				DescriptionSize:       14,
				DescriptionLineHeight: 20,
			},
			AlertDialog: AlertDialogTheme{
				IconSize:      40,
				IconGlyphSize: 20,
				HeaderGap:     12,
				TitleSize:     16,
			},
			Description: DescriptionTheme{
				TextSize: 12,
			},
			Label: LabelTheme{
				TextSize:           14,
				RequiredMarkOffset: 2,
			},
			Input: InputTheme{
				Height:              36,
				Radius:              12,
				PaddingX:            12,
				TextSize:            14,
				LineHeight:          20,
				FocusRingWidth:      2,
				InvalidOutlineWidth: 1,
				ShadowColor:         color.NRGBA{A: 0xff},
				ShadowOpacity:       1,
				ShadowStrength:      1.5,
			},
			TextArea: TextAreaTheme{
				MinHeight:           38,
				Radius:              12,
				PaddingX:            12,
				PaddingY:            8,
				TextSize:            14,
				LineHeight:          20,
				FocusRingWidth:      2,
				InvalidOutlineWidth: 1,
				ShadowColor:         color.NRGBA{A: 0xff},
				ShadowOpacity:       1,
				ShadowStrength:      1.5,
			},
			InputGroup: InputGroupTheme{
				MinHeight:           36,
				Radius:              12,
				PaddingX:            12,
				TextAreaMinHeight:   38,
				TextAreaPaddingY:    8,
				DividerWidth:        0,
				TextSize:            14,
				LineHeight:          20,
				FocusRingWidth:      2,
				InvalidOutlineWidth: 1,
				ShadowColor:         color.NRGBA{A: 0xff},
				ShadowOpacity:       1,
				ShadowStrength:      1.5,
			},
			Checkbox: CheckboxTheme{
				Size:                16,
				FocusSpace:          2,
				FocusRingWidth:      2,
				BorderWidth:         1,
				CheckStroke:         1.5,
				IndeterminateStroke: 1.5,
				IndicatorSize:       12,
				LabelGap:            8,
				DescriptionGap:      4,
				DescriptionIndent:   28,
				ShadowOpacity:       1,
			},
			Switch: SwitchTheme{
				SmallTrackWidth:   32,
				SmallTrackHeight:  16,
				SmallThumbWidth:   16.5,
				SmallThumbHeight:  12,
				MediumTrackWidth:  40,
				MediumTrackHeight: 20,
				MediumThumbWidth:  22,
				MediumThumbHeight: 16,
				LargeTrackWidth:   48,
				LargeTrackHeight:  24,
				LargeThumbWidth:   27.5,
				LargeThumbHeight:  20,
				FocusSpace:        2,
				FocusRingWidth:    2,
				ContentGap:        12,
				DescriptionGap:    4,
				TextSize:          14,
			},
			SwitchGroup: SwitchGroupTheme{
				VerticalGap:   16,
				HorizontalGap: 16,
			},
			RadioGroup: RadioGroupTheme{
				Size:            16,
				FocusSpace:      2,
				BorderWidth:     1,
				ContentGap:      12,
				VerticalGap:     16,
				HorizontalGap:   16,
				DescriptionGap:  2,
				TextSize:        14,
				DescriptionSize: 12,
				DotScale:        0.38,
				DotPressedScale: 0.5,
				PressedScale:    0.95,
			},
			ProgressBar: ProgressBarTheme{
				SmallHeight:  4,
				MediumHeight: 8,
				LargeHeight:  12,
				SmallRadius:  2,
				MediumRadius: 3,
				LargeRadius:  6,
				HeaderGap:    4,
				TextSize:     14,
			},
			ProgressCircle: ProgressCircleTheme{
				SmallSize:   20,
				MediumSize:  28,
				LargeSize:   36,
				StrokeRatio: 4.0 / 36.0,
			},
			Spinner: SpinnerTheme{
				SmallSize:      16,
				MediumSize:     24,
				LargeSize:      32,
				ExtraLargeSize: 40,
				StrokeRatio:    0.125,
				InsetRatio:     0.0625,
			},
			Slider: SliderTheme{
				TrackThickness:  20,
				TrackRadius:     12,
				EdgeInset:       12,
				ThumbLength:     24,
				ThumbExtra:      4,
				HeaderGap:       4,
				VerticalGap:     8,
				TextSize:        14,
				FocusRingWidth:  2,
				FocusRingOffset: 2,
				DraggingScale:   0.9,
			},
			ListBox: ListBoxTheme{
				Padding:               4,
				Gap:                   4,
				MaxHeight:             280,
				SectionHeaderTextSize: 12,
				SectionHeaderPaddingX: 10,
				SectionHeaderPaddingY: 4,
				ItemMinHeight:         36,
				ItemRadius:            16,
				ItemPaddingX:          10,
				ItemPaddingY:          6,
				ItemContentGap:        12,
				ItemTextSize:          14,
				ItemDescriptionSize:   12,
				ItemIndicatorSize:     18,
				ItemIndicatorInset:    4,
				ItemIndicatorStroke:   1.7,
				FocusRingWidth:        2,
				PressedScale:          0.98,
			},
			Tree: TreeTheme{
				Padding:                   4,
				Gap:                       4,
				MaxHeight:                 320,
				RowHeight:                 36,
				DescriptionRowHeight:      52,
				RowRadius:                 16,
				RowPaddingX:               8,
				RowPaddingY:               6,
				Indent:                    20,
				ChevronSlotSize:           20,
				ChevronIconSize:           16,
				ContentGap:                8,
				ItemTextSize:              14,
				ItemDescriptionSize:       12,
				FocusRingWidth:            2,
				SurfaceRadius:             24,
				DragPreviewOffset:         12,
				DragPreviewMaxWidth:       240,
				DragPreviewPaddingX:       10,
				DragPreviewPaddingY:       6,
				DragPreviewRadius:         6,
				SmallPadding:              2,
				SmallGap:                  1,
				SmallRowHeight:            24,
				SmallDescriptionRowHeight: 40,
				SmallRowRadius:            4,
				SmallRowPaddingX:          4,
				SmallRowPaddingY:          2,
				SmallIndent:               12,
				SmallChevronSlotSize:      16,
				SmallChevronIconSize:      12,
				SmallContentGap:           5,
				SmallItemTextSize:         13,
				SmallItemDescriptionSize:  11,
			},
			Sidebar: SidebarTheme{
				Width:                 248,
				CollapsedWidth:        64,
				Padding:               8,
				ContentGap:            8,
				ItemGap:               4,
				ItemHeight:            40,
				ItemRadius:            8,
				ItemPaddingX:          10,
				ItemContentGap:        12,
				ItemTextSize:          14,
				SectionTextSize:       12,
				SectionHeight:         28,
				SectionSeparatorInset: 10,
				BorderWidth:           1,
				FocusRingWidth:        2,
			},
			Scrollbar: ScrollbarTheme{
				TrackWidth:     10,
				ThumbWidth:     6,
				ContentGap:     4,
				MinThumbLength: 32,
				MajorPadding:   2,
				Radius:         3,
				ThumbOpacity:   0.15,
				HoverOpacity:   0.28,
			},
			SplitPane: SplitPaneTheme{
				DividerWidth: 1,
				HitSize:      12,
				ActiveWidth:  2,
				HandleLength: 32,
			},
			TitleBar: TitleBarTheme{
				Height:          35,
				PaddingX:        8,
				LeadingGap:      8,
				ControlWidth:    46,
				IconSize:        12,
				IconStrokeWidth: 1.25,
				TitleTextSize:   12,
				BorderWidth:     1,
				FocusRingWidth:  2,
				ControlPressed:  color.NRGBA{R: 0xda, G: 0xda, B: 0xdc, A: 0xff},
				CloseHover:      color.NRGBA{R: 0xc4, G: 0x2b, B: 0x1c, A: 0xff},
				ClosePressed:    color.NRGBA{R: 0xa3, G: 0x21, B: 0x16, A: 0xff},
			},
			Toolbar: ToolbarTheme{
				Gap:             8,
				Padding:         4,
				Radius:          24,
				SeparatorLength: 20,
				SeparatorWidth:  1,
			},
			Table: TableTheme{
				RootPadding:           4,
				RootRadius:            20,
				HeaderRadius:          16,
				BodyRadius:            16,
				HeaderHeight:          36,
				RowMinHeight:          44,
				EmptyHeight:           144,
				MaxHeight:             360,
				MinColumnWidth:        96,
				CellPaddingX:          16,
				CellPaddingY:          12,
				HeaderTextSize:        12,
				CellTextSize:          14,
				SeparatorWidth:        1,
				ColumnSeparatorHeight: 16,
				ColumnResizerHitSize:  16,
				ColumnResizerWidth:    2,
				ColumnResizeStep:      8,
				FocusRingWidth:        2,
				FocusRadius:           8,
				SelectionColumnWidth:  40,
				SortIconSize:          12,
				SortGap:               8,
				FooterPaddingX:        16,
				FooterPaddingY:        10,
				LoadMoreHeight:        48,
			},
			Pagination: PaginationTheme{
				SmallSize:      28,
				MediumSize:     32,
				LargeSize:      36,
				Radius:         24,
				SmallTextSize:  12,
				MediumTextSize: 14,
				LargeTextSize:  16,
				SmallPaddingX:  8,
				MediumPaddingX: 10,
				LargePaddingX:  12,
				IconSize:       14,
				ItemGap:        4,
				ContentGap:     16,
				NavGap:         6,
				FocusRingWidth: 2,
				CompactWidth:   520,
			},
			Menu: MenuTheme{
				BackgroundColor:            color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				ForegroundColor:            color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
				MutedColor:                 color.NRGBA{R: 0x71, G: 0x71, B: 0x7a, A: 0xff},
				IndicatorColor:             color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff},
				DangerColor:                color.NRGBA{R: 0xff, G: 0x38, B: 0x3c, A: 0xff},
				FocusColor:                 color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
				ShadowColor:                color.NRGBA{A: 0xff},
				Width:                      220,
				MaxHeight:                  0,
				MaxWidthFraction:           0.48,
				Padding:                    6,
				Radius:                     24,
				BorderWidth:                0,
				ItemGap:                    2,
				ItemMinHeight:              36,
				ItemRadius:                 16,
				ItemPaddingX:               10,
				ItemPaddingY:               6,
				ItemContentGap:             12,
				ItemTextSize:               14,
				ItemDescriptionSize:        12,
				ShortcutTextSize:           14,
				ShortcutHeight:             24,
				ShortcutPaddingX:           8,
				IndicatorSize:              16,
				IndicatorContentGap:        2,
				CheckmarkSize:              10,
				RadioDotSize:               8,
				IndicatorOffsetY:           1.5,
				SubmenuIndicatorSize:       14,
				FocusRingWidth:             2,
				FocusRingOffset:            2,
				PressedScale:               0.98,
				SectionTextSize:            12,
				SectionPaddingX:            8,
				SectionPaddingTop:          6,
				SectionPaddingBottom:       4,
				SeparatorMarginX:           6,
				SeparatorMarginY:           0,
				SeparatorWidth:             1,
				DescriptionLeadingHeight:   32,
				DescriptionLeadingInsetTop: 1,
				SubmenuGap:                 8,
				ContextMenuOffset:          2,
				EnterScale:                 0.9,
				ExitScale:                  0.95,
				AnimationDistance:          4,
				ShadowOpacity:              1,
			},
			Dropdown: DropdownTheme{
				FocusColor:             color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
				TriggerFocusRingWidth:  2,
				TriggerFocusRingOffset: 2,
				TriggerFocusRadius:     12,
				TriggerPressedScale:    0.97,
				PanelGap:               4,
				ArrowSize:              12,
			},
			Menubar: MenubarTheme{
				TriggerHeight:          32,
				TriggerPaddingX:        12,
				TriggerRadius:          8,
				TriggerTextSize:        14,
				TriggerFocusRingWidth:  2,
				TriggerFocusRingOffset: 1,
				Gap:                    0,
				PanelGap:               4,
			},
			LineChart: LineChartTheme{
				Height:            320,
				PlotPaddingTop:    12,
				PlotPaddingRight:  16,
				PlotPaddingBottom: 36,
				PlotPaddingLeft:   56,
				AxisNameGap:       6,
				TickLabelGap:      8,
				AxisTextSize:      12,
				LegendTextSize:    12,
				LegendMarkerWidth: 20,
				LegendMarkerSize:  6,
				LegendItemGap:     20,
				LegendLineGap:     8,
				LegendGap:         12,
				GridWidth:         1,
				AxisWidth:         1,
				LineWidth:         2,
				PointSize:         6,
				HoverPointSize:    10,
				CrosshairWidth:    1,
				TooltipGap:        12,
				TooltipRowGap:     6,
				TooltipMarkerSize: 7,
				SeriesColors: [9]color.NRGBA{
					{R: 0x50, G: 0x70, B: 0xdd, A: 0xff},
					{R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
					{R: 0x50, G: 0x53, B: 0x72, A: 0xff},
					{R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
					{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff},
					{R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
					{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff},
					{R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
					{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
				},
			},
			BarChart: BarChartTheme{
				Height:             320,
				PlotPaddingTop:     12,
				PlotPaddingRight:   16,
				PlotPaddingBottom:  36,
				PlotPaddingLeft:    56,
				AxisNameGap:        6,
				TickLabelGap:       8,
				AxisTextSize:       12,
				LegendTextSize:     12,
				LegendMarkerSize:   10,
				LegendMarkerGap:    7,
				LegendMarkerRadius: 2,
				LegendItemGap:      20,
				LegendLineGap:      8,
				LegendGap:          12,
				GridWidth:          1,
				AxisWidth:          1,
				BarRadius:          0,
				MinBarHeight:       0,
				BackgroundRadius:   0,
				TooltipGap:         12,
				TooltipRowGap:      6,
				TooltipMarkerSize:  8,
				SeriesColors: [9]color.NRGBA{
					{R: 0x50, G: 0x70, B: 0xdd, A: 0xff},
					{R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
					{R: 0x50, G: 0x53, B: 0x72, A: 0xff},
					{R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
					{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff},
					{R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
					{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff},
					{R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
					{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
				},
			},
			PieChart: PieChartTheme{
				Height:             360,
				PlotPaddingTop:     12,
				PlotPaddingRight:   16,
				PlotPaddingBottom:  12,
				PlotPaddingLeft:    16,
				LegendTextSize:     12,
				LegendMarkerSize:   10,
				LegendMarkerGap:    7,
				LegendMarkerRadius: 2,
				LegendItemGap:      20,
				LegendLineGap:      8,
				LegendGap:          12,
				LabelTextSize:      12,
				LabelLineLength:    15,
				LabelLineLength2:   30,
				LabelLineWidth:     1,
				LabelGap:           5,
				EmphasisSize:       5,
				TooltipGap:         12,
				TooltipRowGap:      6,
				TooltipMarkerSize:  8,
				SeriesColors: [9]color.NRGBA{
					{R: 0x50, G: 0x70, B: 0xdd, A: 0xff},
					{R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
					{R: 0x50, G: 0x53, B: 0x72, A: 0xff},
					{R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
					{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff},
					{R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
					{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff},
					{R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
					{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
				},
			},
			CandlestickChart: CandlestickChartTheme{
				Height:                360,
				PlotPaddingTop:        12,
				PlotPaddingRight:      16,
				PlotPaddingBottom:     36,
				PlotPaddingLeft:       56,
				AxisNameGap:           6,
				TickLabelGap:          8,
				AxisTextSize:          12,
				GridWidth:             1,
				AxisWidth:             1,
				CrosshairWidth:        1,
				CrosshairLabelPadding: 4,
				WickWidth:             1,
				BorderWidth:           1,
				EmphasisBorderWidth:   2,
				TooltipGap:            12,
				TooltipRowGap:         5,
				UpColor:               color.NRGBA{R: 0xeb, G: 0x54, B: 0x54, A: 0xff},
				DownColor:             color.NRGBA{R: 0x47, G: 0xb2, B: 0x62, A: 0xff},
			},
			Heatmap: HeatmapTheme{
				Height: 320, PlotPaddingTop: 28, PlotPaddingRight: 16,
				PlotPaddingBottom: 16, PlotPaddingLeft: 64, AxisTextSize: 12,
				TickLabelGap: 8, CellSize: 16, CellGap: 3, CellRadius: 2, TooltipGap: 12,
				MinColor:   color.NRGBA{R: 0xe8, G: 0xf1, B: 0xff, A: 0xff},
				MaxColor:   color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
				EmptyColor: color.NRGBA{A: 0},
			},
			GanttChart: GanttChartTheme{
				Height: 360, PlotPaddingTop: 28, PlotPaddingRight: 20,
				PlotPaddingBottom: 34, PlotPaddingLeft: 140, AxisNameGap: 6, TickLabelGap: 8,
				AxisTextSize: 12, GridWidth: 1, AxisWidth: 1, RowHeight: 36,
				BarHeight: 18, BarRadius: 3, TaskIndent: 14, TaskToggleSize: 12, TaskToggleGap: 4, BaselineHeight: 4, TaskLabelPaddingX: 6,
				DependencyWidth: 1, DependencyDash: 4, MarkerWidth: 1, MarkerLabelGap: 4,
				LegendGap: 10, LegendTextSize: 12, LegendMarkerSize: 10, LegendMarkerGap: 5,
				LegendItemGap: 14, LegendLineGap: 6, TooltipGap: 12,
				SeriesColors: [9]color.NRGBA{
					{R: 0x50, G: 0x70, B: 0xdd, A: 0xff},
					{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff},
					{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
					{R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
					{R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
					{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff},
					{R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
					{R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
					{R: 0x50, G: 0x53, B: 0x72, A: 0xff},
				},
			},
			Tabs: TabsTheme{
				RootGap:             8,
				ListPadding:         4,
				ListRadius:          20,
				TabHeight:           32,
				SmallTabHeight:      24,
				LargeTabHeight:      40,
				TabMinWidth:         80,
				TabPaddingX:         8,
				SmallTabPaddingX:    12,
				LargeTabPaddingX:    20,
				TabGap:              4,
				TextSize:            14,
				LargeTextSize:       16,
				IconSize:            16,
				IconGap:             1,
				CloseButtonSize:     16,
				CloseButtonGap:      4,
				ExtraContentGap:     8,
				IndicatorRadius:     24,
				IndicatorLineWidth:  2,
				IndicatorWidth:      0,
				IndicatorMinWidth:   24,
				IndicatorInset:      0,
				FocusRingWidth:      2,
				SeparatorWidth:      1,
				PanelPadding:        8,
				PanelGap:            16,
				ColorDuration:       150 * time.Millisecond,
				IndicatorDuration:   250 * time.Millisecond,
				PanelDuration:       200 * time.Millisecond,
				ScrollButtonSize:    16,
				ScrollButtonInset:   4,
				ScrollShadowSize:    64,
				ScrollChevronSize:   10,
				ScrollChevronStroke: 1.5,
			},
			Workbench: WorkbenchTheme{
				SidebarWidth:              240,
				SidebarMinWidth:           160,
				SidebarBackground:         color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff},
				SidebarForeground:         color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
				SidebarHoverBackground:    color.NRGBA{R: 0xec, G: 0xec, B: 0xee, A: 0xff},
				SidebarActiveBackground:   color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x22},
				SidebarActiveForeground:   color.NRGBA{R: 0x00, G: 0x56, B: 0xbd, A: 0xff},
				SidebarBorder:             color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff},
				EditorBackground:          color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				EditorForeground:          color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
				EditorTabBackground:       color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff},
				EditorTabHoverBackground:  color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff},
				EditorTabActiveBackground: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				EditorTabActiveForeground: color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
				BottomPanelBackground:     color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff},
				BottomPanelForeground:     color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
				BottomPanelBorder:         color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff},
				StatusBarBackground:       color.NRGBA{R: 0xec, G: 0xec, B: 0xee, A: 0xff},
				StatusBarForeground:       color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
				DividerColor:              color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff},
				DividerHoverColor:         color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff},
				DividerWidth:              1,
				DividerHandleSize:         4,
				TabHeight:                 32,
				TabPaddingX:               8,
				TabGap:                    0,
				GroupGap:                  4,
				Density:                   1,
			},
			Collapsible: CollapsibleTheme{
				BodyPadding:       8,
				IndicatorSize:     16,
				IndicatorStroke:   1.7,
				ContentDuration:   200 * time.Millisecond,
				IndicatorDuration: 250 * time.Millisecond,
			},
			Select: SelectTheme{
				Height:            36,
				Radius:            12,
				TextSize:          14,
				ContentGap:        4,
				TriggerPaddingX:   12,
				TriggerPaddingY:   8,
				IndicatorWidth:    28,
				IndicatorSize:     16,
				IndicatorStroke:   1.7,
				PanelGap:          6,
				PanelRadius:       24,
				PanelMaxHeight:    280,
				PanelPadding:      6,
				AnimationScale:    0.95,
				AnimationDistance: 4,
			},
			Popover: PopoverTheme{
				Offset:            8,
				Padding:           16,
				Radius:            24,
				MaxWidth:          320,
				ArrowWidth:        12,
				ArrowHeight:       7,
				HeadingSize:       14,
				BodyTextSize:      14,
				SectionGap:        8,
				AnimationScale:    0.90,
				AnimationDistance: 4,
			},
			Tooltip: TooltipTheme{
				Offset:            3,
				ArrowOffset:       7,
				Padding:           8,
				Radius:            12,
				BorderWidth:       1,
				MaxWidth:          320,
				ArrowSize:         12,
				TextSize:          12,
				AnimationScale:    0.90,
				ExitScale:         0.95,
				AnimationDistance: 4,
				Delay:             1500 * time.Millisecond,
				CloseDelay:        500 * time.Millisecond,
			},
			Toast: ToastTheme{
				Width:             460,
				Inset:             16,
				Gap:               12,
				PaddingX:          16,
				PaddingY:          12,
				Radius:            24,
				ContentGap:        6,
				IndicatorPadding:  4,
				IndicatorSize:     16,
				CloseInset:        -4,
				FocusRingWidth:    2,
				TitleSize:         14,
				DescriptionSize:   14,
				MaxVisible:        3,
				ScaleFactor:       0.05,
				AnimationDuration: 350 * time.Millisecond,
				DefaultTimeout:    4 * time.Second,
			},
			Modal: ModalTheme{
				XSmallWidth:          320,
				SmallWidth:           384,
				MediumWidth:          448,
				LargeWidth:           512,
				Margin:               16,
				DesktopMargin:        40,
				DesktopBreakpoint:    640,
				Radius:               28,
				Padding:              24,
				HeaderGap:            12,
				BodyGap:              8,
				FooterGap:            8,
				SectionGap:           20,
				IconSize:             40,
				CloseInset:           16,
				TitleSize:            16,
				BodyTextSize:         14,
				Backdrop:             color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x66},
				BlurBackdrop:         color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x7a},
				AnimationScale:       0.95,
				AnimationDistance:    24,
				AnimationBounceScale: 1.035,
			},
			ComboBox: ComboBoxTheme{
				Height:              40,
				Radius:              12,
				TextSize:            14,
				TriggerWidth:        36,
				PanelGap:            6,
				PanelPadding:        6,
				PanelRadius:         20,
				PanelMaxHeight:      240,
				ItemHeight:          36,
				ItemRadius:          16,
				ItemPaddingX:        10,
				ItemPaddingY:        6,
				ItemTextSize:        14,
				ItemDescriptionSize: 12,
				ItemCheckSize:       16,
				ItemCheckInset:      3,
				ItemCheckStroke:     1.6,
				ChevronStroke:       1.7,
			},
			DatePicker: DatePickerTheme{
				Height:             36,
				Radius:             12,
				RangeRadius:        8,
				TextSize:           14,
				FieldGap:           4,
				SegmentHeight:      24,
				SegmentRadius:      6,
				YearSegmentWidth:   38,
				SegmentWidth:       24,
				SeparatorWidth:     8,
				RangeSeparatorSize: 20,
				TriggerWidth:       36,
				IconSize:           16,
				IconRadius:         3,
				IconStroke:         1.4,
				PopoverGap:         6,
				PopoverPadding:     12,
				PopoverRadius:      20,
				PopoverMaxHeight:   360,
				CalendarWidth:      252,
				CellSize:           36,
				MonthCellHeight:    48,
				YearCellHeight:     32,
				YearGridGap:        4,
				NavButtonSize:      24,
				NavChevronStroke:   1.8,
				HeaderGap:          4,
				HeaderIconSize:     14,
				HeaderTextSize:     14,
				WeekdayTextSize:    12,
				CellTextSize:       14,
				CellStrikeWidth:    1,
				CellStrikeHalfSize: 7,
			},
			ColorPicker: ColorPickerTheme{
				TriggerGap:        12,
				TriggerRadius:     4,
				TriggerTextSize:   14,
				FocusRingWidth:    2,
				PanelGap:          6,
				PanelWidth:        248,
				PanelPadding:      12,
				PanelRadius:       20,
				PanelMaxHeight:    520,
				ContentGap:        12,
				CompactContentGap: 8,
			},
			ColorWheel: ColorWheelTheme{
				Size:             190,
				ThumbSize:        20,
				ThumbBorderWidth: 3,
				FocusRingWidth:   2,
			},
			ColorArea: ColorAreaTheme{
				Size:              224,
				Radius:            16,
				ThumbSize:         16,
				DraggingThumbSize: 20,
				ThumbBorderWidth:  3,
				FocusRingWidth:    2,
				DotSize:           2,
				DotGap:            8,
			},
			ColorField: ColorFieldTheme{
				Gap: 4,
			},
			ColorSlider: ColorSliderTheme{
				TextSize:         14,
				HeaderGap:        4,
				TrackHeight:      20,
				ThumbSize:        16,
				ThumbBorderWidth: 3,
				FocusRingWidth:   2,
			},
			ColorSwatch: ColorSwatchTheme{
				ExtraSmallSize: 16,
				SmallSize:      24,
				MediumSize:     32,
				LargeSize:      36,
				ExtraLargeSize: 40,
				SquareRadius:   6,
			},
			ColorSwatchPicker: ColorSwatchPickerTheme{
				Gap:                          8,
				ExtraSmallBorderWidth:        1,
				BorderWidth:                  2,
				LargeBorderWidth:             3,
				FocusRingWidth:               2,
				FocusRingGap:                 2,
				CheckStroke:                  1.5,
				SquareItemRadiusExtraSmall:   6,
				SquareItemRadiusSmall:        8,
				SquareItemRadius:             12,
				SquareSwatchRadiusExtraSmall: 6,
				SquareSwatchRadiusSmall:      8,
				SquareSwatchRadius:           8,
				SquareSelectedSmallRadius:    6,
				ShadowOpacity:                1,
				SelectedScale:                0.77,
			},
		},
		DisabledOpacity: 0.5,
		Motion: MotionTheme{
			Enabled:         true,
			DefaultDuration: 200 * time.Millisecond,
			DurationScale:   1,
		},
	}
	theme.Components.Table.StripeBackground = theme.Palette.SurfaceSecondary
	theme.Components.NodeGraph = nodeGraphThemeFor(theme.Palette)
	SyncMaterialTheme(&theme)
	return theme
}

// DefaultShadows returns the framework shadow profiles.
func DefaultShadows() ShadowsTheme {
	return ShadowsTheme{
		Surface: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{OffsetY: 1, Blur: 2, Opacity: .72},
			{OffsetY: 2, Blur: 4, Opacity: .48},
		}},
		Overlay: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{OffsetY: 2, Blur: 6, Opacity: .72},
			{OffsetY: 8, Blur: 22, Spread: 4, Opacity: .56},
		}},
		Menu: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{Blur: 6, Spread: 1, Opacity: .20},
			{OffsetY: 4, Blur: 12, Opacity: .10},
			{OffsetY: 12, Blur: 28, Spread: 2, Opacity: .10},
		}},
		Control: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{Blur: 1, Opacity: 15.0 / 104.0},
			{OffsetY: 1, Blur: 2, Opacity: 15.0 / 104.0},
			{OffsetY: 2, Blur: 4, Opacity: 10.0 / 104.0},
		}},
		Checkbox: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{Blur: 1, Opacity: 16.0 / 104.0},
			{OffsetY: 1, Blur: 2, Opacity: 20.0 / 104.0},
		}},
		SwitchThumb: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{OffsetY: 1, Blur: 3, Opacity: .16},
			{OffsetY: 2, Blur: 8, Opacity: .08},
		}},
	}
}

func DarkTheme() Theme {
	theme := DefaultTheme()
	theme.Palette = Palette{
		Background:                 color.NRGBA{R: 0x06, G: 0x06, B: 0x07, A: 0xff},
		Surface:                    color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		SurfaceForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		SurfaceSecondary:           color.NRGBA{R: 0x23, G: 0x23, B: 0x25, A: 0xff},
		SurfaceSecondaryForeground: color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		SurfaceTertiary:            color.NRGBA{R: 0x26, G: 0x27, B: 0x28, A: 0xff},
		SurfaceTertiaryForeground:  color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		SurfaceHover:               color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff},
		SurfacePressed:             color.NRGBA{R: 0x2e, G: 0x2e, B: 0x31, A: 0xff},
		SurfaceRaised:              color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff},
		Overlay:                    color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		OverlayForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		Foreground:                 color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		MutedForeground:            color.NRGBA{R: 0x9f, G: 0x9f, B: 0xa9, A: 0xff},
		Border:                     color.NRGBA{R: 0x28, G: 0x28, B: 0x2c, A: 0xff},
		Separator:                  color.NRGBA{R: 0x21, G: 0x21, B: 0x24, A: 0xff},
		Default:                    color.NRGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff},
		DefaultForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		DefaultHover:               color.NRGBA{R: 0x2e, G: 0x2e, B: 0x31, A: 0xff},
		FieldBackground:            color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		FieldHover:                 color.NRGBA{R: 0x1c, G: 0x1c, B: 0x1e, A: 0xeb},
		FieldForeground:            color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		FieldPlaceholder:           color.NRGBA{R: 0x9f, G: 0x9f, B: 0xa9, A: 0xff},
		FieldFocus:                 color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		Segment:                    color.NRGBA{R: 0x46, G: 0x46, B: 0x4c, A: 0xff},
		SegmentForeground:          color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		Accent:                     color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
		AccentHover:                color.NRGBA{R: 0x35, G: 0x92, B: 0xf9, A: 0xff},
		AccentPressed:              color.NRGBA{R: 0x00, G: 0x6f, B: 0xd8, A: 0xff},
		AccentForeground:           color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		AccentSoft:                 color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x1f},
		AccentSoftHover:            color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x29},
		AccentSoftForeground:       color.NRGBA{R: 0x61, G: 0xa8, B: 0xfb, A: 0xff},
		Success:                    color.NRGBA{R: 0x17, G: 0xc9, B: 0x64, A: 0xff},
		SuccessForeground:          color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		SuccessSoft:                color.NRGBA{R: 0x17, G: 0xc9, B: 0x64, A: 0x1f},
		SuccessSoftForeground:      color.NRGBA{R: 0x74, G: 0xd8, B: 0x8f, A: 0xff},
		Warning:                    color.NRGBA{R: 0xf7, G: 0xb7, B: 0x50, A: 0xff},
		WarningForeground:          color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff},
		WarningSoft:                color.NRGBA{R: 0xf7, G: 0xb7, B: 0x50, A: 0x1f},
		WarningSoftForeground:      color.NRGBA{R: 0xf9, G: 0xcb, B: 0x86, A: 0xff},
		Danger:                     color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0xff},
		DangerHover:                color.NRGBA{R: 0xe1, G: 0x54, B: 0x51, A: 0xff},
		DangerPressed:              color.NRGBA{R: 0xc6, G: 0x2f, B: 0x33, A: 0xff},
		DangerForeground:           color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
		DangerSoft:                 color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0x26},
		DangerSoftHover:            color.NRGBA{R: 0xdb, G: 0x3b, B: 0x3e, A: 0x33},
		DangerSoftForeground:       color.NRGBA{R: 0xeb, G: 0x78, B: 0x72, A: 0xff},
		Focus:                      color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
		Selection:                  color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0x58},
		OverlayShadow:              color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x4d},
	}
	theme.Shadows.Overlay = ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{{Blur: 1, Opacity: 1}}}
	theme.Components.Workbench.SidebarBackground = theme.Palette.SurfaceSecondary
	theme.Components.Workbench.SidebarForeground = theme.Palette.SurfaceSecondaryForeground
	theme.Components.Workbench.SidebarHoverBackground = theme.Palette.SurfaceHover
	theme.Components.Workbench.SidebarActiveBackground = theme.Palette.AccentSoft
	theme.Components.Workbench.SidebarActiveForeground = theme.Palette.AccentSoftForeground
	theme.Components.Workbench.SidebarBorder = theme.Palette.Border
	theme.Components.Workbench.EditorBackground = theme.Palette.Background
	theme.Components.Workbench.EditorForeground = theme.Palette.Foreground
	theme.Components.Workbench.EditorTabBackground = theme.Palette.SurfaceSecondary
	theme.Components.Workbench.EditorTabHoverBackground = theme.Palette.SurfaceHover
	theme.Components.Workbench.EditorTabActiveBackground = theme.Palette.Surface
	theme.Components.Workbench.EditorTabActiveForeground = theme.Palette.SurfaceForeground
	theme.Components.Workbench.BottomPanelBackground = theme.Palette.SurfaceSecondary
	theme.Components.Workbench.BottomPanelForeground = theme.Palette.SurfaceSecondaryForeground
	theme.Components.Workbench.BottomPanelBorder = theme.Palette.Border
	theme.Components.Workbench.StatusBarBackground = theme.Palette.SurfaceTertiary
	theme.Components.Workbench.StatusBarForeground = theme.Palette.SurfaceTertiaryForeground
	theme.Components.Workbench.DividerColor = theme.Palette.Border
	theme.Components.Workbench.DividerHoverColor = theme.Palette.Accent
	theme.Components.Input.ShadowOpacity = 0
	theme.Components.TextArea.ShadowOpacity = 0
	theme.Components.Menu.ShadowOpacity = 0
	theme.Components.Menu.BorderWidth = 1
	theme.Components.Menu.BorderColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x4d}
	theme.Components.Menu.BackgroundColor = theme.Palette.Overlay
	theme.Components.Menu.ForegroundColor = theme.Palette.OverlayForeground
	theme.Components.Menu.MutedColor = theme.Palette.MutedForeground
	theme.Components.Menu.IndicatorColor = theme.Palette.Accent
	theme.Components.Menu.DangerColor = theme.Palette.Danger
	theme.Components.Menu.FocusColor = theme.Palette.Focus
	theme.Components.Dropdown.FocusColor = theme.Palette.Focus
	theme.Components.Checkbox.ShadowOpacity = 0
	theme.Components.InputGroup.ShadowOpacity = 0
	theme.Components.ColorSwatchPicker.ShadowOpacity = 0
	theme.Components.TitleBar.ControlPressed = theme.Palette.SurfacePressed
	theme.Components.Table.StripeBackground = theme.Palette.SurfaceSecondary
	theme.Components.NodeGraph = nodeGraphThemeFor(theme.Palette)
	theme.Components.Modal.Backdrop = color.NRGBA{A: 0x99}
	theme.Components.Modal.BlurBackdrop = color.NRGBA{A: 0x99}
	SyncMaterialTheme(&theme)
	return theme
}

func nodeGraphThemeFor(palette Palette) NodeGraphTheme {
	return NodeGraphTheme{
		CanvasBackground:    palette.Background,
		CanvasBorder:        palette.Border,
		CanvasRadius:        8,
		CanvasBorderWidth:   1,
		GridColor:           palette.MutedForeground,
		GridOpacity:         .35,
		NodeBackground:      palette.FieldBackground,
		NodeBorder:          palette.Border,
		NodeForeground:      palette.SurfaceForeground,
		NodeMutedForeground: palette.MutedForeground,
		PortColor:           palette.Foreground,
		PortBorder:          palette.Surface,
		EdgeColor:           palette.Accent,
		SelectedEdgeColor:   palette.AccentHover,
		SelectedNodeBorder:  palette.MutedForeground,
		SelectionFill:       palette.AccentSoft,
		SelectionBorder:     palette.Accent,
	}
}

// MaterialOf returns the internal Gio material bridge for text/editor helpers.
// Callers outside this package must not treat it as part of the public theme API.
func MaterialOf(theme *Theme) *material.Theme {
	if theme == nil {
		return nil
	}
	return theme.material
}

// DetachMaterial gives theme a private material bridge copy so later mutations
// do not share state with another Theme value.
func DetachMaterial(theme *Theme) {
	if theme == nil || theme.material == nil {
		return
	}
	clone := *theme.material
	if clone.Shaper == theme.managedShaper {
		clone.Shaper = newTextShaper(theme.Fonts)
		theme.managedShaper = clone.Shaper
	}
	if clone.Face == theme.managedFace {
		clone.Face = theme.Typography.Typeface
		theme.managedFace = clone.Face
	}
	theme.material = &clone
}

// SyncMaterialTheme updates the Gio material bridge from FlowUI tokens.
func SyncMaterialTheme(theme *Theme) {
	if theme == nil {
		return
	}
	if theme.material == nil {
		theme.material = material.NewTheme()
	}
	if theme.material.Shaper == nil || theme.managedShaper == nil || theme.material.Shaper == theme.managedShaper {
		theme.material.Shaper = newTextShaper(theme.Fonts)
		theme.managedShaper = theme.material.Shaper
	}
	if theme.material.Face == theme.managedFace {
		theme.material.Face = theme.Typography.Typeface
		theme.managedFace = theme.material.Face
	}
	theme.material.TextSize = theme.Typography.BodySize
	theme.material.Palette.Fg = theme.Palette.Foreground
	theme.material.Palette.Bg = theme.Palette.Background
	theme.material.Palette.ContrastBg = theme.Palette.Accent
	theme.material.Palette.ContrastFg = theme.Palette.AccentForeground
}

func newTextShaper(fonts FontConfig) *text.Shaper {
	// Keep the common system-only path lazy. Gio's zero-value Shaper loads the
	// platform font map on first use; eagerly scanning every installed font for
	// every temporary theme copy can otherwise cause a large startup spike.
	if fonts.SystemFonts && len(fonts.Collection) == 0 {
		return new(text.Shaper)
	}
	options := make([]text.ShaperOption, 0, 2)
	if !fonts.SystemFonts {
		options = append(options, text.NoSystemFonts())
	}
	if len(fonts.Collection) > 0 {
		options = append(options, text.WithCollection(fonts.Collection))
	}
	return text.NewShaper(options...)
}

func (theme *Theme) DisabledColor(c color.NRGBA) color.NRGBA {
	opacity := theme.DisabledOpacityValue()
	c.A = byte(float32(c.A) * opacity)
	return c
}

// DisabledOpacityValue returns the configured disabled opacity clamped to [0, 1].
func (theme *Theme) DisabledOpacityValue() float32 {
	opacity := float32(0.5)
	if theme != nil {
		opacity = theme.DisabledOpacity
	}
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	return opacity
}
