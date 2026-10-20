package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/tree"
)

const virtualRowCount = 10000

func demoVirtualList(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		muted("10,000 条记录，滚动时按需构建可见行。"),
		fixedHeight(320, layout.List(virtualRowCount, func(index int) ui.Widget {
			id := fmt.Sprintf("record-%05d", index+1)
			return button.New(fmt.Sprintf("记录 %05d · 查看详情", index+1)).Key(id).Variant(button.Ghost).
				OnClick(func() {
					send(demoMsg(func(m *demoState) { m.virtualSelected = id }))
					send(logMsg("已选择：" + id))
				})
		}).Gap(4)),
		text.New("选中记录："+m.virtualSelected),
	).Gap(12)
}

func demoVirtualTable(m demoState, send ui.Send[Msg]) ui.Widget {
	columns := []table.Column{
		{Key: "id", Label: "编号", Width: 100, Sortable: true},
		{Key: "name", Label: "项目", Weight: 1},
		{Key: "status", Label: "状态", Width: 80},
	}
	return layout.Column(
		muted("10,000 行 · 点击编号列排序，点击行选择。"),
		table.NewVirtual("virtual-table", columns, virtualRowCount, func(index int) table.Row {
			if m.virtualSort.Column == "id" && m.virtualSort.Direction == table.SortDescending {
				index = virtualRowCount - index - 1
			}
			id := fmt.Sprintf("record-%05d", index+1)
			return table.Row{Key: id, Cells: []table.Cell{
				{Text: fmt.Sprintf("%05d", index+1)}, {Text: "项目 " + id}, {Text: "已同步"},
			}}
		}).MaxHeight(320).Border(true).Striped(true).
			SortDescriptor(m.virtualSort).OnSortChange(func(value table.SortDescriptor) {
			send(demoMsg(func(m *demoState) { m.virtualSort = value }))
		}).SelectionMode(table.SelectionSingle).SelectedKey(m.virtualSelected).
			OnChange(setString(send, func(m *demoState) *string { return &m.virtualSelected })),
		text.New("选中记录："+m.virtualSelected),
	).Gap(12)
}

func initialExplorer() []tree.SimpleItem {
	return []tree.SimpleItem{
		{Key: "src", Label: "src", AcceptsChildren: true, DragDisabled: true},
		{Key: "main", ParentKey: "src", Label: "main.go", Renamable: true},
		{Key: "app", ParentKey: "src", Label: "app.go", Renamable: true},
		{Key: "docs", Label: "docs", AcceptsChildren: true, DragDisabled: true},
		{Key: "readme", ParentKey: "docs", Label: "README.md", Renamable: true},
		{Key: "remote", Label: "按需加载", AcceptsChildren: true, DragDisabled: true, ChildrenState: tree.ChildrenUnloaded},
	}
}

func demoAdvancedTree(m demoState, send ui.Send[Msg]) ui.Widget {
	files := tree.New("advanced-tree", m.treeSelected, tree.ItemsFromSimple(m.treeItems)).
		DataVersion(m.treeVersion).Filter(m.treeFilter).Guides(true).Checkable(true).
		CheckedKeys(m.treeChecked).HalfCheckedKeys(m.treeHalfChecked).ExpandedKeys(m.treeExpanded).
		OnChange(setString(send, func(m *demoState) *string { return &m.treeSelected })).
		OnExpandedChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.treeExpanded = keys })) }).
		OnCheckedChange(func(checked, half []string) {
			send(demoMsg(func(m *demoState) { m.treeChecked, m.treeHalfChecked = checked, half }))
		}).OnLoadChildrenEvent(func(event tree.LoadEvent) {
		send(demoMsg(func(m *demoState) { loadExplorer(m, event.Key) }))
	}).OnRename(func(key, label string) {
		send(demoMsg(func(m *demoState) { renameExplorer(m, key, label) }))
	}).CanDrop(func(event tree.DropEvent) bool { return canMoveExplorer(m.treeItems, event) }).
		OnDrop(func(event tree.DropEvent) {
			send(demoMsg(func(m *demoState) { moveExplorer(m, event) }))
			send(logMsg("已移动文件到：" + event.TargetKey))
		})
	return layout.Column(
		input.New(m.treeFilter).Key("tree-filter").Placeholder("搜索文件，自动保留祖先目录").Label("过滤文件").FullWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.treeFilter })),
		muted("展开“按需加载”获取本地示例；选中文件按 F2 重命名；将文件拖入已加载的目录。"),
		fixedHeight(280, files),
		text.New("已勾选："+strings.Join(m.treeChecked, ", ")),
	).Gap(12)
}

func loadExplorer(m *demoState, key string) {
	for index, item := range m.treeItems {
		if item.Key != key || item.ChildrenState != tree.ChildrenUnloaded {
			continue
		}
		m.treeItems = slices.Clone(m.treeItems)
		m.treeItems[index].ChildrenState = tree.ChildrenLoaded
		m.treeItems = append(m.treeItems,
			tree.SimpleItem{Key: "remote-guide", ParentKey: key, Label: "guide.md", Renamable: true},
			tree.SimpleItem{Key: "remote-notes", ParentKey: key, Label: "notes.md", Renamable: true},
		)
		m.treeVersion++
		return
	}
}

func renameExplorer(m *demoState, key, label string) {
	label = strings.TrimSpace(label)
	if label == "" {
		return
	}
	for index, item := range m.treeItems {
		if item.Key == key && item.Renamable {
			m.treeItems = slices.Clone(m.treeItems)
			m.treeItems[index].Label = label
			m.treeVersion++
			return
		}
	}
}

// Keep the demo focused on moving files into loaded folders. Restricting the
// source to leaves also makes cycles impossible without a second tree model.
func canMoveExplorer(items []tree.SimpleItem, event tree.DropEvent) bool {
	if event.Position != tree.DropInside || event.SourceKey == event.TargetKey {
		return false
	}
	var source, target bool
	for _, item := range items {
		if item.Key == event.SourceKey {
			source = !item.DragDisabled && !item.AcceptsChildren
		}
		if item.Key == event.TargetKey {
			target = item.AcceptsChildren && item.ChildrenState == tree.ChildrenLoaded
		}
	}
	return source && target
}

func moveExplorer(m *demoState, event tree.DropEvent) {
	if !canMoveExplorer(m.treeItems, event) {
		return
	}
	m.treeItems = slices.Clone(m.treeItems)
	for index, item := range m.treeItems {
		if item.Key == event.SourceKey {
			m.treeItems[index].ParentKey = event.TargetKey
			break
		}
	}
	if !slices.Contains(m.treeExpanded, event.TargetKey) {
		m.treeExpanded = append(slices.Clone(m.treeExpanded), event.TargetKey)
	}
	// Existing cascade checks describe the old hierarchy; reset them on a move.
	m.treeChecked, m.treeHalfChecked = []string{}, []string{}
	m.treeVersion++
}

func demoAnimation(m demoState, send ui.Send[Msg]) ui.Widget {
	target, height := float32(0), unit.Dp(48)
	if m.animationExpanded {
		target, height = 1, 128
	}
	tween := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		value := animation.Tween("demo-tween", target).Duration(700*time.Millisecond).Value(ctx, gtx)
		width := gtx.Metric.PxToDp(gtx.Constraints.Max.X)
		active := ctx.Theme()
		return layout.Box(text.New(fmt.Sprintf("%.0f%%", value*100)).Color(active.Palette.AccentForeground)).
			Style(style.New().Width(64+(width-64)*unit.Dp(value)).Height(40).Padding(8).Radius(6).
				Background(style.Color(active.Palette.Accent))).Layout(ctx, gtx)
	})
	panel := layout.Box(text.New("尺寸变化由 AnimateLayout 平滑过渡")).
		Style(style.New().FillWidth().Height(height).Padding(12))
	return layout.Column(
		button.New("切换动画目标").OnClick(func() {
			send(demoMsg(func(m *demoState) { m.animationExpanded = !m.animationExpanded }))
		}),
		muted("Tween · 数值过渡"), tween,
		muted("AnimateLayout · 布局尺寸过渡"),
		animation.AnimateLayout("demo-layout-animation", panel).Duration(500*time.Millisecond),
	).Gap(16)
}
