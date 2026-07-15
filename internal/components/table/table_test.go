package table

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type tableProbe struct {
	size        image.Point
	layouts     int
	constraints layout.Constraints
	foreground  color.NRGBA
	background  color.NRGBA
}

type tableOverlayProbe struct {
	anchor image.Rectangle
}

func (p *tableOverlayProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       "cell-overlay",
		Anchor:    image.Rect(1, 2, 2, 3),
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			p.anchor = anchor
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(80, 20))}
}

func (p *tableProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.constraints = gtx.Constraints
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestTableOptionsUseValueSemantics(t *testing.T) {
	columns, rows := tableTestData()
	base := New("members", columns, rows)
	configured := base.
		Variant(VariantSecondary).
		SelectionMode(SelectionMultiple).
		SelectedKey("kate").
		SelectedKeys([]string{"kate"}).
		SortDescriptor(SortDescriptor{Column: "name", Direction: SortDescending}).
		DisabledKeys([]string{"sara"}).
		EmptyText("Empty").
		EmptyContent(&tableProbe{}).
		Footer(&tableProbe{}).
		OnChange(func(string) {}).
		OnSelectionChange(func([]string) {}).
		OnSortChange(func(SortDescriptor) {}).
		OnAction(func(string) {}).
		Disabled(true).
		AllowEmptySelection().
		ShowSelectionIndicator().
		MaxHeight(240).
		MinWidth(640)

	if base.variant != VariantPrimary || base.selectionMode != SelectionNone || len(base.selectedKeys) != 0 || base.maxHeight != 0 {
		t.Fatal("configuring a Table mutated the base value")
	}
	if configured.variant != VariantSecondary || configured.selectionMode != SelectionMultiple || configured.selectedKey != "kate" {
		t.Fatal("selection options were not retained")
	}
	if configured.sort.Column != "name" || configured.sort.Direction != SortDescending || configured.emptyText != "Empty" {
		t.Fatal("sort or empty options were not retained")
	}
	if configured.onChange == nil || configured.onSelectionChange == nil || configured.onSortChange == nil || configured.onAction == nil {
		t.Fatal("callbacks were not retained")
	}
	if !configured.disabled || !configured.allowEmpty || !configured.selectionIndicator || configured.maxHeight != 240 || configured.minWidth != 640 {
		t.Fatal("behavior options were not retained")
	}
}

func TestTableRejectsInvalidStructure(t *testing.T) {
	tests := []struct {
		name    string
		columns []Column
		rows    []Row
	}{
		{name: "no columns"},
		{name: "empty column", columns: []Column{{}}},
		{name: "duplicate column", columns: []Column{{Key: "same"}, {Key: "same"}}},
		{name: "empty row", columns: []Column{{Key: "name"}}, rows: []Row{{Cells: []Cell{{Text: "A"}}}}},
		{name: "duplicate row", columns: []Column{{Key: "name"}}, rows: []Row{{Key: "same", Cells: []Cell{{}}}, {Key: "same", Cells: []Cell{{}}}}},
		{name: "cell mismatch", columns: []Column{{Key: "name"}, {Key: "role"}}, rows: []Row{{Key: "one", Cells: []Cell{{}}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid Table structure to panic")
				}
			}()
			new(tableState).check(test.columns, test.rows)
		})
	}
}

func TestTableTracksOverlayAnchorInsideCustomCell(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := tableTestContext(&activeTheme)
	probe := new(tableOverlayProbe)
	table := New(
		"tracked-cell",
		[]Column{{Key: "name", Label: "Name", Width: 200}},
		[]Row{
			{Key: "first", Label: "First", Cells: []Cell{{Text: "First"}}},
			{Key: "member", Label: "Member", Cells: []Cell{{Content: probe}}},
		},
	).Variant(VariantSecondary)
	gtx := tableLayoutContext(nil, image.Pt(400, 240), time.Unix(1, 0))
	frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	table.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	tokens := activeTheme.Components.Table
	cellHeight := gtx.Dp(tokens.CellPaddingY)*2 + 20
	rowOffset := max((gtx.Dp(tokens.RowMinHeight)-cellHeight)/2, 0)
	want := image.Rect(
		gtx.Dp(tokens.CellPaddingX)+1,
		gtx.Dp(tokens.HeaderHeight)+gtx.Dp(tokens.RowMinHeight)+rowOffset+gtx.Dp(tokens.CellPaddingY)+2,
		gtx.Dp(tokens.CellPaddingX)+2,
		gtx.Dp(tokens.HeaderHeight)+gtx.Dp(tokens.RowMinHeight)+rowOffset+gtx.Dp(tokens.CellPaddingY)+3,
	)
	if probe.anchor != want {
		t.Fatalf("custom cell overlay anchor = %v, want %v", probe.anchor, want)
	}
}

func TestTableSingleSelectionAndAction(t *testing.T) {
	columns, rows := tableTestData()
	var changed, action string
	table := New("members", columns, rows).
		SelectionMode(SelectionSingle).
		SelectedKey("kate").
		OnChange(func(key string) { changed = key }).
		OnAction(func(key string) { action = key })
	table.activate("john")
	if changed != "john" || action != "john" {
		t.Fatalf("callbacks = changed %q action %q", changed, action)
	}

	changed = "not-called"
	table.AllowEmptySelection().activate("kate")
	if changed != "" {
		t.Fatalf("allow-empty changed = %q, want empty", changed)
	}
}

func TestTableMultipleSelectionTogglesWithoutMutation(t *testing.T) {
	columns, rows := tableTestData()
	selected := []string{"kate"}
	var changed []string
	New("members", columns, rows).
		SelectionMode(SelectionMultiple).
		SelectedKeys(selected).
		OnSelectionChange(func(keys []string) { changed = keys }).
		activate("john")
	if !sameKeys(selected, []string{"kate"}) {
		t.Fatalf("selected input mutated: %#v", selected)
	}
	if !sameKeys(changed, []string{"kate", "john"}) {
		t.Fatalf("changed = %#v", changed)
	}
}

func TestTableShiftSelectionSelectsRangeAndSkipsDisabledRows(t *testing.T) {
	columns, rows := tableTestData()
	var changed []string
	stateValue := &tableState{selectionAnchor: "kate"}
	table := New("members", columns, rows).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"kate"}).
		DisabledKeys([]string{"john"}).
		OnSelectionChange(func(keys []string) { changed = keys })
	table.activateWithModifiers(stateValue, "sara", key.ModShift)
	if !sameKeys(changed, []string{"kate", "sara"}) {
		t.Fatalf("range selection = %v, want kate and sara", changed)
	}
}

func TestTableSelectAllSkipsDisabledAndPreservesDisabledSelection(t *testing.T) {
	columns, rows := tableTestData()
	var changed []string
	table := New("members", columns, rows).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"sara"}).
		DisabledKeys([]string{"sara"}).
		OnSelectionChange(func(keys []string) { changed = keys })
	table.toggleAll()
	if !sameKeys(changed, []string{"kate", "john", "sara"}) {
		t.Fatalf("select all = %#v", changed)
	}

	table.selectedKeys = changed
	table.toggleAll()
	if !sameKeys(changed, []string{"sara"}) {
		t.Fatalf("clear all = %#v, want disabled selection preserved", changed)
	}
}

func TestTableSortCyclesAndChangesColumn(t *testing.T) {
	columns, rows := tableTestData()
	var got SortDescriptor
	table := New("members", columns, rows).
		SortDescriptor(SortDescriptor{Column: "name", Direction: SortAscending}).
		OnSortChange(func(sort SortDescriptor) { got = sort })
	table.requestSort("name")
	if got.Column != "name" || got.Direction != SortDescending {
		t.Fatalf("same-column sort = %#v", got)
	}
	table.requestSort("role")
	if got.Column != "role" || got.Direction != SortAscending {
		t.Fatalf("new-column sort = %#v", got)
	}
}

func TestTableRowClickSelectsAndDisabledRowDoesNot(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(nil)
	stateValue := new(tableState)
	stateValue.beginFrame()
	stateValue.row("john").clickable.Click()
	stateValue.endFrame()
	tableSetState(ctx, "members", stateValue)
	changed := ""
	New("members", columns, rows).
		SelectionMode(SelectionSingle).
		OnChange(func(key string) { changed = key }).
		Layout(ctx, tableLayoutContext(nil, image.Pt(640, 320), time.Time{}))
	if changed != "john" {
		t.Fatalf("row click changed = %q", changed)
	}

	ctx = tableTestContext(nil)
	stateValue = new(tableState)
	stateValue.beginFrame()
	stateValue.row("john").clickable.Click()
	stateValue.endFrame()
	tableSetState(ctx, "members", stateValue)
	changed = ""
	New("members", columns, rows).
		SelectionMode(SelectionSingle).
		DisabledKeys([]string{"john"}).
		OnChange(func(key string) { changed = key }).
		Layout(ctx, tableLayoutContext(nil, image.Pt(640, 320), time.Time{}))
	if changed != "" {
		t.Fatalf("disabled row changed = %q", changed)
	}
}

func TestTableKeyboardNavigationSkipsDisabledRows(t *testing.T) {
	columns, rows := tableTestData()
	table := New("members", columns, rows).DisabledKeys([]string{"john"})
	next, ok := moveRow(table, 0, 1)
	if !ok || rows[next].Key != "sara" {
		t.Fatalf("next row = %d/%v key %q, want sara", next, ok, rows[next].Key)
	}
	first, ok := firstEnabledRow(table)
	if !ok || rows[first].Key != "kate" {
		t.Fatalf("first row = %d/%v", first, ok)
	}
}

func TestTableKeyboardEnterSelectsFocusedRow(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(nil)
	router := new(input.Router)
	selected := ""
	widget := func() Widget {
		return New("members", columns, rows).
			SelectionMode(SelectionSingle).
			SelectedKey(selected).
			OnChange(func(key string) { selected = key })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget(), start)
	stateValue := tablePeekState(ctx, "members")
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.rows["sara"].clickable})
	layoutTableFrame(ctx, router, widget(), start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameReturn, State: key.Press}, key.Event{Name: key.NameReturn, State: key.Release})
	layoutTableFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	if selected != "sara" {
		t.Fatalf("selected = %q, want sara", selected)
	}
}

func TestTableKeyboardShiftSpaceRequestsRangeSelection(t *testing.T) {
	columns, rows := tableTestData()
	table := New("members", columns, rows).SelectionMode(SelectionMultiple)
	stateValue := new(tableState)
	result := tableKeyResult{}
	stateValue.handleActivation(key.Event{Name: key.NameSpace, State: key.Press, Modifiers: key.ModShift}, table, 1, &result)
	stateValue.handleActivation(key.Event{Name: key.NameSpace, State: key.Release, Modifiers: key.ModShift}, table, 1, &result)
	if result.rangeKey != "john" || result.actionKey != "" {
		t.Fatalf("keyboard range result = %+v", result)
	}
}

func TestTableColumnWidthsFillAndOverflow(t *testing.T) {
	columns := []Column{
		{Key: "fixed", Width: 120},
		{Key: "wide", MinWidth: 100, Weight: 2},
		{Key: "narrow", MinWidth: 80, Weight: 1},
	}
	ctx := tableTestContext(nil)
	gtx := tableLayoutContext(nil, image.Pt(600, 300), time.Time{})
	resolved := New("table", columns, nil).resolveColumns(ctx, gtx, nil, 600)
	if resolved.width != 600 || sumWidths(resolved) != 600 {
		t.Fatalf("resolved widths = %#v total %d", resolved, sumWidths(resolved))
	}
	if resolved.widths[1] <= resolved.widths[2] {
		t.Fatalf("weighted widths = %#v", resolved.widths)
	}

	resolved = New("table", columns, nil).MinWidth(760).resolveColumns(ctx, gtx, nil, 320)
	if resolved.width != 760 || sumWidths(resolved) != 760 {
		t.Fatalf("overflow widths = %#v total %d", resolved, sumWidths(resolved))
	}
}

func TestTableColumnWidthsRespectMaximums(t *testing.T) {
	columns := []Column{
		{Key: "fixed", Width: 120, MaxWidth: 120},
		{Key: "flex", MinWidth: 100, MaxWidth: 160},
	}
	ctx := tableTestContext(nil)
	gtx := tableLayoutContext(nil, image.Pt(600, 300), time.Time{})
	resolved := New("table", columns, nil).resolveColumns(ctx, gtx, nil, 600)
	if resolved.width != 600 || resolved.widths[0] != 120 || resolved.widths[1] != 160 {
		t.Fatalf("resolved capped widths = %#v", resolved)
	}
	if sumWidths(resolved) >= resolved.width {
		t.Fatalf("capped columns should leave trailing table space: %#v", resolved)
	}
}

func TestResizableFlexColumnAdaptsUntilDragged(t *testing.T) {
	columns := []Column{
		{Key: "name", MinWidth: 120, Weight: 2, Resizable: true},
		{Key: "role", MinWidth: 100, Weight: 1},
	}
	ctx := tableTestContext(nil)
	stateValue := new(tableState)
	stateValue.beginFrame()
	first := New("table", columns, nil).resolveColumns(ctx, tableLayoutContext(nil, image.Pt(600, 300), time.Time{}), stateValue, 600)
	stateValue.endFrame()
	stateValue.beginFrame()
	second := New("table", columns, nil).resolveColumns(ctx, tableLayoutContext(nil, image.Pt(780, 300), time.Time{}), stateValue, 780)
	stateValue.endFrame()
	if sumWidths(first) != 600 || sumWidths(second) != 780 || second.widths[0] <= first.widths[0] {
		t.Fatalf("resizable flex widths did not adapt: first %#v second %#v", first, second)
	}
}

func TestTableColumnResizerHandlesPointerDrag(t *testing.T) {
	columns := []Column{
		{Key: "name", Label: "Name", Width: 160, MinWidth: 100, MaxWidth: 260, Resizable: true},
		{Key: "role", Label: "Role", MinWidth: 120},
	}
	rows := []Row{{Key: "one", Cells: []Cell{{Text: "One"}, {Text: "Engineer"}}}}
	ctx := tableTestContext(nil)
	router := new(input.Router)
	width := 0
	widget := New("resizable", columns, rows).OnColumnResize(func(_ string, value int) { width = value })
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(164, 18)})
	layoutTableFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(224, 18)})
	layoutTableFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if width != 220 {
		t.Fatalf("resized width = %d, want 220", width)
	}
	stateValue := tablePeekState(ctx, "resizable")
	if !stateValue.columns["name"].resize.dragging {
		t.Fatal("column resizer stopped before pointer release")
	}
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(224, 18)})
	layoutTableFrame(ctx, router, widget, start.Add(3*time.Millisecond))
	if resize := stateValue.columns["name"].resize; resize.dragging {
		t.Fatal("column resizer remained active after pointer release")
	}
}

func TestVirtualTableOnlyRequestsVisibleRows(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}, {Key: "role", Label: "Role"}}
	calls := 0
	table := NewVirtual("virtual", columns, 10_000, func(index int) Row {
		calls++
		return Row{
			Key:   fmt.Sprintf("row-%d", index),
			Cells: []Cell{{Text: fmt.Sprintf("User %d", index)}, {Text: "Engineer"}},
		}
	}).MaxHeight(140).
		RowHeight(42).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"row-0"}).
		DisabledKeys([]string{"row-2"}).
		ShowSelectionIndicator()
	ctx := tableTestContext(nil)
	frame.BeginFrame(ctx)
	table.Layout(ctx, tableLayoutContext(nil, image.Pt(640, 200), time.Unix(1, 0)))
	frame.EndFrame(ctx)
	if calls == 0 || calls >= 40 {
		t.Fatalf("virtual provider calls = %d, want only visible rows", calls)
	}
	stateValue := tablePeekState(ctx, "virtual")
	if len(stateValue.rows) >= 20 {
		t.Fatalf("retained virtual row states = %d", len(stateValue.rows))
	}
}

func TestVirtualTableDoesNotResolveSelectedRowWithoutKeyboardEvent(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	calls := 0
	table := NewVirtual("virtual-selected", columns, 10_000, func(index int) Row {
		calls++
		return Row{Key: fmt.Sprintf("row-%d", index), Cells: []Cell{{Text: "User"}}}
	}).MaxHeight(140).
		RowHeight(42).
		SelectionMode(SelectionSingle).
		SelectedKey("row-9999")
	ctx := tableTestContext(nil)
	frame.BeginFrame(ctx)
	table.Layout(ctx, tableLayoutContext(nil, image.Pt(640, 200), time.Unix(1, 0)))
	frame.EndFrame(ctx)
	if calls == 0 || calls >= 40 {
		t.Fatalf("virtual provider calls = %d, want only visible rows", calls)
	}
}

func TestTableLoadMoreLatchesUntilRowsGrow(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{{Key: "one", Cells: []Cell{{Text: "One"}}}}
	ctx := tableTestContext(nil)
	router := new(input.Router)
	requests := 0
	widget := func(rows []Row, loading bool) Widget {
		return New("loading", columns, rows).MaxHeight(180).LoadMore(true, loading, func() { requests++ })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget(rows, false), start)
	layoutTableFrame(ctx, router, widget(rows, true), start.Add(time.Millisecond))
	layoutTableFrame(ctx, router, widget(rows, false), start.Add(2*time.Millisecond))
	if requests != 1 {
		t.Fatalf("load requests = %d, want one latched request", requests)
	}
	rows = append(rows, Row{Key: "two", Cells: []Cell{{Text: "Two"}}})
	layoutTableFrame(ctx, router, widget(rows, false), start.Add(3*time.Millisecond))
	if requests != 2 {
		t.Fatalf("load requests after growth = %d, want 2", requests)
	}
}

func TestTableLayoutUsesThemeMaxHeightAndHorizontalViewport(t *testing.T) {
	columns, rows := tableTestData()
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Table.HeaderHeight = 44
	activeTheme.Components.Table.RowMinHeight = 52
	ctx := tableTestContext(&activeTheme)
	dims := New("members", columns, rows).
		MinWidth(760).
		MaxHeight(150).
		Layout(ctx, tableLayoutContext(nil, image.Pt(420, 300), time.Time{}))
	if dims.Size.X != 420 {
		t.Fatalf("table width = %d, want viewport 420", dims.Size.X)
	}
	if dims.Size.Y > 154 {
		t.Fatalf("table height = %d, want <= 154", dims.Size.Y)
	}
	stateValue := tablePeekState(ctx, "members")
	if stateValue.horizontal.Axis != layout.Horizontal || stateValue.vertical.Axis != layout.Vertical {
		t.Fatalf("scroll axes = %v/%v", stateValue.horizontal.Axis, stateValue.vertical.Axis)
	}
}

func TestTableContentHeightIncludesHeaderAndBody(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{{Key: "one", Cells: []Cell{{Text: "One"}}}}
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Table.HeaderHeight = 40
	activeTheme.Components.Table.RowMinHeight = 48
	ctx := tableTestContext(&activeTheme)
	gtx := tableLayoutContext(nil, image.Pt(320, 200), time.Time{})
	stateValue := new(tableState)
	stateValue.beginFrame()
	defer stateValue.endFrame()
	stateValue.check(columns, rows)
	resolved := New("members", columns, rows).resolveColumns(ctx, gtx, nil, 320)
	dims := New("members", columns, rows).layoutContent(ctx, gtx, stateValue, resolved, tableStyleFor(&activeTheme, VariantPrimary))
	if dims.Size.Y != 88 {
		t.Fatalf("content height = %d, want header 40 + row 48", dims.Size.Y)
	}
}

func TestTableCustomCellInheritsSelectedColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	probe := &tableProbe{size: image.Pt(40, 16)}
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{{Key: "selected", Label: "Selected", Cells: []Cell{{Content: probe}}}}
	New("members", columns, rows).
		SelectionMode(SelectionSingle).
		SelectedKey("selected").
		Layout(tableTestContext(&activeTheme), tableLayoutContext(nil, image.Pt(320, 160), time.Time{}))
	want := tableRowStyleFor(&activeTheme, VariantPrimary, true, false, false, false).background
	if probe.layouts != 1 || probe.foreground != activeTheme.Palette.Foreground || probe.background != want {
		t.Fatalf("probe = layouts %d colors %#v/%#v, want fg %#v bg %#v", probe.layouts, probe.foreground, probe.background, activeTheme.Palette.Foreground, want)
	}
}

func TestTableComposedRegionsInheritSurfaceColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	header := &tableProbe{size: image.Pt(40, 16)}
	empty := &tableProbe{size: image.Pt(40, 16)}
	footer := &tableProbe{size: image.Pt(40, 16)}
	columns := []Column{{Key: "name", Label: "Name", Header: header}}
	New("members", columns, nil).
		EmptyContent(empty).
		Footer(footer).
		Layout(tableTestContext(&activeTheme), tableLayoutContext(nil, image.Pt(320, 240), time.Time{}))
	wantHeader := activeTheme.Palette.SurfaceTertiary
	if header.foreground != activeTheme.Palette.MutedForeground || header.background != wantHeader {
		t.Fatalf("header colors = %#v/%#v", header.foreground, header.background)
	}
	if header.constraints.Min.Y != 0 || header.constraints.Max.Y != int(activeTheme.Components.Table.HeaderHeight) {
		t.Fatalf("header height constraints = %+v", header.constraints)
	}
	if empty.foreground != activeTheme.Palette.Foreground || empty.background != activeTheme.Palette.Surface {
		t.Fatalf("empty colors = %#v/%#v", empty.foreground, empty.background)
	}
	if footer.foreground != activeTheme.Palette.Foreground || footer.background != activeTheme.Palette.SurfaceTertiary {
		t.Fatalf("footer colors = %#v/%#v", footer.foreground, footer.background)
	}
}

func TestTableThemeMatchesHeroUIStyle(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := activeTheme.Components.Table
	if tokens.RootPadding != 4 || tokens.RootRadius != 20 || tokens.HeaderRadius != 16 || tokens.BodyRadius != 16 {
		t.Fatalf("Table root geometry = %+v", tokens)
	}
	if tokens.HeaderHeight != 36 || tokens.RowMinHeight != 44 || tokens.CellPaddingX != 16 || tokens.CellPaddingY != 12 {
		t.Fatalf("Table geometry = %+v", tokens)
	}
	if tokens.HeaderTextSize != 12 || tokens.CellTextSize != 14 || tokens.ColumnSeparatorHeight != 16 {
		t.Fatalf("Table typography/separator geometry = %+v", tokens)
	}
	primary := tableStyleFor(&activeTheme, VariantPrimary)
	if primary.root != activeTheme.Palette.SurfaceTertiary || primary.body != activeTheme.Palette.Surface {
		t.Fatalf("primary style = %#v/%#v", primary.root, primary.body)
	}
	secondary := tableStyleFor(&activeTheme, VariantSecondary)
	wantHeader := activeTheme.Palette.SurfaceTertiary
	if secondary.root.A != 0 || secondary.body.A != 0 || secondary.header != wantHeader {
		t.Fatalf("secondary style = %#v", secondary)
	}
	wantHeaderSeparator := activeTheme.Palette.SeparatorColor()
	wantColumnSeparator := wantHeaderSeparator
	wantHeaderSeparator.A = byte(float32(wantHeaderSeparator.A)*0.5 + 0.5)
	wantRowSeparator := render.LerpColor(activeTheme.Palette.Surface, activeTheme.Palette.Foreground, 0.19)
	wantRowSeparator.A = byte(float32(wantRowSeparator.A)*0.5 + 0.5)
	if primary.columnSeparator != wantColumnSeparator || primary.headerSeparator != wantHeaderSeparator || primary.rowSeparator != wantRowSeparator {
		t.Fatalf("Table separators = %#v", primary)
	}
}

func TestTableRowStatesMatchHeroUIVariants(t *testing.T) {
	themes := []struct {
		name  string
		value theme.Theme
	}{
		{name: "light", value: theme.DefaultTheme()},
		{name: "dark", value: theme.DarkTheme()},
	}
	for _, test := range themes {
		t.Run(test.name, func(t *testing.T) {
			activeTheme := test.value
			primaryHover := tableRowStyleFor(&activeTheme, VariantPrimary, false, true, false, false)
			primarySelected := tableRowStyleFor(&activeTheme, VariantPrimary, true, true, false, false)
			secondaryHover := tableRowStyleFor(&activeTheme, VariantSecondary, false, true, false, false)
			secondarySelected := tableRowStyleFor(&activeTheme, VariantSecondary, true, true, false, false)

			if want := render.LerpColor(activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.Surface, 0.4); primaryHover.background != want {
				t.Fatalf("primary hover = %#v, want %#v", primaryHover.background, want)
			}
			if want := render.LerpColor(activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.Surface, 0.1); primarySelected.background != want {
				t.Fatalf("primary selected = %#v, want %#v", primarySelected.background, want)
			}
			if want := render.LerpColor(activeTheme.Palette.Background, activeTheme.Palette.DefaultColor(), 0.5); secondaryHover.background != want {
				t.Fatalf("secondary hover = %#v, want %#v", secondaryHover.background, want)
			}
			if want := render.LerpColor(activeTheme.Palette.Background, activeTheme.Palette.Surface, 0.1); secondarySelected.background != want {
				t.Fatalf("secondary selected = %#v, want %#v", secondarySelected.background, want)
			}
		})
	}
}

func tableTestData() ([]Column, []Row) {
	columns := []Column{
		{Key: "name", Label: "Name", Sortable: true, RowHeader: true},
		{Key: "role", Label: "Role"},
	}
	rows := []Row{
		{Key: "kate", Label: "Kate Moore", Cells: []Cell{{Text: "Kate Moore"}, {Text: "CEO"}}},
		{Key: "john", Label: "John Smith", Cells: []Cell{{Text: "John Smith"}, {Text: "CTO"}}},
		{Key: "sara", Label: "Sara Johnson", Cells: []Cell{{Text: "Sara Johnson"}, {Text: "CMO"}}},
	}
	return columns, rows
}

func tableTestContext(activeTheme *theme.Theme) *frame.Context {
	return frame.New(nil, activeTheme, locale.LanguageEnglish)
}

func tableLayoutContext(router *input.Router, max image.Point, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{Constraints: layout.Constraints{Max: max}, Source: source, Ops: new(op.Ops), Now: now}
}

func layoutTableFrame(ctx *frame.Context, router *input.Router, table Widget, now time.Time) {
	gtx := tableLayoutContext(router, image.Pt(640, 320), now)
	frame.BeginFrame(ctx)
	table.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
}

func tableSetState(ctx *frame.Context, key string, value *tableState) {
	frame.UseStateWith(ctx, key, stateSlotTable, func() *tableState { return value })
}

func tablePeekState(ctx *frame.Context, key string) *tableState {
	value, _ := frame.PeekState[tableState](ctx, key, stateSlotTable)
	return value
}

func sumWidths(columns tableColumns) int {
	total := columns.selection
	for _, width := range columns.widths {
		total += width
	}
	return total
}
