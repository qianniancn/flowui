package table

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type tableProbe struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
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

func TestTableColumnWidthsFillAndOverflow(t *testing.T) {
	columns := []Column{
		{Key: "fixed", Width: 120},
		{Key: "wide", MinWidth: 100, Weight: 2},
		{Key: "narrow", MinWidth: 80, Weight: 1},
	}
	ctx := tableTestContext(nil)
	gtx := tableLayoutContext(nil, image.Pt(600, 300), time.Time{})
	resolved := New("table", columns, nil).resolveColumns(ctx, gtx, 600)
	if resolved.width != 600 || sumWidths(resolved) != 600 {
		t.Fatalf("resolved widths = %#v total %d", resolved, sumWidths(resolved))
	}
	if resolved.widths[1] <= resolved.widths[2] {
		t.Fatalf("weighted widths = %#v", resolved.widths)
	}

	resolved = New("table", columns, nil).MinWidth(760).resolveColumns(ctx, gtx, 320)
	if resolved.width != 760 || sumWidths(resolved) != 760 {
		t.Fatalf("overflow widths = %#v total %d", resolved, sumWidths(resolved))
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
	resolved := New("members", columns, rows).resolveColumns(ctx, gtx, 320)
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
	if header.foreground != activeTheme.Palette.MutedForeground || header.background != activeTheme.Palette.SurfaceSecondary {
		t.Fatalf("header colors = %#v/%#v", header.foreground, header.background)
	}
	if empty.foreground != activeTheme.Palette.Foreground || empty.background != activeTheme.Palette.Surface {
		t.Fatalf("empty colors = %#v/%#v", empty.foreground, empty.background)
	}
	if footer.foreground != activeTheme.Palette.Foreground || footer.background != activeTheme.Palette.SurfaceSecondary {
		t.Fatalf("footer colors = %#v/%#v", footer.foreground, footer.background)
	}
}

func TestTableThemeMatchesHeroUIGeometry(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := activeTheme.Components.Table
	if tokens.RootPadding != 4 || tokens.HeaderHeight != 40 || tokens.RowMinHeight != 48 || tokens.CellPaddingX != 16 || tokens.CellPaddingY != 12 {
		t.Fatalf("Table geometry = %+v", tokens)
	}
	if tokens.HeaderTextSize != 12 || tokens.CellTextSize != 14 || tokens.ColumnSeparatorHeight != 16 {
		t.Fatalf("Table typography/separator geometry = %+v", tokens)
	}
	primary := tableStyleFor(&activeTheme, VariantPrimary)
	if primary.root != activeTheme.Palette.SurfaceSecondary || primary.body != activeTheme.Palette.Surface {
		t.Fatalf("primary style = %#v/%#v", primary.root, primary.body)
	}
	secondary := tableStyleFor(&activeTheme, VariantSecondary)
	if secondary.root.A != 0 || secondary.body.A != 0 || secondary.header != activeTheme.Palette.SurfaceSecondary {
		t.Fatalf("secondary style = %#v", secondary)
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
