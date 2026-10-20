package theme

// ButtonSize identifies the control size exposed to visual renderers.
type ButtonSize uint8

const (
	ButtonMedium ButtonSize = iota
	ButtonSmall
	ButtonLarge
)

// ButtonVariant identifies a button's semantic visual treatment.
type ButtonVariant uint8

const (
	ButtonPrimary ButtonVariant = iota
	ButtonSecondary
	ButtonTertiary
	ButtonGhost
	ButtonOutline
	ButtonDanger
	ButtonDangerSoft
	// ButtonDangerOutline is the bordered danger button: resting surface,
	// danger border and text, hover/pressed recolor both to the danger scale.
	ButtonDangerOutline
)

// InputStyleMode identifies the field surface receiving a default declaration.
type InputStyleMode uint8

const (
	InputModeSingleLine InputStyleMode = iota
	InputModeTextArea
	InputModeGroup
)

// CheckboxVariant identifies the semantic visual treatment of a checkbox.
type CheckboxVariant uint8

const (
	CheckboxPrimary CheckboxVariant = iota
	CheckboxSecondary
)

type AlertStatus uint8

const (
	AlertDefault AlertStatus = iota
	AlertAccent
	AlertSuccess
	AlertWarning
	AlertDanger
)

type AvatarColor uint8

const (
	AvatarColorDefault AvatarColor = iota
	AvatarColorAccent
	AvatarColorSuccess
	AvatarColorWarning
	AvatarColorDanger
)

type AvatarVariant uint8

const (
	AvatarVariantDefault AvatarVariant = iota
	AvatarVariantSoft
)

type AvatarSize uint8

const (
	AvatarSizeMedium AvatarSize = iota
	AvatarSizeSmall
	AvatarSizeLarge
)

type BadgeColor uint8

const (
	BadgeColorDefault BadgeColor = iota
	BadgeColorAccent
	BadgeColorSuccess
	BadgeColorWarning
	BadgeColorDanger
)

type BadgeVariant uint8

const (
	BadgeVariantPrimary BadgeVariant = iota
	BadgeVariantSecondary
	BadgeVariantSoft
)

type BadgeSize uint8

const (
	BadgeSizeMedium BadgeSize = iota
	BadgeSizeSmall
	BadgeSizeLarge
)

// SwitchSize identifies the control size exposed to visual renderers.
type SwitchSize uint8

const (
	SwitchMedium SwitchSize = iota
	SwitchSmall
	SwitchLarge
)

// RadioGroupVariant identifies the semantic visual treatment of a radio group.
type RadioGroupVariant uint8

const (
	RadioPrimary RadioGroupVariant = iota
	RadioSecondary
)

// ToggleButtonVariant identifies a toggle button's visual treatment.
type ToggleButtonVariant uint8

const (
	ToggleButtonDefault ToggleButtonVariant = iota
	ToggleButtonGhost
)

// ToggleButtonSize identifies a toggle button's control size.
type ToggleButtonSize uint8

const (
	ToggleButtonMedium ToggleButtonSize = iota
	ToggleButtonSmall
	ToggleButtonLarge
)

// SliderOrientation identifies a slider's axis.
type SliderOrientation uint8

const (
	SliderHorizontal SliderOrientation = iota
	SliderVertical
)

// ProgressBarColor identifies a progress bar fill treatment.
type ProgressBarColor uint8

const (
	ProgressBarAccent ProgressBarColor = iota
	ProgressBarDefault
	ProgressBarSuccess
	ProgressBarWarning
	ProgressBarDanger
)

// ProgressBarSize identifies a progress bar height treatment.
type ProgressBarSize uint8

const (
	ProgressBarMedium ProgressBarSize = iota
	ProgressBarSmall
	ProgressBarLarge
)

// SpinnerColor identifies a spinner's semantic color.
type SpinnerColor uint8

const (
	SpinnerAccent SpinnerColor = iota
	SpinnerCurrent
	SpinnerSuccess
	SpinnerWarning
	SpinnerDanger
)

// SpinnerSize identifies a spinner's diameter preset.
type SpinnerSize uint8

const (
	SpinnerMedium SpinnerSize = iota
	SpinnerSmall
	SpinnerLarge
	SpinnerExtraLarge
)

// SurfaceVariant identifies a surface's visual prominence.
type SurfaceVariant uint8

const (
	SurfaceDefault SurfaceVariant = iota
	SurfaceSecondary
	SurfaceTertiary
	SurfaceTransparent
)

// CardVariant identifies a card's visual prominence.
type CardVariant uint8

const (
	CardDefault CardVariant = iota
	CardSecondary
	CardTertiary
	CardTransparent
)

// ChipColor identifies a chip's semantic color.
type ChipColor uint8

const (
	ChipColorDefault ChipColor = iota
	ChipColorAccent
	ChipColorSuccess
	ChipColorWarning
	ChipColorDanger
)

// ChipVariant identifies a chip's visual prominence.
type ChipVariant uint8

const (
	ChipVariantSecondary ChipVariant = iota
	ChipVariantPrimary
	ChipVariantTertiary
	ChipVariantSoft
)

// ChipSize identifies a chip's height and padding preset.
type ChipSize uint8

const (
	ChipSizeMedium ChipSize = iota
	ChipSizeSmall
	ChipSizeLarge
)

// ToastVariant identifies a toast's semantic visual treatment.
type ToastVariant uint8

const (
	ToastDefault ToastVariant = iota
	ToastAccent
	ToastSuccess
	ToastWarning
	ToastDanger
)
