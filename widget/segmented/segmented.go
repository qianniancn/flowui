// Package segmented provides compact mutually-exclusive choices.
package segmented

import (
	"github.com/qianniancn/flowui/theme"
	"image"

	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/togglebutton"
)

type Item struct {
	Key      string
	Label    string
	Disabled bool
}

type Widget struct {
	key       string
	selected  string
	items     []Item
	onChange  func(string)
	disabled  bool
	fullWidth bool
	override  *theme.SegmentedOverride
}

// New creates a segmented control. It starts uncontrolled with nothing selected;
// SelectedKey sets the controlled value.
func New(items []Item) Widget {
	return Widget{key: core.AutoKey(), items: append([]Item(nil), items...)}
}

// Key sets an explicit identity for the control state.
func (s Widget) Key(key string) Widget {
	s.key = key
	return s
}

func (s Widget) SelectedKey(key string) Widget {
	s.selected = key
	return s
}

func (s Widget) OnChange(fn func(string)) Widget {
	s.onChange = fn
	return s
}

func (s Widget) Disabled(disabled bool) Widget {
	s.disabled = disabled
	return s
}

func (s Widget) FullWidth() Widget {
	s.fullWidth = true
	return s
}

// Style merges instance declarations into the track.
func (s Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if s.override == nil {
		s.override = &theme.SegmentedOverride{}
	}
	s.override.Declaration = style.Join(s.override.Declaration, declaration)
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
	owner := core.ClaimKey(ctx, core.KindCustom, s.key)
	gtx.Constraints.Min = image.Point{}
	if s.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	children := make([]core.Widget, 0, len(s.items))
	for _, item := range s.items {
		selected := item.Key == s.selected
		if item.Key == "" {
			panic("flowui: segmented item requires a key")
		}
		choice := togglebutton.New(item.Label).Key(core.DerivedKey(ctx, owner, item.Key)).
			Selected(selected).Variant(togglebutton.Ghost).Disabled(s.disabled || item.Disabled)
		if !selected {
			choice = choice.OnChange(func(next bool) {
				if next && s.onChange != nil {
					s.onChange(item.Key)
				}
			})
		}
		var child core.Widget = choice
		if s.fullWidth {
			child = flowlayout.Expanded((choice).Style(style.New().FillWidth()))
		}
		children = append(children, child)
	}
	active := core.ActiveTheme(ctx)
	visual := theme.ResolveVisual(active, theme.SegmentedRequest{Override: s.override},
		theme.ComponentVisualProvider.Segmented, theme.ValidateSegmentedVisual)
	return flowlayout.Box(flowlayout.Row(children...).Gap(int(visual.Metrics.Gap))).
		Style(visual.Root, visual.Override).Layout(ctx, gtx)
}

var _ core.Widget = Widget{}
