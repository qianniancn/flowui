package state

import (
	"fmt"
	"strings"
)

type Kind string

const (
	KindClickable        Kind = "clickable"
	KindDraggable        Kind = "draggable"
	KindEditor           Kind = "editor"
	KindInput            Kind = "input"
	KindComboBox         Kind = "combobox"
	KindSelect           Kind = "select"
	KindDatePicker       Kind = "datepicker"
	KindCheckbox         Kind = "checkbox"
	KindSwitch           Kind = "switch"
	KindToggleButton     Kind = "toggle-button"
	KindRadioGroup       Kind = "radio-group"
	KindProgressBar      Kind = "progress-bar"
	KindProgressCircle   Kind = "progress-circle"
	KindMeter            Kind = "meter"
	KindSlider           Kind = "slider"
	KindListBox          Kind = "listbox"
	KindTree             Kind = "tree"
	KindSidebar          Kind = "sidebar"
	KindTable            Kind = "table"
	KindPagination       Kind = "pagination"
	KindMenu             Kind = "menu"
	KindContextMenu      Kind = "context-menu"
	KindDropdown         Kind = "dropdown"
	KindMenubar          Kind = "menubar"
	KindLineChart        Kind = "line-chart"
	KindBarChart         Kind = "bar-chart"
	KindPieChart         Kind = "pie-chart"
	KindCandlestickChart Kind = "candlestick-chart"
	KindTween            Kind = "tween"
	KindTabs             Kind = "tabs"
	KindPopover          Kind = "popover"
	KindTooltip          Kind = "tooltip"
	KindToast            Kind = "toast"
	KindModal            Kind = "modal"
	KindList             Kind = "list"
	KindScroll           Kind = "scroll"
	KindScrollbar        Kind = "scrollbar"
	KindSplitPane        Kind = "split-pane"
)

type Keys struct {
	frame map[string]Kind
	path  []string
}

// Scope returns a copy of the current key path. It is used by deferred frame
// work that must keep the identity scope from its registration site.
func (k *Keys) Scope() []string {
	return append([]string(nil), k.path...)
}

// UseScope temporarily replaces the current key path.
func (k *Keys) UseScope(path []string) func() {
	previous := k.path
	k.path = append(k.path[:0:0], path...)
	return func() {
		k.path = previous
	}
}

const (
	encodedRootPrefix = "~r"
	derivedKeyPrefix  = byte(0)
)

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
	return k.claim(kind, k.FullKey(key))
}

func (k *Keys) ClaimDerived(kind Kind, owner, role string) string {
	if owner == "" {
		panic("flowui: empty derived owner key")
	}
	return k.ClaimDerivedResolved(kind, k.FullKey(owner), role)
}

func (k *Keys) ClaimDerivedResolved(kind Kind, owner, role string) string {
	return k.claim(kind, k.Derived(owner, role))
}

// Derived returns an internal identity for a role owned by an already resolved
// key. User keys escape derivedKeyPrefix and cannot enter this namespace.
func (k *Keys) Derived(owner, role string) string {
	if owner == "" {
		panic("flowui: empty derived owner key")
	}
	if role == "" {
		panic("flowui: empty derived role")
	}

	var key strings.Builder
	key.WriteByte(derivedKeyPrefix)
	writeKeyPart(&key, owner)
	key.WriteByte('/')
	writeKeyPart(&key, role)
	return key.String()
}

func (k *Keys) claim(kind Kind, key string) string {
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
		// Scoped keys always contain '/', so ambiguous roots use a slash-free namespace.
		if strings.Contains(key, "/") || strings.IndexByte(key, derivedKeyPrefix) >= 0 || strings.HasPrefix(key, encodedRootPrefix) {
			var b strings.Builder
			b.WriteString(encodedRootPrefix)
			writeKeyPart(&b, key)
			return b.String()
		}
		return key
	}
	var b strings.Builder
	for _, part := range k.path {
		writeKeyPart(&b, part)
		b.WriteByte('/')
	}
	writeKeyPart(&b, key)
	return b.String()
}

func writeKeyPart(b *strings.Builder, part string) {
	for _, char := range []byte(part) {
		switch char {
		case derivedKeyPrefix:
			b.WriteString("%00")
		case '%':
			b.WriteString("%25")
		case '/':
			b.WriteString("%2F")
		default:
			b.WriteByte(char)
		}
	}
}

// Sweep removes keyed state that was not claimed with kind in the current frame.
func Sweep[T any](states map[string]*T, frame map[string]Kind, kind Kind) {
	for key := range states {
		if frame[key] != kind {
			delete(states, key)
		}
	}
}
