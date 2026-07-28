package main

import (
	"testing"

	"github.com/qianniancn/flowui/ui"
)

func TestMoveTreeItemAcrossLevels(t *testing.T) {
	items := []ui.TreeItem{
		{Key: "folder", Children: []ui.TreeItem{{Key: "one"}, {Key: "two"}}},
		{Key: "other"},
	}
	if !moveTreeItems(&items, ui.TreeDropEvent{SourceKey: "one", TargetKey: "other", Position: ui.TreeDropBefore}) {
		t.Fatal("valid move was rejected")
	}
	if len(items) != 3 || items[0].Key != "folder" || items[1].Key != "one" || items[2].Key != "other" || len(items[0].Children) != 1 {
		t.Fatalf("moved items = %#v", items)
	}
	if moveTreeItems(&items, ui.TreeDropEvent{SourceKey: "folder", TargetKey: "two", Position: ui.TreeDropInside}) {
		t.Fatal("moving a parent into its descendant was accepted")
	}
	if _, ok := findTreeItem(items, "folder"); !ok {
		t.Fatal("invalid move removed the source item")
	}
}

func TestMoveTreeItemsPreservesBatchOrder(t *testing.T) {
	items := []ui.TreeItem{{Key: "one"}, {Key: "two"}, {Key: "target"}}
	event := ui.TreeDropEvent{
		SourceKey: "one", SourceKeys: []string{"one", "two"},
		TargetKey: "target", Position: ui.TreeDropAfter,
	}
	if !moveTreeItems(&items, event) {
		t.Fatal("valid batch move was rejected")
	}
	if len(items) != 3 || items[0].Key != "target" || items[1].Key != "one" || items[2].Key != "two" {
		t.Fatalf("batch order = %#v", items)
	}
}

func TestTreeExampleAsyncAndRenameUpdates(t *testing.T) {
	async := asyncItems()
	if !setTreeChildrenState(async, "remote", ui.TreeChildrenLoading, "") || async[0].ChildrenState != ui.TreeChildrenLoading {
		t.Fatal("async loading state was not applied")
	}
	if !setTreeChildren(async, "remote", loadedTreeChildren("remote")) || async[0].ChildrenState != ui.TreeChildrenLoaded || len(async[0].Children) != 2 {
		t.Fatalf("async children = %#v", async[0])
	}
	items := fileItems()
	if !renameTreeItem(items, "tree", "outline.go") {
		t.Fatal("rename target was not found")
	}
	item, ok := findTreeItem(items, "tree")
	if !ok || item.Label != "outline.go" {
		t.Fatalf("renamed item = %#v", item)
	}
}

func TestTreeExampleRenameRequestAdvancesRevision(t *testing.T) {
	model := Model{}
	Update(&model, Msg{RenameTree: "compact", RenameRequest: "tree"})
	if model.RenameTree != "compact" || model.RenameTarget != "tree" || model.RenameRevision != 1 {
		t.Fatalf("rename request = tree %q target %q revision %d", model.RenameTree, model.RenameTarget, model.RenameRevision)
	}
	Update(&model, Msg{RenameTree: "compact", RenameRequest: "tree"})
	if model.RenameRevision != 2 {
		t.Fatalf("repeat rename revision = %d, want 2", model.RenameRevision)
	}
}
