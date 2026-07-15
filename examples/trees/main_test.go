package main

import (
	"testing"

	"github.com/qianniancn/FlowUI/ui"
)

func TestMoveTreeItemAcrossLevels(t *testing.T) {
	items := []ui.TreeItem{
		{Key: "folder", Children: []ui.TreeItem{{Key: "one"}, {Key: "two"}}},
		{Key: "other"},
	}
	if !moveTreeItem(&items, ui.TreeDropEvent{SourceKey: "one", TargetKey: "other", Position: ui.TreeDropBefore}) {
		t.Fatal("valid move was rejected")
	}
	if len(items) != 3 || items[0].Key != "folder" || items[1].Key != "one" || items[2].Key != "other" || len(items[0].Children) != 1 {
		t.Fatalf("moved items = %#v", items)
	}
	if moveTreeItem(&items, ui.TreeDropEvent{SourceKey: "folder", TargetKey: "two", Position: ui.TreeDropInside}) {
		t.Fatal("moving a parent into its descendant was accepted")
	}
	if _, ok := findTreeItem(items, "folder"); !ok {
		t.Fatal("invalid move removed the source item")
	}
}
