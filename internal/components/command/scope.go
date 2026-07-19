package command

import (
	"fmt"
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// ScopeWidget installs shortcuts for its child subtree.
type ScopeWidget struct {
	commands                []Command
	child                   frame.Widget
	disableWhenFieldFocused bool
}

func Scope(commands []Command, child frame.Widget) ScopeWidget {
	return ScopeWidget{commands: append([]Command(nil), commands...), child: child}
}

// DisableWhenFieldFocused prevents application commands from taking over
// editing shortcuts while an input field has keyboard focus.
func (s ScopeWidget) DisableWhenFieldFocused() ScopeWidget {
	s.disableWhenFieldFocused = true
	return s
}

func (s ScopeWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	s.validate()
	dims := layout.Dimensions{Size: gtx.Constraints.Constrain(image.Point{})}
	if s.child != nil {
		dims = s.child.Layout(ctx, gtx)
	}
	if gtx.Enabled() && !(s.disableWhenFieldFocused && frame.AnyFieldFocused(ctx, gtx)) {
		s.update(gtx)
	}
	return dims
}

func (s ScopeWidget) validate() {
	for index, command := range s.commands {
		command.validate()
		for _, previous := range s.commands[:index] {
			if previous.key == command.key {
				panic(fmt.Sprintf("flowui: duplicate command key %q", command.key))
			}
			if command.shortcut.empty() || previous.shortcut.empty() {
				continue
			}
			if command.shortcut.name == previous.shortcut.name && command.shortcut.modifiers == previous.shortcut.modifiers {
				panic(fmt.Sprintf("flowui: shortcut %q is used by commands %q and %q", command.shortcut, previous.key, command.key))
			}
		}
	}
}

func (s ScopeWidget) update(gtx layout.Context) {
	filters := make([]event.Filter, 0, len(s.commands))
	for _, command := range s.commands {
		if command.disabled || command.shortcut.empty() {
			continue
		}
		filters = append(filters, command.shortcut.filter())
	}
	if len(filters) == 0 {
		return
	}
	for {
		value, ok := gtx.Event(filters...)
		if !ok {
			return
		}
		e, ok := value.(key.Event)
		if !ok || e.State != key.Press {
			continue
		}
		for _, command := range s.commands {
			if command.shortcut.matches(e) {
				command.execute()
				break
			}
		}
	}
}
