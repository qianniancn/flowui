package radiogroup

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type RadioItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
	Invalid     bool
}

type RadioGroupWidget struct {
	theme       func(*theme.Theme)
	key         string
	selectedKey string
	items       []RadioItem
	onChange    func(string)
	variant     RadioGroupVariant
	disabled    bool
	invalid     bool
	horizontal  bool
}

type RadioGroupVariant int

const (
	RadioPrimary RadioGroupVariant = iota
	RadioSecondary
)

const (
	radioSelectDuration   = 200 * time.Millisecond
	radioFocusDuration    = 100 * time.Millisecond
	radioPressInDuration  = 80 * time.Millisecond
	radioPressOutDuration = 140 * time.Millisecond
)

func RadioGroup(key, selectedKey string, items []RadioItem) RadioGroupWidget {
	return RadioGroupWidget{
		key:         key,
		selectedKey: selectedKey,
		items:       items,
	}
}

func (r RadioGroupWidget) OnChange(fn func(string)) RadioGroupWidget {
	r.onChange = fn
	return r
}

func (r RadioGroupWidget) Disabled(disabled bool) RadioGroupWidget {
	r.disabled = disabled
	return r
}

func (r RadioGroupWidget) Invalid(invalid bool) RadioGroupWidget {
	r.invalid = invalid
	return r
}

func (r RadioGroupWidget) Variant(variant RadioGroupVariant) RadioGroupWidget {
	r.variant = variant
	return r
}

func (r RadioGroupWidget) Horizontal() RadioGroupWidget {
	r.horizontal = true
	return r
}

func (r RadioGroupWidget) Theme(fn func(*theme.Theme)) RadioGroupWidget {
	r.theme = fn
	return r
}

func (r RadioGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, r.theme); restore != nil {
		defer restore()
	}
	state := radioGroupStateFor(ctx, r.key)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(r.items)

	if r.disabled {
		gtx = gtx.Disabled()
	} else if key, ok := state.updateKeys(gtx, r.items, r.selectedKey); ok {
		if key != r.selectedKey && r.onChange != nil {
			r.onChange(key)
		}
		frame.RequestFocus(ctx, &state.item(key).clickable)
	}

	children := make([]layout.Widget, 0, len(r.items))
	for _, item := range r.items {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return r.layoutItem(ctx, gtx, state, item)
		})
	}

	theme := frame.ActiveTheme(ctx).Components.RadioGroup
	return layoutui.LayoutItems(ctx, gtx, r.horizontal, gtx.Dp(theme.HorizontalGap), gtx.Dp(theme.VerticalGap), children)
}
