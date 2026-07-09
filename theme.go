package flowui

import (
	"image/color"

	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Theme contains FlowUI design tokens. Gio's material theme is kept only as a
// bridge for lower-level editor and text helpers.
type Theme struct {
	Palette         Palette
	Typography      Typography
	Shape           Shape
	Spacing         Spacing
	Components      ComponentsTheme
	DisabledOpacity float32
	Material        *material.Theme
}

type Palette struct {
	Background           color.NRGBA
	Surface              color.NRGBA
	SurfaceHover         color.NRGBA
	SurfacePressed       color.NRGBA
	SurfaceRaised        color.NRGBA
	Foreground           color.NRGBA
	MutedForeground      color.NRGBA
	Border               color.NRGBA
	Accent               color.NRGBA
	AccentHover          color.NRGBA
	AccentForeground     color.NRGBA
	AccentSoft           color.NRGBA
	AccentSoftHover      color.NRGBA
	AccentSoftForeground color.NRGBA
	Danger               color.NRGBA
	DangerHover          color.NRGBA
	DangerForeground     color.NRGBA
	DangerSoft           color.NRGBA
	DangerSoftHover      color.NRGBA
	DangerSoftForeground color.NRGBA
	Focus                color.NRGBA
	Selection            color.NRGBA
	Shadow               color.NRGBA
}

type Typography struct {
	BodySize    unit.Sp
	ControlSize unit.Sp
	SmallSize   unit.Sp
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

type ComponentsTheme struct {
	Button     ButtonTheme
	Input      InputTheme
	Checkbox   CheckboxTheme
	RadioGroup RadioGroupTheme
	ComboBox   ComboBoxTheme
	DatePicker DatePickerTheme
}

type ButtonTheme struct {
	ContentGap         unit.Dp
	FocusRingWidth     unit.Dp
	SpinnerSmall       unit.Dp
	SpinnerMedium      unit.Dp
	SpinnerLarge       unit.Dp
	PressedScaleSmall  float32
	PressedScaleMedium float32
	PressedScaleLarge  float32
}

type InputTheme struct {
	PaddingX      unit.Dp
	ShadowOpacity float32
}
type CheckboxTheme struct {
	Size        unit.Dp
	FocusSpace  unit.Dp
	BorderWidth unit.Dp
	CheckStroke unit.Dp
	LabelGap    unit.Dp
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
	TextSize           unit.Sp
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
	HeaderTextSize     unit.Sp
	WeekdayTextSize    unit.Sp
	CellTextSize       unit.Sp
	CellStrikeWidth    unit.Dp
	CellStrikeHalfSize unit.Dp
}

func DefaultTheme() Theme {
	theme := Theme{
		Palette: Palette{
			Background:           color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			Surface:              color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			SurfaceHover:         color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff},
			SurfacePressed:       color.NRGBA{R: 0xec, G: 0xec, B: 0xee, A: 0xff},
			SurfaceRaised:        color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff},
			Foreground:           color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff},
			MutedForeground:      color.NRGBA{R: 0x76, G: 0x76, B: 0x7a, A: 0xff},
			Border:               color.NRGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff},
			Accent:               color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff},
			AccentHover:          color.NRGBA{R: 0x1a, G: 0x7f, B: 0xf0, A: 0xff},
			AccentForeground:     color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
			AccentSoft:           color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x22},
			AccentSoftHover:      color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x33},
			AccentSoftForeground: color.NRGBA{R: 0x00, G: 0x56, B: 0xbd, A: 0xff},
			Danger:               color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff},
			DangerHover:          color.NRGBA{R: 0xf5, G: 0x3a, B: 0x79, A: 0xff},
			DangerForeground:     color.NRGBA{R: 0xff, G: 0xf7, B: 0xfb, A: 0xff},
			DangerSoft:           color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0x26},
			DangerSoftHover:      color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0x33},
			DangerSoftForeground: color.NRGBA{R: 0xba, G: 0x0f, B: 0x49, A: 0xff},
			Focus:                color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x99},
			Selection:            color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0x40},
			Shadow:               color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x68},
		},
		Typography: Typography{
			BodySize:    14,
			ControlSize: 14,
			SmallSize:   12,
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
		Components: ComponentsTheme{
			Button: ButtonTheme{
				ContentGap:         8,
				FocusRingWidth:     2,
				SpinnerSmall:       14,
				SpinnerMedium:      16,
				SpinnerLarge:       18,
				PressedScaleSmall:  0.98,
				PressedScaleMedium: 0.97,
				PressedScaleLarge:  0.96,
			},
			Input: InputTheme{
				PaddingX:      12,
				ShadowOpacity: 1,
			},
			Checkbox: CheckboxTheme{
				Size:        16,
				FocusSpace:  2,
				BorderWidth: 1,
				CheckStroke: 1.5,
				LabelGap:    10,
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
				ItemCheckSize:       18,
				ItemCheckInset:      4,
				ItemCheckStroke:     1.6,
				ChevronStroke:       1.7,
			},
			DatePicker: DatePickerTheme{
				Height:             36,
				Radius:             12,
				TextSize:           14,
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
				HeaderTextSize:     14,
				WeekdayTextSize:    12,
				CellTextSize:       14,
				CellStrikeWidth:    1,
				CellStrikeHalfSize: 7,
			},
		},
		DisabledOpacity: 0.5,
		Material:        material.NewTheme(),
	}
	syncMaterialTheme(&theme)
	return theme
}

func DarkTheme() Theme {
	theme := DefaultTheme()
	theme.Palette.Background = color.NRGBA{R: 0x16, G: 0x18, B: 0x1d, A: 0xff}
	theme.Palette.Surface = color.NRGBA{R: 0x20, G: 0x23, B: 0x29, A: 0xff}
	theme.Palette.SurfaceHover = color.NRGBA{R: 0x29, G: 0x2d, B: 0x34, A: 0xff}
	theme.Palette.SurfacePressed = color.NRGBA{R: 0x32, G: 0x37, B: 0x40, A: 0xff}
	theme.Palette.SurfaceRaised = color.NRGBA{R: 0x25, G: 0x28, B: 0x2f, A: 0xff}
	theme.Palette.Foreground = color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff}
	theme.Palette.MutedForeground = color.NRGBA{R: 0xa1, G: 0xa1, B: 0xaa, A: 0xff}
	theme.Palette.Border = color.NRGBA{R: 0x3f, G: 0x3f, B: 0x46, A: 0xff}
	theme.Palette.Shadow = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x90}
	syncMaterialTheme(&theme)
	return theme
}

func syncMaterialTheme(theme *Theme) {
	if theme.Material == nil {
		theme.Material = material.NewTheme()
	}
	theme.Material.TextSize = theme.Typography.BodySize
	theme.Material.Palette.Fg = theme.Palette.Foreground
	theme.Material.Palette.Bg = theme.Palette.Background
	theme.Material.Palette.ContrastBg = theme.Palette.Accent
	theme.Material.Palette.ContrastFg = theme.Palette.AccentForeground
}

func (theme *Theme) DisabledColor(c color.NRGBA) color.NRGBA {
	opacity := theme.disabledOpacity()
	c.A = byte(float32(c.A) * opacity)
	return c
}

func (theme *Theme) disabledOpacity() float32 {
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
