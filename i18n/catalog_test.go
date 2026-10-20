package i18n

import "testing"

func TestCatalogTranslationFallback(t *testing.T) {
	catalog := NewCatalog(
		Translation{Language: Language("fr"), Key: ActionClose, Value: "Fermer"},
		Translation{Language: Language("fr-CA"), Key: ActionCopy, Value: "Copier (CA)"},
	)
	tests := []struct {
		language Language
		key      Key
		want     string
	}{
		{Language("fr-CA"), ActionClose, "Fermer"},
		{Language("fr-ca"), ActionCopy, "Copier (CA)"},
		{Language("fr-FR"), ActionCopy, "Copy"},
		{LanguageChinese, ActionClose, "关闭"},
		{Language("unknown"), ActionClose, "Close"},
		{LanguageEnglish, Key("unknown.key"), "unknown.key"},
	}
	for _, test := range tests {
		if got := catalog.Text(test.language, test.key); got != test.want {
			t.Errorf("Text(%q, %q) = %q, want %q", test.language, test.key, got, test.want)
		}
	}
}

func TestCatalogWithReturnsIndependentValue(t *testing.T) {
	base := NewCatalog(Translation{Language: Language("fr"), Key: ActionClose, Value: "Fermer"})
	changed := base.With(Language("fr"), ActionClose, "Quitter")

	if got := base.Text(Language("fr"), ActionClose); got != "Fermer" {
		t.Fatalf("base catalog changed to %q", got)
	}
	if got := changed.Text(Language("fr"), ActionClose); got != "Quitter" {
		t.Fatalf("changed catalog returned %q", got)
	}
}

func TestCatalogFallsBackAcrossLanguageSubtags(t *testing.T) {
	catalog := NewCatalog(Translation{Language: Language("zh"), Key: ActionClose, Value: "關閉"})
	if got := catalog.Text(Language("zh-Hant-TW"), ActionClose); got != "關閉" {
		t.Fatalf("regional fallback = %q, want 關閉", got)
	}
}

func TestBuiltInCatalogTranslationsAreComplete(t *testing.T) {
	keys := []Key{
		ActionCut, ActionCopy, ActionPaste, ActionSelectAll, ActionClose,
		DatePickerChooseDate, DatePickerChooseDateRange, DateFieldYear, DateFieldMonth, DateFieldDay,
		ColorPickerPickColor, ColorPickerWheel, ColorPickerAlpha, ColorPickerHue, ColorPickerArea, ColorPickerColor,
		TitleBarMinimize, TitleBarRestore, TitleBarMaximize, TitleBarClose,
		ComboBoxNoResults, CollectionNoItems, MenuNoActions, SidebarNoDestinations,
		CommandPaletteTitle, CommandPaletteSearch, CommandPaletteEmpty,
		SidebarGroupExpand, SidebarGroupCollapse, TableNoResults,
		ChartNoData, GanttNoVisibleTasks, SelectPlaceholderItem, SelectPlaceholderItems, TableLoadingMore,
		PaginationLabel, PaginationPrevious, PaginationNext, MenubarLabel, TabsMore, TabsList, TabsTab,
		TabsSelectedTab, TabsScrollLeft, TabsScrollRight, TabsScrollUp, TabsScrollDown, TabsAddTab,
		TabsClose, TabsPanel,
		TreePartiallySelected, TreeCollapsed, TreeExpanded, TreeChildrenNotLoaded, TreeLoadingChildren,
		TreeLoadChildrenFailed, SpinnerLoading,
		TableSelectAllRows, TableResizeColumn, TableSortColumn,
		NodeGraphDescription, NodeGraphFocused,
		ChartBarDefaultLabel, ChartBarDescription, ChartLineDefaultLabel, ChartLineDescription,
		ChartCandlestickDefaultLabel, ChartCandlestickDescription, ChartPieDefaultLabel,
		ChartPieDescriptionOne, ChartPieDescriptionMany, ChartCandlestickOpen, ChartCandlestickHigh,
		ChartCandlestickLow, ChartCandlestickClose, ChartMoreSeries, GanttProgress,
		GanttTimeMonthDay, GanttTimeMonthDayTime, GanttTimeMonthYear, GanttTimeYear,
		ProgressDefault, ProgressInProgress, ProgressValue, MeterDefault, SliderDefault, SliderValue,
		SliderMinimum, SliderMaximum, ToolbarDefault, WorkbenchNextTab, WorkbenchPreviousTab,
		WorkbenchCloseTab, WorkbenchToggleSidebar, WorkbenchTogglePanel, WorkbenchToggleStatus,
		ShortcutLeft, ShortcutRight, ShortcutUp, ShortcutDown, ShortcutReturn, ShortcutEnter, ShortcutEscape,
		ShortcutHome, ShortcutEnd, ShortcutBackspace, ShortcutDelete, ShortcutPageUp, ShortcutPageDown,
		ShortcutTab, ShortcutSpace,
		CalendarMonthJanuary, CalendarMonthFebruary, CalendarMonthMarch, CalendarMonthApril, CalendarMonthMay,
		CalendarMonthJune, CalendarMonthJuly, CalendarMonthAugust, CalendarMonthSeptember, CalendarMonthOctober,
		CalendarMonthNovember, CalendarMonthDecember, CalendarWeekdaySunday, CalendarWeekdayMonday, CalendarWeekdayTuesday,
		CalendarWeekdayWednesday, CalendarWeekdayThursday, CalendarWeekdayFriday, CalendarWeekdaySaturday,
		CalendarMonthFullJanuary, CalendarMonthFullFebruary, CalendarMonthFullMarch, CalendarMonthFullApril,
		CalendarMonthFullMay, CalendarMonthFullJune, CalendarMonthFullJuly, CalendarMonthFullAugust,
		CalendarMonthFullSeptember, CalendarMonthFullOctober, CalendarMonthFullNovember, CalendarMonthFullDecember,
		CalendarMonthYear,
	}
	for _, language := range []Language{LanguageEnglish, LanguageChinese} {
		t.Run(string(language), func(t *testing.T) {
			messages := builtInCatalog.entries[language]
			if len(messages) != len(keys) {
				t.Fatalf("built-in message count = %d, want %d", len(messages), len(keys))
			}
			for _, key := range keys {
				if messages[key] == "" {
					t.Errorf("built-in %s translation for %q is missing", language, key)
				}
			}
		})
	}
}
