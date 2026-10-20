package i18n

import (
	"fmt"
	"maps"
	"strings"

	xlanguage "golang.org/x/text/language"
)

// Translation is one message override for a language.
type Translation struct {
	Language Language
	Key      Key
	Value    string
}

// Catalog is an immutable set of application overrides. A zero Catalog uses
// FlowUI's built-in translations only.
type Catalog struct {
	entries map[Language]map[Key]string
}

// NewCatalog creates a Catalog from translations. Later entries for the same
// language and key replace earlier entries.
func NewCatalog(translations ...Translation) Catalog {
	var catalog Catalog
	for _, translation := range translations {
		catalog = catalog.With(translation.Language, translation.Key, translation.Value)
	}
	return catalog
}

// With returns a copy of the Catalog with value registered for key and
// language. Language should be a BCP-47 tag such as "fr" or "fr-CA".
func (catalog Catalog) With(language Language, key Key, value string) Catalog {
	language, ok := normalizeLanguage(language)
	if !ok || key == "" {
		return catalog.Clone()
	}
	copy := catalog.Clone()
	if copy.entries == nil {
		copy.entries = make(map[Language]map[Key]string)
	}
	if copy.entries[language] == nil {
		copy.entries[language] = make(map[Key]string)
	}
	copy.entries[language][key] = value
	return copy
}

// Clone returns a deep copy of the Catalog.
func (catalog Catalog) Clone() Catalog {
	if len(catalog.entries) == 0 {
		return Catalog{}
	}
	copy := Catalog{entries: make(map[Language]map[Key]string, len(catalog.entries))}
	for language, messages := range catalog.entries {
		copy.entries[language] = make(map[Key]string, len(messages))
		maps.Copy(copy.entries[language], messages)
	}
	return copy
}

// Text resolves key using application overrides, built-in translations, then
// English. Unknown keys are returned as their string value for diagnosis.
func (catalog Catalog) Text(language Language, key Key) string {
	if value, ok := catalog.lookup(language, key); ok {
		return value
	}
	if value, ok := builtInCatalog.lookup(language, key); ok {
		return value
	}
	if value, ok := builtInCatalog.lookup(LanguageEnglish, key); ok {
		return value
	}
	return string(key)
}

// Textf resolves a FlowUI-owned message and formats it with the message's
// documented Go-style arguments. Use this only for framework-owned dynamic
// text; application content remains responsible for its own localization.
func (catalog Catalog) Textf(language Language, key Key, args ...any) string {
	return fmt.Sprintf(catalog.Text(language, key), args...)
}

func (catalog Catalog) lookup(language Language, key Key) (string, bool) {
	language, ok := normalizeLanguage(language)
	if !ok {
		return "", false
	}
	for tag := string(language); tag != ""; {
		if messages := catalog.entries[Language(tag)]; messages != nil {
			if value, ok := messages[key]; ok {
				return value, true
			}
		}
		separator := strings.LastIndexByte(tag, '-')
		if separator < 0 {
			break
		}
		tag = tag[:separator]
	}
	return "", false
}

func normalizeLanguage(value Language) (Language, bool) {
	parsed, err := xlanguage.Parse(string(value))
	if err != nil || parsed == xlanguage.Und {
		return LanguageAuto, false
	}
	return Language(parsed.String()), true
}

var builtInCatalog = Catalog{entries: map[Language]map[Key]string{
	LanguageEnglish: {
		ActionCut: "Cut", ActionCopy: "Copy", ActionPaste: "Paste", ActionSelectAll: "Select All", ActionClose: "Close",
		DatePickerChooseDate: "Choose date", DatePickerChooseDateRange: "Choose date range",
		DateFieldYear: "Year", DateFieldMonth: "Month", DateFieldDay: "Day",
		ColorPickerPickColor: "Pick a color", ColorPickerWheel: "Color wheel", ColorPickerAlpha: "Opacity",
		ColorPickerHue: "Hue", ColorPickerArea: "Color area", ColorPickerColor: "Color",
		TitleBarMinimize: "Minimize", TitleBarRestore: "Restore", TitleBarMaximize: "Maximize", TitleBarClose: "Close",
		ComboBoxNoResults:     "No results found",
		CollectionNoItems:     "No items",
		MenuNoActions:         "No actions",
		SidebarNoDestinations: "No destinations",
		SidebarGroupExpand:    "Navigation group. Activate to expand.",
		SidebarGroupCollapse:  "Navigation group. Activate to collapse.",
		TableNoResults:        "No results",
		ChartNoData:           "No data",
		GanttNoVisibleTasks:   "No visible tasks",
		SelectPlaceholderItem: "Select an item", SelectPlaceholderItems: "Select items",
		TableLoadingMore:   "Loading more rows",
		PaginationLabel:    "Pagination",
		PaginationPrevious: "Prev", PaginationNext: "Next",
		MenubarLabel: "Menu bar", TabsMore: "More", TabsList: "Tab list", TabsTab: "Tab", TabsSelectedTab: "Selected tab",
		TabsScrollLeft: "Scroll tabs left", TabsScrollRight: "Scroll tabs right", TabsScrollUp: "Scroll tabs up", TabsScrollDown: "Scroll tabs down", TabsAddTab: "Add tab",
		TabsClose: "Close %s", TabsPanel: "Tab panel: %s",
		TreePartiallySelected: "Partially selected", TreeCollapsed: "Collapsed", TreeExpanded: "Expanded",
		TreeChildrenNotLoaded: "Children not loaded", TreeLoadingChildren: "Loading children",
		TreeLoadChildrenFailed: "Load failed; activate the expand control to retry",
		SpinnerLoading:         "Loading",
		CommandPaletteTitle:    "Commands",
		CommandPaletteSearch:   "Search commands",
		CommandPaletteEmpty:    "No matching commands",
		TableSelectAllRows:     "Select all rows", TableResizeColumn: "Resize %s column", TableSortColumn: "Sort column",
		NodeGraphDescription: "Node graph", NodeGraphFocused: "Focused node",
		ChartBarDefaultLabel: "Bar chart", ChartBarDescription: "%s, %d series, %d categories",
		ChartLineDefaultLabel: "Line chart", ChartLineDescription: "%s, %d series",
		ChartCandlestickDefaultLabel: "Candlestick chart", ChartCandlestickDescription: "%s, %d candles",
		ChartPieDefaultLabel: "Pie chart", ChartPieDescriptionOne: "%s, %d slice", ChartPieDescriptionMany: "%s, %d slices",
		ChartCandlestickOpen: "Open  %s", ChartCandlestickHigh: "High  %s", ChartCandlestickLow: "Low  %s", ChartCandlestickClose: "Close  %s",
		ChartMoreSeries: "+%d series", GanttProgress: "Progress  %.0f%%",
		GanttTimeMonthDay: "%[1]s %[2]d", GanttTimeMonthDayTime: "%[1]s %[2]d %[3]s",
		GanttTimeMonthYear: "%[1]s %04[2]d", GanttTimeYear: "%04d",
		ProgressDefault: "Progress", ProgressInProgress: "%s in progress", ProgressValue: "%s %s",
		MeterDefault: "Meter", SliderDefault: "Slider", SliderValue: "%s %s",
		SliderMinimum: "%s minimum %s", SliderMaximum: "%s maximum %s", ToolbarDefault: "Toolbar",
		WorkbenchNextTab: "Next Tab", WorkbenchPreviousTab: "Previous Tab", WorkbenchCloseTab: "Close Tab",
		WorkbenchToggleSidebar: "Toggle Sidebar", WorkbenchTogglePanel: "Toggle Bottom Panel", WorkbenchToggleStatus: "Toggle Status Bar",
		ShortcutLeft: "Left", ShortcutRight: "Right", ShortcutUp: "Up", ShortcutDown: "Down",
		ShortcutReturn: "Return", ShortcutEnter: "Enter", ShortcutEscape: "Esc", ShortcutHome: "Home", ShortcutEnd: "End",
		ShortcutBackspace: "Backspace", ShortcutDelete: "Delete", ShortcutPageUp: "PageUp", ShortcutPageDown: "PageDown",
		ShortcutTab: "Tab", ShortcutSpace: "Space",
		CalendarMonthJanuary: "Jan", CalendarMonthFebruary: "Feb", CalendarMonthMarch: "Mar", CalendarMonthApril: "Apr",
		CalendarMonthMay: "May", CalendarMonthJune: "Jun", CalendarMonthJuly: "Jul", CalendarMonthAugust: "Aug",
		CalendarMonthSeptember: "Sep", CalendarMonthOctober: "Oct", CalendarMonthNovember: "Nov", CalendarMonthDecember: "Dec",
		CalendarMonthFullJanuary: "January", CalendarMonthFullFebruary: "February", CalendarMonthFullMarch: "March", CalendarMonthFullApril: "April",
		CalendarMonthFullMay: "May", CalendarMonthFullJune: "June", CalendarMonthFullJuly: "July", CalendarMonthFullAugust: "August",
		CalendarMonthFullSeptember: "September", CalendarMonthFullOctober: "October", CalendarMonthFullNovember: "November", CalendarMonthFullDecember: "December",
		CalendarMonthYear:     "%[1]s %04[2]d",
		CalendarWeekdaySunday: "Sun", CalendarWeekdayMonday: "Mon", CalendarWeekdayTuesday: "Tue", CalendarWeekdayWednesday: "Wed",
		CalendarWeekdayThursday: "Thu", CalendarWeekdayFriday: "Fri", CalendarWeekdaySaturday: "Sat",
	},
	LanguageChinese: {
		ActionCut: "剪切", ActionCopy: "复制", ActionPaste: "粘贴", ActionSelectAll: "全选", ActionClose: "关闭",
		DatePickerChooseDate: "选择日期", DatePickerChooseDateRange: "选择日期范围",
		DateFieldYear: "年", DateFieldMonth: "月", DateFieldDay: "日",
		ColorPickerPickColor: "选择颜色", ColorPickerWheel: "色轮", ColorPickerAlpha: "透明度",
		ColorPickerHue: "色相", ColorPickerArea: "颜色区域", ColorPickerColor: "颜色",
		TitleBarMinimize: "最小化", TitleBarRestore: "还原", TitleBarMaximize: "最大化", TitleBarClose: "关闭",
		ComboBoxNoResults:     "无匹配结果",
		CollectionNoItems:     "暂无项目",
		MenuNoActions:         "暂无操作",
		SidebarNoDestinations: "暂无导航项",
		SidebarGroupExpand:    "导航组。激活以展开。",
		SidebarGroupCollapse:  "导航组。激活以折叠。",
		TableNoResults:        "暂无结果",
		ChartNoData:           "暂无数据",
		GanttNoVisibleTasks:   "暂无可见任务",
		SelectPlaceholderItem: "请选择一项", SelectPlaceholderItems: "请选择多个项目",
		TableLoadingMore:   "正在加载更多行",
		PaginationLabel:    "分页",
		PaginationPrevious: "上一页", PaginationNext: "下一页",
		MenubarLabel: "菜单栏", TabsMore: "更多", TabsList: "标签页列表", TabsTab: "标签页", TabsSelectedTab: "当前标签页",
		TabsScrollLeft: "向左滚动标签页", TabsScrollRight: "向右滚动标签页", TabsScrollUp: "向上滚动标签页", TabsScrollDown: "向下滚动标签页", TabsAddTab: "添加标签页",
		TabsClose: "关闭%s", TabsPanel: "%s标签页面板",
		TreePartiallySelected: "部分选中", TreeCollapsed: "已折叠", TreeExpanded: "已展开",
		TreeChildrenNotLoaded: "子项尚未加载", TreeLoadingChildren: "正在加载子项",
		TreeLoadChildrenFailed: "子项加载失败，可激活展开控件重试",
		SpinnerLoading:         "正在加载",
		CommandPaletteTitle:    "命令面板",
		CommandPaletteSearch:   "搜索命令",
		CommandPaletteEmpty:    "没有匹配的命令",
		TableSelectAllRows:     "选择所有行", TableResizeColumn: "调整%s列大小", TableSortColumn: "排序列",
		NodeGraphDescription: "节点图", NodeGraphFocused: "已聚焦节点",
		ChartBarDefaultLabel: "柱状图", ChartBarDescription: "%s，%d 个系列，%d 个类别",
		ChartLineDefaultLabel: "折线图", ChartLineDescription: "%s，%d 个系列",
		ChartCandlestickDefaultLabel: "蜡烛图", ChartCandlestickDescription: "%s，%d 根蜡烛",
		ChartPieDefaultLabel: "饼图", ChartPieDescriptionOne: "%s，%d 个扇区", ChartPieDescriptionMany: "%s，%d 个扇区",
		ChartCandlestickOpen: "开盘  %s", ChartCandlestickHigh: "最高  %s", ChartCandlestickLow: "最低  %s", ChartCandlestickClose: "收盘  %s",
		ChartMoreSeries: "+%d 个系列", GanttProgress: "进度  %.0f%%",
		GanttTimeMonthDay: "%[1]s%[2]d日", GanttTimeMonthDayTime: "%[1]s%[2]d日 %[3]s",
		GanttTimeMonthYear: "%04[2]d年%[1]s", GanttTimeYear: "%04d",
		ProgressDefault: "进度", ProgressInProgress: "%s 进行中", ProgressValue: "%s %s",
		MeterDefault: "仪表", SliderDefault: "滑块", SliderValue: "%s %s",
		SliderMinimum: "%s 最小值 %s", SliderMaximum: "%s 最大值 %s", ToolbarDefault: "工具栏",
		WorkbenchNextTab: "下一个标签页", WorkbenchPreviousTab: "上一个标签页", WorkbenchCloseTab: "关闭标签页",
		WorkbenchToggleSidebar: "切换侧边栏", WorkbenchTogglePanel: "切换底部面板", WorkbenchToggleStatus: "切换状态栏",
		ShortcutLeft: "左", ShortcutRight: "右", ShortcutUp: "上", ShortcutDown: "下",
		ShortcutReturn: "返回", ShortcutEnter: "回车", ShortcutEscape: "Esc", ShortcutHome: "首页", ShortcutEnd: "末尾",
		ShortcutBackspace: "退格", ShortcutDelete: "删除", ShortcutPageUp: "上页", ShortcutPageDown: "下页",
		ShortcutTab: "Tab", ShortcutSpace: "空格",
		CalendarMonthJanuary: "1月", CalendarMonthFebruary: "2月", CalendarMonthMarch: "3月", CalendarMonthApril: "4月",
		CalendarMonthMay: "5月", CalendarMonthJune: "6月", CalendarMonthJuly: "7月", CalendarMonthAugust: "8月",
		CalendarMonthSeptember: "9月", CalendarMonthOctober: "10月", CalendarMonthNovember: "11月", CalendarMonthDecember: "12月",
		CalendarMonthFullJanuary: "1月", CalendarMonthFullFebruary: "2月", CalendarMonthFullMarch: "3月", CalendarMonthFullApril: "4月",
		CalendarMonthFullMay: "5月", CalendarMonthFullJune: "6月", CalendarMonthFullJuly: "7月", CalendarMonthFullAugust: "8月",
		CalendarMonthFullSeptember: "9月", CalendarMonthFullOctober: "10月", CalendarMonthFullNovember: "11月", CalendarMonthFullDecember: "12月",
		CalendarMonthYear:     "%04[2]d年%[1]s",
		CalendarWeekdaySunday: "日", CalendarWeekdayMonday: "一", CalendarWeekdayTuesday: "二", CalendarWeekdayWednesday: "三",
		CalendarWeekdayThursday: "四", CalendarWeekdayFriday: "五", CalendarWeekdaySaturday: "六",
	},
}}
