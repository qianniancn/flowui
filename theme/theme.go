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

// Theme contains FlowUI semantic tokens and framework-level component values.
//
// The Gio material theme is an unexported substrate bridge for text/editor
// helpers. Applications theme through Palette, Typography, Components, Fonts,
// Motion, and an explicit visual provider — never through material.Theme.
type Theme struct {
	Palette    Palette
	Typography Typography
	Fonts      FontConfig
	Shadows    ShadowsTheme
	Motion     MotionTheme
	Components ComponentsTheme
	// Code is the syntax-highlighting palette. Assign it to customize token
	// colors; leave it zero to use the palette-derived default.
	Code            CodeColors
	DisabledOpacity float32
	// VisualProvider supplies component visuals for an explicit style family.
	// Neutral themes leave it nil so they do not imply a default appearance.
	VisualProvider VisualProvider
	material       *material.Theme
	managedShaper  *text.Shaper
	managedFace    font.Typeface
	// derived, when set, is the semantic layer computed from a Seed. It is then
	// the source of truth for colours and Palette is only its projection — see
	// UseSemantic.
	derived *Semantic
}

type MotionTheme struct {
	// Enabled controls all framework animations.
	Enabled bool
	// DefaultDuration is used by Tween when no duration is specified.
	DefaultDuration time.Duration
	// DurationScale scales animation time globally. Zero disables motion.
	DurationScale float32
	// ControlDuration is the style-family baseline for state feedback
	// (color, focus, press). Zero lets each component use its own constant.
	ControlDuration time.Duration
	// Ease is the style-family curve for state transitions.
	Ease EaseKind
}

// EaseKind names a transition curve for a style family. The mapping to an
// actual easing function lives resolver-side (animation.EasingFor) so rule
// tables stay data.
type EaseKind uint8

const (
	EaseDefault EaseKind = iota
	EaseLinear
	EaseOut
	EaseInOut
	EaseBack
)

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
	// WarningHover is the hover-step border/foreground of the warning scale.
	// Zero alpha falls back to Warning.
	WarningHover color.NRGBA
	WarningSoft  color.NRGBA
	// WarningSoftHover is the filled warning surface on hover. Zero falls
	// back to WarningSoft.
	WarningSoftHover      color.NRGBA
	WarningSoftForeground color.NRGBA
	Danger                color.NRGBA
	DangerHover           color.NRGBA
	DangerPressed         color.NRGBA
	DangerForeground      color.NRGBA
	DangerSoft            color.NRGBA
	DangerSoftHover       color.NRGBA
	DangerSoftForeground  color.NRGBA
	Focus                 color.NRGBA
	Selection             color.NRGBA
	// Disabled gray-scheme tokens: when DisabledBackground has alpha, disabled
	// states swap to the gray palette instead of fading opacity.
	DisabledBackground color.NRGBA
	DisabledForeground color.NRGBA
	DisabledBorder     color.NRGBA
	SurfaceShadow      color.NRGBA
	OverlayShadow      color.NRGBA
	// AccentGradient, when set, replaces the flat Accent fill on
	// accent-covered surfaces (e.g. the primary button). Nil keeps the flat color.
	// Once a derived layer is installed this is the projection of
	// Semantic.PrimaryGradient; set it there instead.
	AccentGradient *Gradient
}

// GradientStop is one color stop of a Gradient. An Offset outside [0, 1]
// is normalized by the brush, like style.Gradient.
type GradientStop struct {
	Offset float32
	Color  color.NRGBA
}

// Gradient is the theme-side gradient value. It mirrors style.Gradient in
// shape; components convert it at the declaration boundary.
type Gradient struct {
	AngleDegrees float32
	Stops        []GradientStop
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
	ToggleButton      ToggleButtonTheme
	Image             ImageTheme
	Code              CodeTheme
	Skeleton          SkeletonTheme
	Segmented         SegmentedTheme
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
	Icon              IconTheme
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
	Timeline          TimelineTheme
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

// ImageTheme is the provider-owned shape baseline for images.
type ImageTheme struct {
	Radius unit.Dp
}

// CodeTheme is the surface geometry of a highlighted code block.
type CodeTheme struct {
	Padding       unit.Dp
	Radius        unit.Dp
	HeaderGap     unit.Dp
	LineNumberGap unit.Dp
	TextSize      unit.Sp
}

// SkeletonTheme is the placeholder geometry and pulse timing.
type SkeletonTheme struct {
	Height unit.Dp
	Radius unit.Dp
	Period time.Duration
}

// SegmentedTheme is the track geometry of a segmented control.
type SegmentedTheme struct {
	Padding unit.Dp
	Radius  unit.Dp
	Gap     unit.Dp
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

type IconTheme struct {
	Size unit.Dp
}

type ToggleButtonTheme struct {
	SmallHeight        unit.Dp
	MediumHeight       unit.Dp
	LargeHeight        unit.Dp
	SmallPaddingX      unit.Dp
	MediumPaddingX     unit.Dp
	LargePaddingX      unit.Dp
	Radius             unit.Dp
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
	Height          unit.Dp
	SmallHeight     unit.Dp
	LargeHeight     unit.Dp
	Radius          unit.Dp
	SmallRadius     unit.Dp
	LargeRadius     unit.Dp
	PaddingX        unit.Dp
	SmallPaddingX   unit.Dp
	LargePaddingX   unit.Dp
	TextSize        unit.Sp
	LargeTextSize   unit.Sp
	LineHeight      unit.Sp
	LargeLineHeight unit.Sp
	FocusRingWidth  unit.Dp
	// FocusGlowBlur adds a blurred focus-color glow behind the ring. Zero keeps the hard ring.
	FocusGlowBlur unit.Dp
	// FocusHaloBlur adds a soft focus halo without changing the ring path.
	// Zero disables it.
	FocusHaloBlur unit.Dp
	// FocusHaloOpacity scales the halo colour. Zero uses full opacity.
	FocusHaloOpacity float32
	// FocusShadowSpread draws the focus/error/warning feedback as a spread
	// shadow ring of this width. Zero keeps the outline ring.
	FocusShadowSpread   unit.Dp
	InvalidOutlineWidth unit.Dp
	// BorderWidth draws the field box outline. Zero keeps the borderless field.
	BorderWidth    unit.Dp
	ShadowColor    color.NRGBA
	ShadowOpacity  float32
	ShadowStrength float32
}

type TextAreaTheme struct {
	MinHeight      unit.Dp
	Radius         unit.Dp
	PaddingX       unit.Dp
	PaddingY       unit.Dp
	TextSize       unit.Sp
	LineHeight     unit.Sp
	FocusRingWidth unit.Dp
	// FocusGlowBlur adds a blurred focus-color glow behind the ring. Zero keeps the hard ring.
	FocusGlowBlur unit.Dp
	// FocusHaloBlur adds a soft focus halo without changing the ring path.
	// Zero disables it.
	FocusHaloBlur unit.Dp
	// FocusHaloOpacity scales the halo colour. Zero uses full opacity.
	FocusHaloOpacity float32
	// FocusShadowSpread draws the focus/error/warning feedback as a spread
	// shadow ring of this width. Zero keeps the outline ring.
	FocusShadowSpread   unit.Dp
	InvalidOutlineWidth unit.Dp
	// BorderWidth draws the field box outline. Zero keeps the borderless field.
	BorderWidth    unit.Dp
	ShadowColor    color.NRGBA
	ShadowOpacity  float32
	ShadowStrength float32
}

type InputGroupTheme struct {
	MinHeight         unit.Dp
	Radius            unit.Dp
	PaddingX          unit.Dp
	ActionPadding     unit.Dp
	TextAreaMinHeight unit.Dp
	TextAreaPaddingY  unit.Dp
	DividerWidth      unit.Dp
	TextSize          unit.Sp
	LineHeight        unit.Sp
	FocusRingWidth    unit.Dp
	// FocusGlowBlur adds a blurred focus-color glow behind the ring. Zero keeps the hard ring.
	FocusGlowBlur unit.Dp
	// FocusHaloBlur adds a soft focus halo without changing the ring path.
	// Zero disables it.
	FocusHaloBlur unit.Dp
	// FocusHaloOpacity scales the halo colour. Zero uses full opacity.
	FocusHaloOpacity float32
	// FocusShadowSpread draws the focus/error/warning feedback as a spread
	// shadow ring of this width. Zero keeps the outline ring.
	FocusShadowSpread   unit.Dp
	InvalidOutlineWidth unit.Dp
	// BorderWidth draws the group box outline. Zero keeps the borderless group.
	BorderWidth    unit.Dp
	ShadowColor    color.NRGBA
	ShadowOpacity  float32
	ShadowStrength float32
}
type CheckboxTheme struct {
	Size                unit.Dp
	Radius              unit.Dp
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
	SmallSize    unit.Dp
	MediumSize   unit.Dp
	LargeSize    unit.Dp
	StrokeRatio  float32
	SpinDuration time.Duration
}

type SpinnerTheme struct {
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	ExtraLargeSize unit.Dp
	StrokeRatio    float32
	InsetRatio     float32
	SpinDuration   time.Duration
}

type SliderTheme struct {
	// ThumbSize sets a circular thumb independent of track thickness.
	// Zero retains the elongated thumb defined by ThumbLength and ThumbExtra.
	ThumbSize       unit.Dp
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
	InlineIndent          unit.Dp
	SwitcherSize          unit.Dp
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
	// HoverColor overrides the fill of a highlighted item. An alpha of zero
	// falls back to the accent fill.
	HoverColor color.NRGBA
	// SelectedColor is the fill of a selected item, such as a checked item.
	// An alpha of zero falls back to the accent tint.
	SelectedColor              color.NRGBA
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
	EnterDuration              time.Duration
	ExitDuration               time.Duration
	SubmenuOpenDelay           time.Duration
	AnimationDistance          unit.Dp
	ShadowOpacity              float32
}

type DropdownTheme struct {
	FocusColor             color.NRGBA
	TriggerFocusRingWidth  unit.Dp
	TriggerFocusRingOffset unit.Dp
	TriggerFocusRadius     unit.Dp
	TriggerPressedScale    float32
	PressDuration          time.Duration
	EnterDuration          time.Duration
	ExitDuration           time.Duration
	LongPressDuration      time.Duration
	HoverOpenDuration      time.Duration
	HoverCloseDuration     time.Duration
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
	Height              unit.Dp
	PlotPaddingTop      unit.Dp
	PlotPaddingRight    unit.Dp
	PlotPaddingBottom   unit.Dp
	PlotPaddingLeft     unit.Dp
	AxisNameGap         unit.Dp
	TickLabelGap        unit.Dp
	AxisTextSize        unit.Sp
	LegendTextSize      unit.Sp
	LegendMarkerWidth   unit.Dp
	LegendMarkerSize    unit.Dp
	LegendItemGap       unit.Dp
	LegendLineGap       unit.Dp
	LegendGap           unit.Dp
	GridWidth           unit.Dp
	AxisWidth           unit.Dp
	LineWidth           unit.Dp
	PointSize           unit.Dp
	AnnotationLineWidth unit.Dp
	AnnotationPointSize unit.Dp
	HoverPointSize      unit.Dp
	CrosshairWidth      unit.Dp
	TooltipGap          unit.Dp
	TooltipRowGap       unit.Dp
	TooltipMarkerSize   unit.Dp
	SeriesColors        [9]color.NRGBA
}

type BarChartTheme struct {
	Height              unit.Dp
	PlotPaddingTop      unit.Dp
	PlotPaddingRight    unit.Dp
	PlotPaddingBottom   unit.Dp
	PlotPaddingLeft     unit.Dp
	AxisNameGap         unit.Dp
	TickLabelGap        unit.Dp
	AxisTextSize        unit.Sp
	LegendTextSize      unit.Sp
	TextWeight          font.Weight
	EmphasisWeight      font.Weight
	LegendMarkerSize    unit.Dp
	LegendMarkerGap     unit.Dp
	LegendMarkerRadius  unit.Dp
	LegendItemGap       unit.Dp
	LegendLineGap       unit.Dp
	LegendGap           unit.Dp
	GridWidth           unit.Dp
	AxisWidth           unit.Dp
	AnnotationLineWidth unit.Dp
	AnnotationPointSize unit.Dp
	BarRadius           unit.Dp
	MinBarHeight        unit.Dp
	BackgroundRadius    unit.Dp
	TooltipGap          unit.Dp
	TooltipRowGap       unit.Dp
	TooltipMarkerSize   unit.Dp
	SeriesColors        [9]color.NRGBA
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
	AnnotationLineWidth      unit.Dp
	AnnotationPointSize      unit.Dp
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
	Height               unit.Dp
	PlotPaddingTop       unit.Dp
	PlotPaddingRight     unit.Dp
	PlotPaddingBottom    unit.Dp
	PlotPaddingLeft      unit.Dp
	AxisTextSize         unit.Sp
	AxisFontWeight       font.Weight
	TickLabelGap         unit.Dp
	CellSize             unit.Dp
	CellGap              unit.Dp
	CellRadius           unit.Dp
	SelectionBorderWidth unit.Dp
	MinColor             color.NRGBA
	MaxColor             color.NRGBA
	EmptyColor           color.NRGBA
	TooltipGap           unit.Dp
	TooltipMarkerSize    unit.Dp
	TooltipRowGap        unit.Dp
	EmptyFontWeight      font.Weight
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
	RootGap            unit.Dp
	ListPadding        unit.Dp
	ListRadius         unit.Dp
	TabHeight          unit.Dp
	SmallTabHeight     unit.Dp
	LargeTabHeight     unit.Dp
	TabMinWidth        unit.Dp
	TabPaddingX        unit.Dp
	SmallTabPaddingX   unit.Dp
	LargeTabPaddingX   unit.Dp
	TabGap             unit.Dp
	TextSize           unit.Sp
	LargeTextSize      unit.Sp
	IconSize           unit.Dp
	IconGap            unit.Dp
	CloseButtonSize    unit.Dp
	CloseButtonGap     unit.Dp
	ExtraContentGap    unit.Dp
	IndicatorRadius    unit.Dp
	IndicatorLineWidth unit.Dp
	IndicatorWidth     unit.Dp
	IndicatorMinWidth  unit.Dp
	IndicatorInset     unit.Dp
	FocusRingWidth     unit.Dp
	SeparatorWidth     unit.Dp
	PanelPadding       unit.Dp
	PanelGap           unit.Dp
	// Card-panel chrome: when PanelBackground has alpha the panel draws the
	// boxed card look.
	PanelBackground     color.NRGBA
	PanelBorderColor    color.NRGBA
	PanelRadius         unit.Dp
	ColorDuration       time.Duration
	IndicatorDuration   time.Duration
	PanelDuration       time.Duration
	ScrollButtonSize    unit.Dp
	ScrollButtonInset   unit.Dp
	ScrollShadowSize    unit.Dp
	ScrollChevronSize   unit.Dp
	ScrollChevronStroke unit.Dp
}

// TimelineTheme controls the rail, marker, and text spacing of a time line.
type TimelineTheme struct {
	TailColor            color.NRGBA
	TailWidth            unit.Dp
	DotBorderWidth       unit.Dp
	DotSize              unit.Dp
	DotBackground        color.NRGBA
	ItemPaddingBottom    unit.Dp
	TitleContentGap      unit.Dp
	SectionGap           unit.Dp
	InlineInset          unit.Dp
	HorizontalTitleGap   unit.Dp
	HorizontalContentGap unit.Dp
	TitleColor           color.NRGBA
	ContentColor         color.NRGBA
	MutedColor           color.NRGBA
	PrimaryColor         color.NRGBA
	ErrorColor           color.NRGBA
	SuccessColor         color.NRGBA
	DisabledColor        color.NRGBA
	HorizontalItemGap    unit.Dp
	ProcessDashLength    unit.Dp
	ProcessDashGap       unit.Dp
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
	ContentGap        unit.Dp
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
	PanelBorderWidth  unit.Dp
	AnimationScale    float32
	AnimationDistance unit.Dp
	OpenDuration      time.Duration
	CloseDuration     time.Duration
	IndicatorDuration time.Duration
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
	EnterDuration     time.Duration
	ExitDuration      time.Duration
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
	ActionInsetTop    unit.Dp
	IndicatorPadding  unit.Dp
	IndicatorSize     unit.Dp
	CloseInset        unit.Dp
	CloseButtonSize   unit.Dp
	FocusRingWidth    unit.Dp
	TitleSize         unit.Sp
	DescriptionSize   unit.Sp
	MaxVisible        int
	ScaleFactor       float32
	AnimationDuration time.Duration
	DefaultTimeout    time.Duration
}

type ModalTheme struct {
	XSmallWidth                unit.Dp
	SmallWidth                 unit.Dp
	MediumWidth                unit.Dp
	LargeWidth                 unit.Dp
	Margin                     unit.Dp
	DesktopMargin              unit.Dp
	DesktopBreakpoint          unit.Dp
	Radius                     unit.Dp
	Padding                    unit.Dp
	HeaderGap                  unit.Dp
	BodyGap                    unit.Dp
	FooterGap                  unit.Dp
	SectionGap                 unit.Dp
	IconSize                   unit.Dp
	CloseInset                 unit.Dp
	TitleSize                  unit.Sp
	BodyTextSize               unit.Sp
	Backdrop                   color.NRGBA
	BlurBackdrop               color.NRGBA
	AnimationScale             float32
	AnimationDuration          time.Duration
	AnimationDistance          unit.Dp
	AnimationBounceScale       float32
	AnimationBounceSettleScale float32
	AnimationZoomScale         float32
	AnimationPopStartScale     float32
	AnimationPopOvershootScale float32
	AnimationPopSettleScale    float32
}

type ComboBoxTheme struct {
	Height              unit.Dp
	Radius              unit.Dp
	TextSize            unit.Sp
	TriggerWidth        unit.Dp
	PanelGap            unit.Dp
	PanelPadding        unit.Dp
	PanelRadius         unit.Dp
	PanelBorderWidth    unit.Dp
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
	AnimationDuration   time.Duration
	ItemColorDuration   time.Duration
	ItemSelectDuration  time.Duration
	ItemPressInDuration time.Duration
	ItemPressDuration   time.Duration
	IconDuration        time.Duration
	AnimationScale      float32
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
	PopoverBorderWidth unit.Dp
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
	// CurrentTextColor overrides the normal in-month day label. Zero falls
	// back to Semantic.Foreground.
	CurrentTextColor color.NRGBA
	// OutsideTextColor overrides the adjacent-month day label. Zero falls back
	// to the disabled muted colour.
	OutsideTextColor   color.NRGBA
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
	EnterDuration     time.Duration
	ExitDuration      time.Duration
	EnterScale        float32
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
	SelectionDuration            time.Duration
	CheckDuration                time.Duration
	HoverDuration                time.Duration
}

// NeutralTheme returns the framework's provider-free baseline theme. It is
// intended for low-level tests and migration tooling; applications must use a
// complete theme from theme/base or an external Provider package.
func NeutralTheme() Theme { return New(DefaultSeed()) }

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

// NeutralDarkTheme returns the provider-free neutral seed on a dark canvas.
func NeutralDarkTheme() Theme {
	theme := NewDark(DefaultSeed())
	// A panel's drop shadow is dark, and a dark one on a black canvas reads as
	// nothing - so the raised surfaces drop theirs, the overlay carries a white
	// one, and the menu states its own edge instead of a shadow.
	theme.Shadows.Overlay = ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{{Blur: 1, Opacity: 1}}}
	theme.Components.Input.ShadowOpacity = 0
	theme.Components.TextArea.ShadowOpacity = 0
	theme.Components.Menu.ShadowOpacity = 0
	theme.Components.Menu.BorderWidth = 1
	theme.Components.Menu.BorderColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x4d}
	theme.Components.Checkbox.ShadowOpacity = 0
	theme.Components.InputGroup.ShadowOpacity = 0
	theme.Components.ColorSwatchPicker.ShadowOpacity = 0
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
