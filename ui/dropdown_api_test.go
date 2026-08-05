package ui_test

import (
	"testing"

	"github.com/qianniancn/flowui/ui"
)

func TestDropdownPublicItemKinds(t *testing.T) {
	var kind ui.DropdownItemKind = ui.DropdownItemCheckbox
	if kind != ui.MenuItemCheckbox {
		t.Fatalf("dropdown checkbox kind = %v, want menu checkbox kind", kind)
	}
	if item := ui.DropdownGroupLabel("More"); item.Kind != ui.DropdownItemGroupLabel || item.Label != "More" {
		t.Fatalf("dropdown group label = %#v", item)
	}
}
