package state

import (
	"fmt"
	"strings"
)

type Kind string

const (
	KindClickable   Kind = "clickable"
	KindEditor      Kind = "editor"
	KindInput       Kind = "input"
	KindComboBox    Kind = "combobox"
	KindDatePicker  Kind = "datepicker"
	KindCheckbox    Kind = "checkbox"
	KindRadioGroup  Kind = "radio-group"
	KindProgressBar Kind = "progress-bar"
	KindList        Kind = "list"
	KindScroll      Kind = "scroll"
)

type Keys struct {
	frame map[string]Kind
	path  []string
}

func (k *Keys) BeginFrame() {
	if k.frame == nil {
		k.frame = make(map[string]Kind)
		return
	}
	clear(k.frame)
}

func (k *Keys) Frame() map[string]Kind {
	return k.frame
}

func (k *Keys) Push(key string) func() {
	if key == "" {
		panic("flowui: empty key")
	}
	k.path = append(k.path, key)
	return func() {
		k.path = k.path[:len(k.path)-1]
	}
}

func (k *Keys) Claim(kind Kind, key string) string {
	if key == "" {
		panic(fmt.Sprintf("flowui: empty %s key", kind))
	}
	key = k.FullKey(key)
	if k.frame == nil {
		k.BeginFrame()
	}
	if _, ok := k.frame[key]; ok {
		panic(fmt.Sprintf("flowui: duplicate key %q", key))
	}
	k.frame[key] = kind
	return key
}

func (k *Keys) FullKey(key string) string {
	if len(k.path) == 0 {
		return key
	}
	var b strings.Builder
	for _, part := range k.path {
		b.WriteString(part)
		b.WriteByte('/')
	}
	b.WriteString(key)
	return b.String()
}

func Sweep[T any](states map[string]*T, frame map[string]Kind, kind Kind) {
	for key := range states {
		if frame[key] != kind {
			delete(states, key)
		}
	}
}
