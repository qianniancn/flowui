// Package commandpalette presents searchable application commands in a modal.
package commandpalette

import (
	"strings"

	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/listbox"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/modal"
	"github.com/qianniancn/flowui/widget/text"
)

// Widget uses controlled open state and internal search/highlight state.
// Commands are shared with command.Scope, buttons, and menus; opening the
// palette does not register a second set of application shortcuts.
type Widget struct {
	key, title, placeholder, emptyText string
	open                               bool
	commands                           []command.Command
	onOpenChange                       func(bool)
}

func New(key string, open bool, commands []command.Command) Widget {
	return Widget{key: key, open: open, commands: append([]command.Command{}, commands...)}
}

func (p Widget) Title(value string) Widget         { p.title = value; return p }
func (p Widget) Placeholder(value string) Widget   { p.placeholder = value; return p }
func (p Widget) EmptyText(value string) Widget     { p.emptyText = value; return p }
func (p Widget) OnOpenChange(fn func(bool)) Widget { p.onOpenChange = fn; return p }

type paletteState struct {
	query, selected   string
	open, focusSearch bool
}

func (p Widget) Layout(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
	owner := core.ClaimKey(ctx, core.KindCustom, p.key)
	state := core.UseState[paletteState](ctx, owner, "command-palette")
	if p.open && !state.open {
		state.query, state.selected, state.focusSearch = "", "", true
	}
	state.open = p.open
	searchKey := core.DerivedKey(ctx, owner, "search")
	dialogKey := core.DerivedKey(ctx, owner, "dialog")
	dialogFullKey := core.FullKey(ctx, dialogKey)
	body := core.WidgetFunc(func(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
		items := p.matches(ctx, state.query)
		state.ensureSelection(items)
		searchTag := core.FieldFocusTag(ctx, core.FullKey(ctx, searchKey))
		if searchTag != nil && p.open && gtx.Enabled() {
			for {
				event, ok := gtx.Event(
					key.Filter{Focus: searchTag, Name: key.NameUpArrow},
					key.Filter{Focus: searchTag, Name: key.NameDownArrow},
				)
				if !ok {
					break
				}
				if event, ok := event.(key.Event); ok && event.State == key.Press {
					delta := 1
					if event.Name == key.NameUpArrow {
						delta = -1
					}
					state.move(items, delta)
				}
			}
		}
		activate := func(selected string) {
			if !p.open || !gtx.Enabled() {
				return
			}
			for _, item := range p.matches(ctx, state.query) {
				if item.Key != selected || item.Disabled {
					continue
				}
				if !item.KeepOpen && p.onOpenChange != nil {
					p.onOpenChange(false)
				}
				item.OnAction()
				return
			}
		}
		placeholder := p.placeholder
		if placeholder == "" {
			placeholder = ctx.Text(i18n.CommandPaletteSearch)
		}
		field := input.New(state.query).Key(searchKey).Placeholder(placeholder).Label(placeholder).FullWidth().
			OnChange(func(value string) { state.query, state.selected = value, "" }).
			OnSubmit(func(value string) {
				state.query = value
				state.ensureSelection(p.matches(ctx, value))
				activate(state.selected)
			})
		list := core.WidgetFunc(func(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
			// The editor may update the query earlier in this layout pass.
			items := p.matches(ctx, state.query)
			state.ensureSelection(items)
			rows := make([]listbox.Item, 0, len(items))
			for _, item := range items {
				row := listbox.Item{
					Key: item.Key, Label: item.Label, Description: item.Description,
					Leading: item.Leading, Disabled: item.Disabled,
				}
				if item.Shortcut != "" {
					row.Trailing = text.New(item.Shortcut)
				}
				rows = append(rows, row)
			}
			empty := p.emptyText
			if empty == "" {
				empty = ctx.Text(i18n.CommandPaletteEmpty)
			}
			return listbox.New(rows).Key(core.DerivedKey(ctx, owner, "results")).SelectedKey(state.selected).
				FullWidth().MaxHeight(320).HideIndicator().EmptyText(empty).ScrollToKey(state.selected).
				OnChange(func(value string) { state.selected = value }).OnAction(activate).Layout(ctx, gtx)
		})
		dims := layout.Column(field, list).Gap(12).Layout(ctx, gtx)
		if state.focusSearch && p.open && gtx.Enabled() {
			fullKey := core.FullKey(ctx, searchKey)
			// Modal restores its own focus after layout; select the search field
			// after that restoration so the first keystroke reaches the editor.
			core.AfterOverlays(ctx, func() {
				if core.OverlayFocusScopeTopmost(ctx, core.OverlayLayerModal, dialogFullKey) {
					core.RequestFieldFocus(ctx, fullKey)
					state.focusSearch = false
				}
			})
		}
		return dims
	})
	title := p.title
	if title == "" {
		title = ctx.Text(i18n.CommandPaletteTitle)
	}
	return modal.New(title, body).Key(dialogKey).Open(p.open).
		Open(p.open).Placement(modal.Top).OnOpenChange(p.onOpenChange).Layout(ctx, gtx)
}

func (p Widget) matches(ctx *core.Context, query string) []menu.Item {
	words := strings.Fields(strings.ToLower(query))
	items := make([]menu.Item, 0, len(p.commands))
	keys := make(map[string]bool, len(p.commands))
	for _, cmd := range p.commands {
		item := command.MenuItemFor(ctx, cmd)
		if keys[item.Key] {
			panic("flowui: duplicate command palette key " + item.Key)
		}
		keys[item.Key] = true
		haystack := strings.ToLower(item.Label + " " + item.Description + " " + item.Key)
		matched := true
		for _, word := range words {
			if !strings.Contains(haystack, word) {
				matched = false
				break
			}
		}
		if matched {
			items = append(items, item)
		}
	}
	return items
}

func (s *paletteState) ensureSelection(items []menu.Item) {
	for _, item := range items {
		if item.Key == s.selected && !item.Disabled {
			return
		}
	}
	s.selected = ""
	for _, item := range items {
		if !item.Disabled {
			s.selected = item.Key
			return
		}
	}
}

func (s *paletteState) move(items []menu.Item, delta int) {
	current := -1
	for index, item := range items {
		if item.Key == s.selected {
			current = index
			break
		}
	}
	for step := 1; step <= len(items); step++ {
		index := (current + delta*step + len(items)*2) % len(items)
		if !items[index].Disabled {
			s.selected = items[index].Key
			return
		}
	}
}

var _ core.Widget = Widget{}
