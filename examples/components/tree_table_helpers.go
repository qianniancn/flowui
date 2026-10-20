package main

import (
	"slices"
	"sort"
	"strings"

	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/tree"
)

func catalogTreeCanDrop(items []tree.Item, event tree.DropEvent) bool {
	if event.TargetKey == "" || len(event.SourceKeys) == 0 {
		return false
	}
	target, ok := catalogTreeItem(items, event.TargetKey)
	if !ok || target.Disabled {
		return false
	}
	if event.Position == tree.DropInside && !catalogTreeAcceptsChildren(*target) {
		return false
	}
	for _, sourceKey := range event.SourceKeys {
		if sourceKey == "" || sourceKey == event.TargetKey || catalogTreeIsAncestor(items, sourceKey, event.TargetKey) {
			return false
		}
		source, exists := catalogTreeItem(items, sourceKey)
		if !exists || source.Disabled {
			return false
		}
	}
	return true
}

func catalogTreeMove(items *[]tree.Item, event tree.DropEvent) bool {
	if items == nil || !catalogTreeCanDrop(*items, event) {
		return false
	}
	working := append([]tree.Item(nil), (*items)...)
	var moved tree.Item
	var removed bool
	for _, sourceKey := range event.SourceKeys {
		var item tree.Item
		working, item, removed = catalogTreeRemove(working, sourceKey)
		if !removed {
			return false
		}
		if moved.Key == "" {
			moved = item
		}
	}
	working, inserted := catalogTreeInsert(working, moved, event.TargetKey, event.Position)
	if !inserted {
		return false
	}
	*items = working
	return true
}

func catalogTreeRemove(items []tree.Item, key string) ([]tree.Item, tree.Item, bool) {
	result := make([]tree.Item, 0, len(items))
	var removed tree.Item
	found := false
	for index := range items {
		item := items[index]
		if item.Key == key {
			removed = item
			found = true
			continue
		}
		if !found {
			children, child, ok := catalogTreeRemove(item.Children, key)
			if ok {
				item.Children = children
				removed = child
				found = true
			}
		}
		result = append(result, item)
	}
	return result, removed, found
}

func catalogTreeInsert(items []tree.Item, moved tree.Item, targetKey string, position tree.DropPosition) ([]tree.Item, bool) {
	result := make([]tree.Item, 0, len(items)+1)
	for index := range items {
		item := items[index]
		if item.Key == targetKey {
			switch position {
			case tree.DropBefore:
				result = append(result, moved, item)
			case tree.DropAfter:
				result = append(result, item, moved)
			case tree.DropInside:
				item.Children = append(append([]tree.Item(nil), item.Children...), moved)
				result = append(result, item)
			default:
				return items, false
			}
			result = append(result, items[index+1:]...)
			return result, true
		}
		children, ok := catalogTreeInsert(item.Children, moved, targetKey, position)
		if ok {
			item.Children = children
			result = append(result, item)
			result = append(result, items[index+1:]...)
			return result, true
		}
		result = append(result, item)
	}
	return result, false
}

func catalogTreeItem(items []tree.Item, key string) (*tree.Item, bool) {
	for index := range items {
		if items[index].Key == key {
			return &items[index], true
		}
		if item, ok := catalogTreeItem(items[index].Children, key); ok {
			return item, true
		}
	}
	return nil, false
}

func catalogTreeIsAncestor(items []tree.Item, ancestorKey, targetKey string) bool {
	item, ok := catalogTreeItem(items, ancestorKey)
	return ok && catalogTreeHasKey(item.Children, targetKey)
}

func catalogTreeHasKey(items []tree.Item, key string) bool {
	for _, item := range items {
		if item.Key == key || catalogTreeHasKey(item.Children, key) {
			return true
		}
	}
	return false
}

func catalogTreeAcceptsChildren(item tree.Item) bool {
	return item.AcceptsChildren || len(item.Children) > 0
}

func treeDropPositionLabel(position tree.DropPosition) string {
	switch position {
	case tree.DropBefore:
		return "before"
	case tree.DropInside:
		return "inside"
	case tree.DropAfter:
		return "after"
	default:
		return "near"
	}
}

func containsCatalogTreeKey(keys []string, key string) bool {
	return slices.Contains(keys, key)
}

func sortedCatalogTableRows(descriptor table.SortDescriptor) []table.Row {
	rows := append([]table.Row(nil), catalogTableRows...)
	if descriptor.Column == "" {
		return rows
	}
	sort.SliceStable(rows, func(left, right int) bool {
		first := catalogTableSortValue(rows[left], descriptor.Column)
		second := catalogTableSortValue(rows[right], descriptor.Column)
		comparison := strings.Compare(first, second)
		if descriptor.Direction == table.SortDescending {
			return comparison > 0
		}
		return comparison < 0
	})
	return rows
}

func catalogTableSortValue(row table.Row, column string) string {
	switch column {
	case "name", "category", "status":
		index := 0
		if column == "category" {
			index = 1
		} else if column == "status" {
			index = 2
		}
		if index < len(row.Cells) {
			if row.Cells[index].Text != "" {
				return row.Cells[index].Text
			}
			return row.Key
		}
	}
	return row.Key
}
