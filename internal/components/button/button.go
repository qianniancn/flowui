package button

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/interaction"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type ButtonWidget struct {
	key         string
	child       frame.Widget
	label       string
	onClick     func()
	variant     ButtonVariant
	variantSet  bool
	size        ButtonSize
	sizeSet     bool
	disabled    bool
	disabledSet bool
	loading     bool
	fullWidth   bool
	iconOnly    bool
	group       buttonGroupItemStyle
	prepared    buttonPreparedContent
	preparedSet bool
	customStyle flowstyle.Style
}

type ButtonVariant int

const (
	ButtonPrimary ButtonVariant = iota
	ButtonSecondary
	ButtonTertiary
	ButtonGhost
	ButtonOutline
	ButtonDanger
	ButtonDangerSoft
)

type ButtonSize int

const (
	ButtonMedium ButtonSize = iota
	ButtonSmall
	ButtonLarge
)

const (
	buttonPressInDuration  = 90 * time.Millisecond
	buttonPressOutDuration = 160 * time.Millisecond
	buttonColorDuration    = 100 * time.Millisecond
	buttonFocusDuration    = 100 * time.Millisecond
	buttonSpinnerPeriod    = 900 * time.Millisecond
)

func Button(key string, child frame.Widget) ButtonWidget {
	return ButtonWidget{key: key, child: child}
}

func (b ButtonWidget) OnClick(fn func()) ButtonWidget {
	b.onClick = fn
	return b
}

func (b ButtonWidget) Label(value string) ButtonWidget {
	b.label = value
	return b
}

func (b ButtonWidget) Disabled(disabled bool) ButtonWidget {
	b.disabled = disabled
	b.disabledSet = true
	return b
}

func (b ButtonWidget) Loading(loading bool) ButtonWidget {
	b.loading = loading
	return b
}

func (b ButtonWidget) Variant(variant ButtonVariant) ButtonWidget {
	b.variant = variant
	b.variantSet = true
	return b
}

func (b ButtonWidget) Size(size ButtonSize) ButtonWidget {
	b.size = size
	b.sizeSet = true
	return b
}

func (b ButtonWidget) FullWidth() ButtonWidget {
	b.fullWidth = true
	return b
}

func (b ButtonWidget) IconOnly() ButtonWidget {
	b.iconOnly = true
	return b
}

// Style applies an instance style after the Button defaults, variant, size,
// and inherited StyleScope declarations.
func (b ButtonWidget) Style(value flowstyle.Style) ButtonWidget {
	b.customStyle = value
	return b
}

func (b ButtonWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, nil, true)
}

// LayoutWithClickable renders a button with caller-owned interaction state.
// It is intended for internal component composition.
func LayoutWithClickable(b ButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, true)
}

// LayoutWithClickableNoEvents renders a button with caller-owned event handling.
func LayoutWithClickableNoEvents(b ButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, false)
}

func layoutWithClickable(b ButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable, handleEvents bool) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	animGtx := gtx
	click := interaction.BeginClick(ctx, gtx, b.key, clickable, !b.disabled && !b.loading, handleEvents, b.onClick)
	styleState := click.StyleState
	styleState.Disabled = b.disabled || !gtx.Enabled()
	styleState.Loading = b.loading
	resolvedStyle := b.resolveStyle(ctx, animGtx, click.Key, styleState)

	if b.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if b.loading && !b.iconOnly && !b.fullWidth {
		resolvedStyle = buttonLoadingStyle(gtx, activeTheme, b.size, resolvedStyle)
	}

	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		motion := activeTheme.Motion
		if b.loading && theme.ResolveMotionDuration(motion, buttonSpinnerPeriod) > 0 {
			animGtx.Execute(op.InvalidateCmd{})
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if b.preparedSet {
				b.prepared.call.Add(gtx.Ops)
				return b.prepared.dims
			}
			return b.layoutContent(ctx, gtx, resolvedStyle, activeTheme)
		})
	})

	return layoutui.LayoutInteractiveResolved(ctx, gtx, resolvedStyle.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return click.Layout(gtx, visual, b.label)
	})
}
