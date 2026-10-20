package main

import (
	"strings"

	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/sidebar"
)

type entry struct {
	key, title, group, summary string
	packages, source, code     string
	importPath                 string
	demo                       func(demoState, ui.Send[Msg]) ui.Widget
}

var catalog = []entry{
	{key: "button", title: "Button · 按钮", group: "操作", packages: "button", source: "demo_actions.go",
		summary: "主要操作、次要操作、危险操作，以及尺寸、加载和禁用状态。", demo: demoButton,
	},
	{key: "togglebutton", title: "ToggleButton · 切换按钮", group: "操作", packages: "togglebutton", source: "demo_actions.go",
		summary: "带选中状态的操作，适合编辑器格式和工具开关。", demo: demoToggle,
		code: `togglebutton.New("加粗").Key("bold").Selected(m.Bold).OnChange(setBold)`},
	{key: "closebutton", title: "CloseButton · 关闭按钮", group: "操作", packages: "closebutton", source: "demo_actions.go",
		summary: "紧凑的关闭操作，带键盘焦点和无障碍标签。", demo: demoClose,
		code: `closebutton.New("dismiss").OnClick(dismiss)`},
	{key: "command", title: "Command · 命令与快捷键", group: "操作", packages: "command", source: "demo_actions.go",
		summary: "让按钮、菜单和快捷键共享同一个操作。点击演示区域后可按 Ctrl+S。", demo: demoCommand,
		code: `save := command.New("save", "保存").
    Shortcut(command.KeyShortcut("S", command.ShortcutPrimary)).OnExecute(saveFile)
command.Scope([]command.Command{save}, command.Button("save-button", save))`},
	{key: "input", title: "Input · 输入框", group: "表单", packages: "input", source: "demo_forms.go",
		summary: "单行、多行、密码、校验状态，以及前后缀和附属操作。", demo: demoInput,
		code: `input.New(m.Name).Key("name").Hint("项目名称").OnChange(setName)
input.NewTextArea(m.Notes).Key("notes").OnChange(setNotes)
input.Group(input.New(m.Site).Key("site")).Prefix(text.New("https://"))`},
	{key: "checkbox", title: "Checkbox · 复选框", group: "表单", packages: "checkbox", source: "demo_forms.go",
		summary: "多项选择、部分选中与禁用状态。", demo: demoCheckbox,
		code: `checkbox.New("接收通知").Key("notice").Checked(m.Notice).OnChange(setNotice)`},
	{key: "numericinput", title: "NumericInput · 数字输入", group: "表单", packages: "numericinput", source: "demo_controls.go",
		summary: "支持负数、小数、步进、精度和边界，编辑草稿在回车或失焦时提交。", demo: demoNumericInput,
		code: `numericinput.New(m.Amount).Key("amount").Range(-100, 100).Step(0.25).
    Precision(2).Label("金额").OnChange(setAmount)`},
	{key: "segmented", title: "Segmented · 分段选择", group: "表单", packages: "segmented", source: "demo_controls.go",
		summary: "紧凑的互斥选项，支持禁用选项与等宽布局。", demo: demoSegmented,
		code: `segmented.New([]segmented.Item{
    {Key: "day", Label: "日"}, {Key: "week", Label: "周"},
}).Key("period").SelectedKey(m.Period).OnChange(setPeriod)`},
	{key: "switchcontrol", title: "Switch · 开关", group: "表单", packages: "switchcontrol", source: "demo_forms.go",
		summary: "立即生效的布尔设置。", demo: demoSwitch,
		code: `switchcontrol.New("自动同步").Key("sync").Checked(m.Sync).OnChange(setSync)`},
	{key: "radiogroup", title: "RadioGroup · 单选组", group: "表单", packages: "radiogroup", source: "demo_forms.go",
		summary: "从一组可见选项中选择一项。", demo: demoRadio,
		code: `radiogroup.New(items).Key("platform").SelectedKey(m.Platform).OnChange(setPlatform)`},
	{key: "selectfield", title: "Select · 选择器", group: "表单", packages: "selectfield", source: "demo_forms.go",
		summary: "单选、多选以及带分组的选项列表。", demo: demoSelect,
		code: `selectfield.New(items).Key("platform").SelectedKey(m.Platform).OnChange(setPlatform)
selectfield.New(items).Key("tags").SelectedKeys(m.Tags).OnSelectionChange(setTags)`},
	{key: "combobox", title: "ComboBox · 搜索选择", group: "表单", packages: "combobox", source: "demo_forms.go",
		summary: "输入关键字筛选选项，选中值与搜索文字分别受控。", demo: demoCombo,
		code: `combobox.New(items).Key("language").SelectedKey(m.Language).
    InputValue(m.Query).OnInputChange(setQuery).OnChange(setLanguage)`},
	{key: "datepicker", title: "DatePicker · 日期", group: "表单", packages: "datepicker", source: "demo_forms.go",
		summary: "日历弹层、可编辑日期字段与日期范围，使用中文本地化。", demo: demoDate,
		code: `datepicker.New("date", m.Date).Locale(datepicker.Chinese()).OnChange(setDate)
datepicker.DateRangePicker("range", m.Range).OnChange(setRange)`},
	{key: "colorpicker", title: "ColorPicker · 颜色", group: "表单", packages: "colorpicker", source: "demo_forms.go",
		summary: "完整选择器与独立色域、色轮、色条、色值字段和色板共享颜色。", demo: demoColor,
		code: `colorpicker.New("color", m.Color).ShowRGB(true).OnChange(setColor)
colorpicker.ColorWheel("wheel", m.Color).OnChange(setColor)`},
	{key: "slider", title: "Slider · 滑块", group: "表单", packages: "slider", source: "demo_forms.go",
		summary: "连续数值、步进数值与双端范围选择。", demo: demoSlider,
		code: `slider.New("volume", m.Volume).Range(0, 100).ShowValue().OnChange(setVolume)
slider.RangeSlider("range", m.Lower, m.Upper).OnRangeChange(setRange)`},
	{key: "text", title: "Text · 文本", group: "内容", packages: "text label description", source: "demo_content.go",
		summary: "标题、正文、可选择文本，以及关联输入框的标签和辅助说明。", demo: demoText,
		code: `text.New("项目概览").Size(24).Weight(font.Bold)
label.New("项目名称").For("project")
description.New("使用易于识别的名称").For("project")`},
	{key: "icon", title: "Icon · 图标", group: "内容", packages: "icon", source: "demo_content.go",
		summary: "使用已有 Lucide 图标包，搭配文字或用于纯图标按钮。", demo: demoIcon,
		code: `icon.New(lucide.Search).Size(20)
button.New("添加").Content(icon.New(lucide.Plus)).IconOnly()`},
	{key: "image", title: "Image · 图片", group: "内容", packages: "image", source: "demo_content.go",
		summary: "展示仓库内嵌图片，比较 Cover 裁剪和 Contain 完整显示。", demo: demoImage,
		code: `image.New(paint.NewImageOp(img)).Width(480).Height(200).
    Fit(image.FitCover).Radius(8).Alt("沙漠")`},
	{key: "avatar", title: "Avatar · 头像", group: "内容", packages: "avatar", source: "demo_content.go",
		summary: "文字头像及不同尺寸，适合用户、团队和项目身份。", demo: demoAvatar,
		code: `avatar.New("FL").Size(avatar.SizeLarge)`},
	{key: "badge", title: "Badge · 徽章", group: "内容", packages: "badge", source: "demo_content.go",
		summary: "附着于触发元素的数量标记。点击通知增加未读数。", demo: demoBadge,
		code: `badge.New(button.New("通知"), "8")`},
	{key: "chip", title: "Chip · 标签", group: "内容", packages: "chip", source: "demo_content.go",
		summary: "紧凑地展示状态和分类。", demo: demoChip,
		code: `chip.New("已发布")`},
	{key: "card", title: "Card · 卡片", group: "内容", packages: "card", source: "demo_content.go",
		summary: "把标题、说明与操作组织成一个内容单元。", demo: demoCard,
		code: `card.New(text.New("FlowUI"), text.New("桌面组件库"), saveButton)`},
	{key: "code", title: "Code · 代码块", group: "内容", packages: "code", source: "demo_content.go",
		summary: "只读代码块：语法高亮、拖拽选择与复制。", demo: demoCode,
		code: `code.New("snippet", "go", source).Size(13)`},
	{key: "surface", title: "Surface · 表面", group: "内容", packages: "surface", source: "demo_content.go",
		summary: "容器表面与前景色随默认主题一起切换。", demo: demoSurface,
		code: `surface.New(content).Variant(surface.Secondary)`},
	{key: "listbox", title: "ListBox · 列表选择", group: "数据与导航", packages: "listbox", source: "demo_navigation.go",
		summary: "展示可滚动的单选和多选列表。", demo: demoList,
		code: `listbox.New(items).Key("projects").SelectedKey(m.Project).OnChange(setProject)
listbox.New(items).Key("languages").SelectedKeys(m.Languages).OnSelectionChange(setLanguages)`},
	{key: "table", title: "Table · 表格", group: "数据与导航", packages: "table", source: "demo_navigation.go",
		summary: "列宽、行选择和状态单元格。点击行查看选择反馈。", demo: demoTable,
		code: `table.New("projects", columns, rows).Border(true).
    SelectionMode(table.SelectionSingle).SelectedKey(m.Project).OnChange(setProject)`},
	{key: "virtual-list", title: "VirtualList · 虚拟列表", group: "数据与导航", source: "demo_advanced.go",
		summary: "layout.List 按需构建可见行，展示一万条记录和独立滚动区域。", demo: demoVirtualList,
		code: `layout.Box(layout.List(10000, func(index int) ui.Widget {
    return text.New(fmt.Sprintf("记录 %05d", index+1))
})).Style(style.New().FillWidth().Height(320))`},
	{key: "virtual-table", title: "VirtualTable · 虚拟表格", group: "数据与导航", packages: "table", source: "demo_advanced.go",
		summary: "一万行数据按视口请求，带受控排序和行选择。", demo: demoVirtualTable,
		code: `table.NewVirtual("records", columns, 10000, rowAt).MaxHeight(320).
    SortDescriptor(m.Sort).OnSortChange(setSort).
    SelectionMode(table.SelectionSingle).SelectedKey(m.Selected).OnChange(setSelected)`},
	{key: "tree", title: "Tree · 树", group: "数据与导航", packages: "tree", source: "demo_navigation.go",
		summary: "分层资源浏览，展开状态与选中节点由应用持有。", demo: demoTree,
		code: `tree.New("files", m.File, items).ExpandedKeys(m.Expanded).
    OnExpandedChange(setExpanded).OnChange(setFile)`},
	{key: "tree-advanced", title: "Tree · 高级操作", group: "数据与导航", packages: "tree", source: "demo_advanced.go",
		summary: "按需加载、过滤、级联勾选、F2 重命名、文件拖入目录；操作写回同一份数据。", demo: demoAdvancedTree,
		code: `tree.New("files", m.Selected, tree.ItemsFromSimple(m.Items)).
    DataVersion(m.Version).Filter(m.Query).Checkable(true).
    CheckedKeys(m.Checked).OnCheckedChange(setChecked).
    OnLoadChildrenEvent(loadChildren).OnRename(rename).CanDrop(canDrop).OnDrop(move)`},
	{key: "breadcrumb", title: "Breadcrumb · 面包屑", group: "数据与导航", packages: "breadcrumb", source: "demo_controls.go",
		summary: "点击祖先层级导航，最后一项表示当前位置，窄窗口下自动换行。", demo: demoBreadcrumb,
		code: `breadcrumb.New("path", []breadcrumb.Item{
    {Key: "home", Label: "首页"}, {Key: "docs", Label: "文档"},
}).OnChange(navigate)`},
	{key: "tabs", title: "Tabs · 标签页", group: "数据与导航", packages: "tabs", source: "demo_navigation.go",
		summary: "切换内容面板，并比较主次两种导航层级。", demo: demoTabs,
		code: `tabs.New("tabs", m.Tab, items).OnChange(setTab)`},
	{key: "pagination", title: "Pagination · 分页", group: "数据与导航", packages: "pagination", source: "demo_navigation.go",
		summary: "页码导航与当前页反馈。", demo: demoPagination,
		code: `pagination.New("pages", m.Page, 12).OnChange(setPage)`},
	{key: "sidebar", title: "Sidebar · 侧栏", group: "数据与导航", packages: "sidebar", source: "demo_navigation.go",
		summary: "分组导航与固定页头页脚；本展示的左侧目录也使用此组件。", demo: demoSidebar,
		code: `sidebar.NewSections("nav", m.Page, sections).OnChange(setPage)`},
	{key: "timeline", title: "Timeline · 时间线", group: "数据与导航", packages: "timeline", source: "demo_navigation.go",
		summary: "用事件顺序解释一个项目的进展。", demo: demoTimeline,
		code: `timeline.New(items).Mode(timeline.ModeAlternate)`},
	{key: "collapsible", title: "Collapsible · 折叠面板", group: "数据与导航", packages: "collapsible", source: "demo_navigation.go",
		summary: "独立折叠区域与成组手风琴面板。", demo: demoCollapsible,
		code: `collapsible.New("详情", content).Key("details").Expanded(m.Open).OnExpandedChange(setOpen)
collapsible.Group("faq", m.OpenKeys, items).OnExpandedChange(setOpenKeys)`},
	{key: "menu", title: "Menu · 菜单", group: "菜单与浮层", packages: "menu", source: "demo_overlays.go",
		summary: "常驻菜单与右键上下文菜单，选项具有实际动作反馈。", demo: demoMenu,
		code: `menu.New("actions", items)
menu.ContextMenu("context", trigger, menu.New("context-actions", items))`},
	{key: "dropdown", title: "Dropdown · 下拉菜单", group: "菜单与浮层", packages: "dropdown", source: "demo_overlays.go",
		summary: "触发器菜单以及带独立主操作的组合按钮。", demo: demoDropdown,
		code: `dropdown.New("more", button.New("更多"), items).OnChange(onAction)
dropdown.Button("publish", button.New("发布"), items).OnClick(publish)`},
	{key: "menubar", title: "MenuBar · 菜单栏", group: "菜单与浮层", packages: "menubar", source: "demo_overlays.go",
		summary: "桌面应用中的文件、编辑等菜单入口。", demo: demoMenubar,
		code: `menubar.New("menu", []menubar.Item{menubar.NewMenu("file", "文件", items)})`},
	{key: "modal", title: "Modal · 对话框", group: "菜单与浮层", packages: "modal", source: "demo_overlays.go",
		summary: "打开编辑对话框，体验焦点约束、遮罩与 Esc 关闭。", demo: demoModal,
		code: `modal.New("编辑项目", body).Key("edit").Open(m.Open).Open(m.Open).OnOpenChange(setOpen)`},
	{key: "commandpalette", title: "CommandPalette · 命令面板", group: "菜单与浮层", packages: "commandpalette", source: "demo_controls.go",
		summary: "搜索与键盘执行应用命令，复用 command.Command，与按钮、菜单和快捷键共享操作。", demo: demoCommandPalette,
		code: `commandpalette.New("commands", m.Open, commands).OnOpenChange(setOpen)`},
	{key: "alertdialog", title: "AlertDialog · 确认框", group: "菜单与浮层", packages: "alertdialog", source: "demo_overlays.go",
		summary: "危险操作要求明确确认。本演示只更新反馈，不删除文件。", demo: demoAlertDialog,
		code: `alertdialog.New("删除项目？", "请确认此操作").Key("delete").Open(m.Open).
    Open(m.Open).OnOpenChange(setOpen).Footer(actions)`},
	{key: "popover", title: "Popover · 弹出卡片", group: "菜单与浮层", packages: "popover", source: "demo_overlays.go",
		summary: "依附触发元素的可交互内容，点击外部或 Esc 关闭。", demo: demoPopover,
		code: `trigger := button.New("项目信息").OnClick(func() { setOpen(!m.Open) })
popover.New(trigger, content).Key("info").Open(m.Open).Open(m.Open).OnOpenChange(setOpen)`},
	{key: "tooltip", title: "Tooltip · 提示", group: "菜单与浮层", packages: "tooltip", source: "demo_overlays.go",
		summary: "悬停或键盘聚焦按钮，查看补充说明。", demo: demoTooltip,
		code: `tooltip.New("save-tip", saveButton, "保存当前文档")`},
	{key: "portal", title: "Portal · 根浮层", group: "菜单与浮层", packages: "portal", source: "demo_overlays.go",
		summary: "把自定义内容挂到窗口根浮层，按触发器位置定位，避免被容器裁剪。", demo: demoPortal,
		code: `portal.New("hint", m.Visible, trigger, content).Passive(true)`},
	{key: "alert", title: "Alert · 提醒", group: "反馈", packages: "alert", source: "demo_overlays.go",
		summary: "信息、成功、警告与错误的静态反馈。", demo: demoAlert,
		code: `alert.New("已保存", "所有更改已同步").Status(alert.Success)`},
	{key: "toast", title: "Toast · 消息", group: "反馈", packages: "toast", source: "demo_overlays.go",
		summary: "点击生成消息，支持自动消失、手动关闭与动作按钮。", demo: demoToast,
		code: `toast.New("messages", m.Messages).OnClose(removeMessage).OnAction(handleAction)`},
	{key: "progress", title: "Progress · 进度与计量", group: "反馈", packages: "progress", source: "demo_overlays.go",
		summary: "条形进度、环形进度与有界计量，共享滑块控制的数值。", demo: demoProgress,
		code: `progress.Bar("upload", m.Progress).ShowValue()
progress.Circle("circle", m.Progress)
progress.Meter("storage", m.Usage).Label("存储").ShowValue()`},
	{key: "spinner", title: "Spinner · 加载", group: "反馈", packages: "spinner", source: "demo_overlays.go",
		summary: "不同尺寸的循环加载反馈。", demo: demoSpinner,
		code: `spinner.New().Label("正在加载")`},
	{key: "skeleton", title: "Skeleton · 骨架屏", group: "反馈", packages: "skeleton", source: "demo_controls.go",
		summary: "内容加载前的文本、头像和卡片占位，动效服从主题设置。", demo: demoSkeleton,
		code: `skeleton.New().Width(180).Height(16)
skeleton.New().Circle().Height(48)
skeleton.New().Height(100).Animated(false)`},
	{key: "linechart", title: "LineChart · 折线图", group: "图表", packages: "linechart", source: "demo_charts.go",
		summary: "平滑折线、面积填充和多序列；点击数据点查看反馈。", demo: demoLine,
		code: `linechart.New("visits", series).Categories(days).Height(280).OnDataClick(onSelect)`},
	{key: "barchart", title: "BarChart · 柱状图", group: "图表", packages: "barchart", source: "demo_charts.go",
		summary: "用分组柱比较计划和实际数据。", demo: demoBar,
		code: `barchart.New("tasks", series).Categories(weeks).Height(280)`},
	{key: "piechart", title: "PieChart · 环形图", group: "图表", packages: "piechart", source: "demo_charts.go",
		summary: "查看不同平台的占比，点击扇区查看数据。", demo: demoPie,
		code: `piechart.New("platforms", data).InnerRadius(0.3).Height(300)`},
	{key: "candlestick", title: "Candlestick · K 线", group: "图表", packages: "candlestick", source: "demo_charts.go",
		summary: "开盘、收盘、最低、最高值的行情展示。", demo: demoCandle,
		code: `candlestick.New("price", []candlestick.Candle{candlestick.OHLC(20, 24, 18, 26)}).Height(280)`},
	{key: "heatmap", title: "Heatmap · 热力图", group: "图表", packages: "heatmap", source: "demo_charts.go",
		summary: "以日历形式展示一段时间内的活动强度。", demo: demoHeatmap,
		code: `heatmap.New("activity", start, end, values).Height(240)`},
	{key: "gantt", title: "Gantt · 甘特图", group: "图表", packages: "gantt", source: "demo_charts.go",
		summary: "按时间轴展示项目任务，点击任务查看选择。", demo: demoGantt,
		code: `gantt.New("schedule", tasks).Height(320).Legend(true)`},
	{key: "nodegraph", title: "NodeGraph · 节点图", group: "图表", packages: "nodegraph", source: "demo_charts.go",
		summary: "拖动节点、平移和缩放画布。节点和连线变更写回应用模型。", demo: demoGraph,
		code: `nodegraph.New("pipeline", m.Graph).Height(360).
    OnNodesChange(applyNodeChanges).OnEdgesChange(applyEdgeChanges)`},
	{key: "toolbar", title: "Toolbar · 工具栏", group: "应用布局", packages: "toolbar", source: "demo_workspace.go",
		summary: "组合按钮、切换操作和分隔线，支持键盘导航。", demo: demoToolbar,
		code: `toolbar.New(saveButton, toolbar.Separator(), boldToggle)`},
	{key: "statusbar", title: "StatusBar · 状态栏", group: "应用布局", packages: "statusbar", source: "demo_workspace.go",
		summary: "左侧显示工作状态，右侧显示辅助信息。窗口底部也使用此组件。", demo: demoStatusbar,
		code: `statusbar.New(text.New("已同步"), text.New("UTF-8 · Go"))`},
	{key: "titlebar", title: "TitleBar · 标题栏", group: "应用布局", packages: "titlebar", source: "demo_workspace.go",
		summary: "自定义窗口标题和菜单。此处关闭窗口操作已隐藏，方便安全试用。", demo: demoTitlebar,
		code: `titlebar.New("title", "FlowUI Studio", menuBar).
    ShowMinimize(false).ShowMaximize(false).ShowClose(false)`},
	{key: "panel", title: "Panel · 内容面板", group: "应用布局", packages: "panel", source: "demo_workspace.go",
		summary: "按 key 切换内容，保留非活动面板的内部状态。", demo: demoPanel,
		code: `panel.New("panels", m.Panel, items).KeepAlive(true)`},
	{key: "dock", title: "Dock · 停靠布局", group: "应用布局", packages: "dock", source: "demo_workspace.go",
		summary: "嵌套分区和可拖动分隔条，适合资源栏、编辑区与输出区。", demo: demoDock,
		code: `dock.New("dock", dock.Split("split", dock.Horizontal,
    dock.Panel("files", files), dock.Panel("editor", editor)).Ratio(0.3))`},
	{key: "workbench", title: "Workbench · 工作台", group: "应用布局", packages: "workbench", source: "demo_workspace.go",
		summary: "将标签、面板和命令绑定到同一份工作台状态；支持切换、关闭和重新打开文档。", demo: demoWorkbench,
		code: `controller := workbench.NewController(m.Workbench)
controller.BindTabs("editor", tabs.New("documents", "", tabItems))
controller.BindPanel("editor", panel.New("documents-body", "", panels))`},
	{key: "layout", title: "Layout · 布局构件", group: "应用布局", packages: "", source: "demo_workspace.go",
		summary: "行列、自动换行、网格、叠放、滚动和可拖动分割布局。", demo: demoLayout,
		code: `layout.Wrap(children...).Gap(8).LineGap(8)
layout.SplitPane(first, second).DefaultRatio(0.4)`},
	{key: "building-blocks", title: "辅助包 · 组合构件", group: "应用布局", packages: "chart disclosure navigation optionrow", source: "demo_workspace.go",
		summary: "这四个包为图表、折叠、导航和选项行提供共享协议，由上面的组件使用。", demo: demoBuildingBlocks,
		code: `// chart: 图表选择与坐标公共类型
// disclosure: 展开/收起事件
// navigation: 导航事件
// optionrow: 统一选项行布局`},
	{key: "animation", title: "Animation · 动画", group: "应用布局", source: "demo_advanced.go",
		importPath: "github.com/qianniancn/flowui/animation",
		summary:    "Tween 数值插值和 AnimateLayout 尺寸过渡，使用现有动画包。", demo: demoAnimation,
		code: `value := animation.Tween("progress", target).Duration(700*time.Millisecond).Value(ctx, gtx)
animation.AnimateLayout("panel-size", content).Duration(500*time.Millisecond)`},
}

func findEntry(key string) (entry, bool) {
	for _, item := range catalog {
		if item.key == key {
			return item, true
		}
	}
	return entry{}, false
}

func filteredSections(query string) ([]sidebar.Section, int) {
	query = strings.ToLower(strings.TrimSpace(query))
	sections := []sidebar.Section{}
	count := 0
	for _, item := range catalog {
		search := item.title + " " + item.group + " " + item.summary + " " + item.packages
		if !strings.Contains(strings.ToLower(search), query) {
			continue
		}
		if len(sections) == 0 || sections[len(sections)-1].Title != item.group {
			sections = append(sections, sidebar.Section{Title: item.group, Items: []sidebar.Item{}})
		}
		section := &sections[len(sections)-1]
		section.Items = append(section.Items, sidebar.Item{Key: item.key, Label: item.title})
		count++
	}
	return sections, count
}
