// Package testtheme contains explicit visual fixtures for widget behavior tests.
//
// Production code must not import this package. Behavior fixtures are
// intentionally independent of production theme recipes.
package testtheme

import (
	"image/color"
	"time"

	"github.com/qianniancn/flowui/theme"
)

// Behavior returns a neutral provider with only the geometry and state
// declarations needed by widget behavior tests.
func Behavior() *theme.Theme {
	value := theme.New(theme.DefaultSeed())
	value.Motion = theme.MotionTheme{
		Enabled: true, DefaultDuration: 100 * time.Millisecond,
		ControlDuration: 100 * time.Millisecond, DurationScale: 1,
	}
	value.VisualProvider = behaviorProvider{}
	return &value
}

// BehaviorProvider returns the complete neutral provider used by widget
// behavior tests. Callers can embed it and override only the capability under
// test without copying the rest of the visual protocol.
func BehaviorProvider() theme.ComponentVisualProvider {
	provider, err := theme.ComponentProvider(Behavior())
	if err != nil {
		panic(err)
	}
	return provider
}

type behaviorProvider struct{}

var _ theme.ComponentVisualProvider = behaviorProvider{}

var _ theme.ImmutableVisualProvider = behaviorProvider{}

var _ theme.CloseButtonProvider = behaviorProvider{}

var _ theme.ChipProvider = behaviorProvider{}

var _ theme.AvatarProvider = behaviorProvider{}

var _ theme.DescriptionProvider = behaviorProvider{}

var _ theme.LabelProvider = behaviorProvider{}

var _ theme.TextProvider = behaviorProvider{}

var _ theme.IconProvider = behaviorProvider{}

var _ theme.SurfaceProvider = behaviorProvider{}

var _ theme.CardProvider = behaviorProvider{}

var _ theme.StatusBarProvider = behaviorProvider{}

var _ theme.TitleBarProvider = behaviorProvider{}

var _ theme.ToolbarProvider = behaviorProvider{}

var _ theme.SwitchProvider = behaviorProvider{}

var _ theme.ToggleProvider = behaviorProvider{}

var _ theme.CheckboxProvider = behaviorProvider{}

var _ theme.RadioGroupProvider = behaviorProvider{}

var _ theme.SliderProvider = behaviorProvider{}

var _ theme.MenuProvider = behaviorProvider{}

var _ theme.MenubarProvider = behaviorProvider{}

var _ theme.DropdownProvider = behaviorProvider{}

var _ theme.PopoverProvider = behaviorProvider{}

var _ theme.TooltipProvider = behaviorProvider{}

var _ theme.ToastProvider = behaviorProvider{}

var _ theme.TabsProvider = behaviorProvider{}

var _ theme.ModalProvider = behaviorProvider{}

var _ theme.AlertProvider = behaviorProvider{}

var _ theme.AlertDialogProvider = behaviorProvider{}

var _ theme.TableProvider = behaviorProvider{}

var _ theme.LineChartProvider = behaviorProvider{}

var _ theme.BarChartProvider = behaviorProvider{}

var _ theme.GanttProvider = behaviorProvider{}

var _ theme.TimelineProvider = behaviorProvider{}

var _ theme.NodeGraphProvider = behaviorProvider{}

var _ theme.TreeProvider = behaviorProvider{}

var _ theme.ProgressProvider = behaviorProvider{}

var _ theme.SpinnerProvider = behaviorProvider{}

var _ theme.CollapsibleProvider = behaviorProvider{}

func (behaviorProvider) VisualsAreImmutable() {}

var (
	foreground = color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
	background = color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff}
	muted      = color.NRGBA{R: 0x70, G: 0x70, B: 0x70, A: 0xff}
	accent     = color.NRGBA{R: 0xb0, G: 0xb0, B: 0xb0, A: 0xff}
	accentFg   = color.NRGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xff}
	focus      = color.NRGBA{R: 0x30, G: 0x70, B: 0xc0, A: 0xff}
)
