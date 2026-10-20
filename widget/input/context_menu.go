package input

import (
	"io"
	"runtime"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/menu"
)

// primaryShortcutLabel renders the platform primary modifier for display-only
// shortcut hints. It mirrors widget/command's labels without importing it,
// because command's test suite imports this package.
func primaryShortcutLabel(keyName string) string {
	if runtime.GOOS == "darwin" {
		return "Command+" + keyName
	}
	return "Ctrl+" + keyName
}

// EditAction identifies a queued editor operation triggered by the built-in
// context menu. Menu activation runs in the popup overlay pass, so actions are
// queued on the field's state and executed by the field's next Layout.
type EditAction uint8

const (
	editNone EditAction = iota
	EditCut
	EditCopy
	EditPaste
	EditSelectAll
)

// EditSnapshot captures the editor state that decides which default menu
// entries are enabled.
type EditSnapshot struct {
	Selection int
	Length    int
	ReadOnly  bool
	Secret    bool
}

func SnapshotEditor(editor *widget.Editor, readOnly, secret bool) EditSnapshot {
	start, end := editor.Selection()
	selection := end - start
	if selection < 0 {
		selection = -selection
	}
	return EditSnapshot{Selection: selection, Length: editor.Len(), ReadOnly: readOnly, Secret: secret}
}

// editItemSpec is the testable shape of one default edit menu entry.
type editItemSpec struct {
	key      string
	label    string
	shortcut string
	disabled bool
	action   EditAction
}

// editMenuLabels holds the built-in edit menu labels for one language.
type editMenuLabels struct {
	cut       string
	copy      string
	paste     string
	selectAll string
}

func editMenuLabelsFor(ctx *core.Context) editMenuLabels {
	return editMenuLabels{
		cut:       ctx.Text(i18n.ActionCut),
		copy:      ctx.Text(i18n.ActionCopy),
		paste:     ctx.Text(i18n.ActionPaste),
		selectAll: ctx.Text(i18n.ActionSelectAll),
	}
}

func defaultEditItemSpecs(snapshot EditSnapshot, labels editMenuLabels) []editItemSpec {
	return []editItemSpec{
		{
			key: "cut", label: labels.cut,
			shortcut: primaryShortcutLabel("X"),
			disabled: snapshot.ReadOnly || snapshot.Secret || snapshot.Selection == 0,
			action:   EditCut,
		},
		{
			key: "copy", label: labels.copy,
			shortcut: primaryShortcutLabel("C"),
			disabled: snapshot.Secret || snapshot.Selection == 0,
			action:   EditCopy,
		},
		{
			key: "paste", label: labels.paste,
			shortcut: primaryShortcutLabel("V"),
			disabled: snapshot.ReadOnly,
			action:   EditPaste,
		},
		{
			key: "select-all", label: labels.selectAll,
			shortcut: primaryShortcutLabel("A"),
			disabled: snapshot.Length == 0,
			action:   EditSelectAll,
		},
	}
}

// defaultEditMenu builds the built-in cut/copy/paste menu for a field. queue
// receives the activated action and must route it to the field's state.
func defaultEditMenu(queue func(EditAction), snapshot EditSnapshot, labels editMenuLabels) menu.Widget {
	specs := defaultEditItemSpecs(snapshot, labels)
	items := make([]menu.Item, 0, len(specs)+1)
	for index, spec := range specs {
		if index == 3 {
			items = append(items, menu.Separator())
		}
		spec := spec
		items = append(items, menu.Item{
			Key:      spec.key,
			Label:    spec.label,
			Shortcut: spec.shortcut,
			Disabled: spec.disabled,
			OnAction: func() {
				if queue != nil {
					queue(spec.action)
				}
			},
		})
	}
	return menu.New("edit", items)
}

// RunEditAction executes a queued edit against the field's editor. Cut
// mirrors Gio's own shortcut handling: write the selected text to the
// clipboard, then delete the selection.
func RunEditAction(gtx layout.Context, editor *widget.Editor, action EditAction, readOnly bool) {
	switch action {
	case EditCut, EditCopy:
		text := editor.SelectedText()
		if text == "" {
			return
		}
		gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
		if action == EditCut && !readOnly {
			editor.Delete(1)
		}
	case EditPaste:
		gtx.Execute(clipboard.ReadCmd{Tag: editor})
	case EditSelectAll:
		editor.SetCaret(0, editor.Len())
	}
}

// EditContextMenu wraps a field trigger with a right-click edit menu.
// custom replaces the default menu when present; the editor doubles as the
// focus target so focus returns to the field after the menu closes.
func EditContextMenu(
	ctx *core.Context,
	gtx layout.Context,
	ownerKey string,
	trigger core.Widget,
	editor *widget.Editor,
	snapshot EditSnapshot,
	custom menu.Widget,
	hasCustom bool,
	queue func(EditAction),
) layout.Dimensions {
	content := custom
	if !hasCustom {
		content = defaultEditMenu(queue, snapshot, editMenuLabelsFor(ctx))
	}
	key := core.DerivedKey(ctx, core.FullKey(ctx, ownerKey), "context-menu")
	// The menu steals focus from the editor; keep the selection painted while open.
	return menu.ContextMenu(key, trigger, content).
		FocusTarget(editor).
		OnOpenChange(func(open bool) { editor.PaintSelectionWhenUnfocused = open }).
		Layout(ctx, gtx)
}
