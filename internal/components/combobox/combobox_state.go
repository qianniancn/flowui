package combobox

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotComboBox = "combobox"

func comboBoxStateFor(ctx *frame.Context, key string) *comboBoxState {
	key = frame.ClaimKey(ctx, state.KindComboBox, key)
	return frame.UseState[comboBoxState](ctx, key, stateSlotComboBox)
}

type comboBoxState struct {
	editor             widget.Editor
	input              field.State
	trigger            widget.Clickable
	list               layout.List
	open               bool
	wasFocused         bool
	highlight          int
	syncedSelectedKey  string
	syncedInputValue   string
	ignoreEditorChange bool
	popover            float32
	popoverFrom        float32
	popoverTo          float32
	popoverAt          time.Time
	popoverReady       bool
	icon               float32
	iconFrom           float32
	iconTo             float32
	iconAt             time.Time
	iconReady          bool
	items              map[string]*comboBoxItemState
	frameItems         map[string]struct{}
	itemKeys           map[string]struct{}
}

func (s *comboBoxState) beginFrame() {
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
}

func (s *comboBoxState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
}

func (s *comboBoxState) item(key string) *comboBoxItemState {
	if s.items == nil {
		s.items = make(map[string]*comboBoxItemState)
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(comboBoxItemState)
	s.items[key] = item
	return item
}

func (s *comboBoxState) checkItems(items []ComboBoxItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty combobox item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate combobox item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
	}
}

func (s *comboBoxState) syncEditor(editor *widget.Editor, c ComboBoxWidget) {
	if c.hasInputValue {
		if c.inputValue != s.syncedInputValue {
			s.setEditorText(editor, c.inputValue)
			s.syncedInputValue = c.inputValue
		}
		return
	}
	if label, ok := c.selectedLabel(); ok {
		if c.selectedKey != s.syncedSelectedKey || label != s.syncedInputValue {
			s.setEditorText(editor, label)
			s.syncedInputValue = label
		}
	} else if !c.allowCustomValue {
		if c.selectedKey != s.syncedSelectedKey || s.syncedInputValue != "" {
			s.setEditorText(editor, "")
			s.syncedInputValue = ""
		}
	}
	s.syncedSelectedKey = c.selectedKey
}

func (s *comboBoxState) setEditorText(editor *widget.Editor, text string) {
	if editor.Text() == text {
		return
	}
	editor.SetText(text)
	s.ignoreEditorChange = true
}

func (s *comboBoxState) consumeEditorChange() bool {
	if !s.ignoreEditorChange {
		return false
	}
	s.ignoreEditorChange = false
	return true
}

func (s *comboBoxState) updateFocus(focused, disabled bool) {
	if disabled {
		s.open = false
		s.wasFocused = false
		return
	}
	if focused && !s.wasFocused {
		s.open = true
	}
	if !focused {
		s.open = false
	}
	s.wasFocused = focused
}

func (s *comboBoxState) updateKeys(gtx layout.Context, editor *widget.Editor, items []ComboBoxItem, visible []int) (int, bool) {
	if !s.wasFocused {
		return 0, false
	}
	filters := []event.Filter{
		key.Filter{Focus: editor, Name: key.NameDownArrow},
		key.Filter{Focus: editor, Name: key.NameUpArrow},
	}
	if s.open {
		filters = append(filters,
			key.Filter{Focus: editor, Name: key.NameReturn},
			key.Filter{Focus: editor, Name: key.NameEnter},
			key.Filter{Focus: editor, Name: key.NameEscape},
		)
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return 0, false
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			if !s.open {
				s.open = true
				s.highlight = comboBoxFirstEnabled(items, visible)
				continue
			}
			s.highlight = comboBoxMoveHighlight(items, visible, s.highlight, 1)
		case key.NameUpArrow:
			if !s.open {
				s.open = true
				s.highlight = comboBoxLastEnabled(items, visible)
				continue
			}
			s.highlight = comboBoxMoveHighlight(items, visible, s.highlight, -1)
		case key.NameReturn, key.NameEnter:
			if s.highlight >= 0 && s.highlight < len(visible) && !items[visible[s.highlight]].Disabled {
				return s.highlight, true
			}
		case key.NameEscape:
			s.open = false
		}
	}
}

func comboBoxFirstEnabled(items []ComboBoxItem, visible []int) int {
	for i, index := range visible {
		if !items[index].Disabled {
			return i
		}
	}
	return -1
}

func comboBoxLastEnabled(items []ComboBoxItem, visible []int) int {
	for i := len(visible) - 1; i >= 0; i-- {
		if !items[visible[i]].Disabled {
			return i
		}
	}
	return -1
}

func comboBoxMoveHighlight(items []ComboBoxItem, visible []int, current, delta int) int {
	count := len(visible)
	if count == 0 {
		return -1
	}
	if current < 0 || current >= count || items[visible[current]].Disabled {
		if delta < 0 {
			return comboBoxLastEnabled(items, visible)
		}
		return comboBoxFirstEnabled(items, visible)
	}
	next := current
	for range count {
		next = (next + delta + count) % count
		if !items[visible[next]].Disabled {
			return next
		}
	}
	return current
}

func (s *comboBoxState) clampHighlight(items []ComboBoxItem, visible []int) {
	if len(visible) == 0 {
		s.highlight = -1
		return
	}
	if s.highlight >= 0 && s.highlight < len(visible) && !items[visible[s.highlight]].Disabled {
		return
	}
	s.highlight = comboBoxFirstEnabled(items, visible)
}

func (s *comboBoxState) popoverProgress(gtx layout.Context, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return comboBoxProgress(gtx, target, &s.popover, &s.popoverFrom, &s.popoverTo, &s.popoverAt, &s.popoverReady)
}

func (s *comboBoxState) iconProgress(gtx layout.Context, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return comboBoxProgress(gtx, target, &s.icon, &s.iconFrom, &s.iconTo, &s.iconAt, &s.iconReady)
}

func comboBoxProgress(gtx layout.Context, target float32, value, from, to *float32, at *time.Time, ready *bool) float32 {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(*at), comboBoxAnimationDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = render.Lerp(*from, *to, progress)
	return *value
}

func comboBoxItemScale(gtx layout.Context, history []widget.Press, disabled bool) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	press := history[len(history)-1]
	target := float32(0.98)
	if press.End.IsZero() {
		progress := render.Ease(render.Progress(gtx.Now.Sub(press.Start), comboBoxItemPressInDuration))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return render.Lerp(1, target, progress)
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(press.End), comboBoxItemPressDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return render.Lerp(target, 1, progress)
}

type comboBoxItemState struct {
	clickable widget.Clickable
	bg        color.NRGBA
	bgFrom    color.NRGBA
	bgTo      color.NRGBA
	bgAt      time.Time
	bgReady   bool
}

func (s *comboBoxItemState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return comboBoxItemColor(gtx, target, &s.bg, &s.bgFrom, &s.bgTo, &s.bgAt, &s.bgReady)
}

func comboBoxItemColor(gtx layout.Context, target color.NRGBA, value, from, to *color.NRGBA, at *time.Time, ready *bool) color.NRGBA {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(*at), comboBoxItemColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = render.LerpColor(*from, *to, progress)
	return *value
}

func comboBoxItemTransform(size image.Point, scale float32) op.TransformOp {
	origin := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	return op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale)))
}
