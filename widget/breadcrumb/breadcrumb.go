// Package breadcrumb provides navigation through a hierarchy.
package breadcrumb

import (
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
)

// Item is one ancestor in the path. The final item is the current location.
type Item struct {
	Key, Label string
	Disabled   bool
}

type Widget struct {
	key       string
	items     []Item
	separator string
	onChange  func(string)
}

func New(key string, items []Item) Widget {
	return Widget{key: key, items: append([]Item{}, items...), separator: "/"}
}

func (b Widget) Separator(value string) Widget   { b.separator = value; return b }
func (b Widget) OnChange(fn func(string)) Widget { b.onChange = fn; return b }

func (b Widget) Layout(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
	owner := core.ClaimKey(ctx, core.KindCustom, b.key)
	children := make([]core.Widget, 0, len(b.items)*2)
	for index, item := range b.items {
		if index > 0 {
			children = append(children, text.New(b.separator))
		}
		if index == len(b.items)-1 {
			children = append(children, text.New(item.Label))
			continue
		}
		if item.Key == "" {
			panic("flowui: breadcrumb ancestor requires a key")
		}
		children = append(children, button.New(item.Label).Key(core.DerivedKey(ctx, owner, item.Key)).
			Variant(button.Ghost).Size(button.Small).Disabled(item.Disabled).
			OnClick(func() {
				if b.onChange != nil {
					b.onChange(item.Key)
				}
			}))
	}
	return layout.Wrap(children...).Gap(6).LineGap(4).AlignMiddle().Layout(ctx, gtx)
}

var _ core.Widget = Widget{}
