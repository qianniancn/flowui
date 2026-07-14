# FlowUI

[English](README.md)

FlowUI 是一个基于 Gio 构建的小型 MVU UI 框架。

应用统一导入 `github.com/qianniancn/FlowUI/ui`。包依赖方向和状态归属规则见
[架构说明](docs/architecture.md)。

它把应用状态放在 `Model` 中，由视图发送强类型消息，并让 `Update`
成为唯一修改状态的地方。FlowUI 的 API 尽量保持 Go 风格：组件是普通值，
配置通过链式方法完成，本地控件状态由 `Context` 和显式 key 管理。

## 当前特性

- MVU 架构：`Model`、`Msg`、`Update`、`View`。
- 通过 `ui.Send[Msg]` 发送强类型消息。
- 通过 `RunCmd`、`Cmd`、`Do` 支持异步副作用。
- Gio 控件状态由 `Context` 统一管理。
- 基于 key 的局部状态隔离，并在每帧结束后自动清理不再使用的状态。
- HeroUI 风格控件，支持变体、尺寸、禁用、校验、加载、焦点和基础动画状态。
- DatePicker 支持本地化文案，可跟随系统语言，内置英文和中文配置，并支持日期、
  月份、年份选择视图。
- FlowUI 自己的主题 token，覆盖颜色、字体、圆角、间距和组件样式；Gio
  material theme 只作为底层文本/编辑器桥接。
- 提供固定、弹性、响应式、滚动、层叠、网格等布局能力。
- 支持 overlay / popup 类浮层，并内置缓存软阴影绘制。
- 支持窗口标题、窗口尺寸、本地化和主题配置。

## 简单使用

```go
package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct {
	Name string
}

type Msg struct {
	Name string
}

func Update(m *Model, msg Msg) {
	m.Name = msg.Name
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI").Size(24),
				ui.Input("name", m.Name).
					Hint("Name").
					OnChange(func(text string) {
						send(Msg{Name: text})
					}),
			).Gap(12),
		).Width(320).Padding(24),
	)
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI"), ui.Size(900, 600))
}
```

普通同步更新使用 `Run`。如果 `Update` 需要启动异步任务，并在任务完成后
继续发送消息，可以使用 `RunCmd`。

## 主题

FlowUI 暴露自己的 `Theme`，应用层不需要直接按 Gio 组件写样式。可以从
`DefaultTheme` 或 `DarkTheme` 开始，再通过 `WithTheme` 传入；也可以用
`CustomizeTheme` 按 token 修改：

```go
theme := ui.DarkTheme()
theme.Palette.Accent = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}

ui.Run(Model{}, Update, View, ui.WithTheme(theme))
```

`MaterialTheme` 只用于 FlowUI 内部文本/编辑器使用的 Gio material 桥接。

## 组件

### 核心

- `Run`、`RunCmd`
- `Send`、`Update`、`UpdateCmd`、`View`
- `Cmd`、`Do`
- `Context`
- `Widget`
- `Key`

### 应用配置

- `Title`
- `Size`
- `WithTheme`
- `CustomizeTheme`
- `MaterialTheme`
- `Locale`

### 基础控件

- `Text`
- `Label`
- `Description`
- `Image`
- `Avatar`
- `Badge`
- `Button`
- `ToggleButton`
- `Input`
- `TextArea`
- `Checkbox`
- `Switch`
- `SwitchGroup`
- `RadioGroup`
- `ProgressBar`
- `ProgressCircle`
- `Meter`
- `LineChart`
- `BarChart`
- `PieChart`
- `CandlestickChart`
- `Spinner`
- `Slider`、`RangeSlider`
- `ListBox`
- `Tree`
- `Sidebar`
- `StatusBar`
- `Table`
- `Pagination`
- `Tabs`
- `Select`
- `ComboBox`
- `DatePicker`
- `Popover`
- `Tooltip`
- `Menu`、`ContextMenu`
- `Menubar`
- `Dropdown`
- `Modal`

### 布局

- `Surface`
- `Card`、`CardHeader`、`CardTitle`、`CardDescription`、`CardContent`、`CardFooter`
- `Box`、`Spacer`
- `Center`
- `Row`、`Column`
- `Expanded`、`Flexible`
- `Adaptive`
- `Wrap`
- `Scroll`、`Scrollbar`、`List`
- `SplitPane`
- `Stack`、`Stacked`、`Overlay`
- `AspectRatio`
- `Grid`、`AutoGrid`
- `Divider`、`Separator`

### 绘制能力

- `DrawShadow`
- `SurfaceShadow`
- `PopupShadow`
- `RoundedShadowCorners`
- `EllipseShadow`

### 动画能力

- `Tween`
- 兼容 ECharts/zrender 的线性、二次、三次、四次、五次、正弦、指数、圆形、弹性、回退和弹跳缓动族
- `LerpFloat`、`LerpFloat64`、`LerpColor`、`LerpPoint`、`LerpRect`

## 示例目录

示例统一放在 `examples/` 下：

- `animations`：可复用缓动、Tween 状态、数值插值、方向切换和自定义 Gio 绘制。
- `counter`：基础 MVU 状态更新。
- `async`：基于 command 的异步消息。
- `buttons`：按钮变体、加载、禁用和交互状态。
- `toggle_buttons`：对齐 HeroUI 的受控切换按钮，包含变体、尺寸、纯图标、选中和禁用状态。
- `close_buttons`：对齐 HeroUI 的关闭按钮，包含禁用、自定义图标和交互状态。
- `chips`：对齐 HeroUI 的紧凑标签，包含颜色、变体、尺寸、图标和语义状态。
- `images`：可复用图片，包含缩放、定位、尺寸、圆角裁剪、透明度和无障碍标签。
- `avatars`：对齐 HeroUI 的头像，包含尺寸、语义色、柔和变体，以及文本或图标回退内容。
- `badges`：对齐 HeroUI 的锚定标记，包含颜色、变体、尺寸、位置、文本、图标和状态圆点。
- `inputs`：对齐 HeroUI 的输入框变体、类型、状态和受控事件。
- `input_groups`：对齐 HeroUI 的组合输入框，包含单行或多行编辑器、前后缀、变体、校验状态和可交互操作。
- `textareas`：对齐 HeroUI 的多行文本框，包含变体、受控值、行数、状态、InputGroup 集成和 Surface 用法。
- `labels`：表单标签状态，以及与输入框、组合框和选择器的字段关联。
- `descriptions`：辅助文本状态、字段关联、自动换行和组件兼容性。
- `checkboxes`：对齐 HeroUI 的变体、不确定态、只读态、描述、校验和自定义指示器。
- `switches`：开关状态、尺寸、描述、label 位置、thumb 内容和开关组。
- `radio_groups`：互斥单选选项。
- `progress_bars`：确定态和不定态进度指示器。
- `progress_circles`：对齐 HeroUI 的圆形确定态与不定态进度指示器，支持尺寸和语义颜色。
- `meters`：对齐 HeroUI 的已知范围测量组件，包含标签、格式化数值、颜色、尺寸和无可见标签用法。
- `line_charts`：原生 Gio 直线、平滑、阶梯、虚线、点线和面积折线图，包含动画、交互式图例、数据窗口、标注、数据激活、自定义提示框、断点和十字准线。
- `bar_charts`：原生 Gio 纵向与横向分组或堆叠柱状图，包含动画、数值标签、交互式图例、数据窗口、标注、数据激活、自定义提示框、正负值和分类高亮。
- `pie_charts`：原生 Gio 饼图、环形图与南丁格尔玫瑰图，包含对齐 ECharts 的角度、标签、动画、交互式图例、数据激活和自定义提示框。
- `spinners`：与 HeroUI 对齐的颜色和尺寸加载指示器。
- `sliders`：受控单值、范围、纵向、禁用、步进和格式化滑块。
- `list_boxes`：单选、多选、分组、禁用 key、自定义 indicator 和可触发动作的列表框。
- `trees`：受控的层级导航，包含展开、选择、自定义内容、禁用节点和滚动。
- `sidebars`：受控桌面导航，包含分组、折叠导航轨、自定义内容、禁用目的地和键盘导航。
- `status_bars`：紧凑的桌面应用状态栏，支持左右内容、语义化 Surface 和可选强调色样式。
- `scrollbars`：对齐 HeroUI 的细型纵向与横向滚动条，使用 Gio 的轨道与滑块交互。
- `split_panes`：支持最小尺寸和键盘操作的横向、纵向与嵌套桌面分栏。
- `tables`：对齐 HeroUI 的数据表格，包含列宽拖拽、范围选择、分页、异步加载、虚拟行、自定义单元格和滚动。
- `paginations`：对齐 HeroUI 的受控分页导航，包含摘要、省略号、尺寸和禁用状态。
- `tabs`：主次变体、横向与纵向布局、禁用标签、分隔线、紧凑强调色样式和溢出滚动。
- `modals`：受控模态弹窗，包含尺寸、位置、遮罩变体和关闭行为。
- `alert_dialogs`：对齐 HeroUI 的确认弹窗，包含语义状态、受控关闭、尺寸、位置和遮罩变体。
- `popovers`：受控浮层，包含位置、箭头、关闭行为和交互内容。
- `tooltips`：对齐 HeroUI 的悬停与焦点提示，包含延时、箭头、位置和视口翻转。
- `context_menus`：用于表格行的右键和长按菜单，包含勾选、单选、禁用、危险操作和子菜单项。
- `menubars`：支持菜单协同、悬停切换、键盘导航和嵌套命令的应用菜单栏。
- `dropdowns`：对齐 HeroUI 的动作与选择下拉菜单，包含分组、丰富内容、定制内容、长按和子菜单。
- `toasts`：对齐 HeroUI 的受控通知，包含变体、操作、超时、可展开堆叠和六种位置。
- `surfaces`：语义化 Surface 层级、前景色上下文、圆角和 Surface 阴影。
- `cards`：对齐 HeroUI 的卡片变体、语义分区和组合操作。
- `selects`：单选、多选、分组、禁用选项、校验、受控打开状态和 Surface 样式。
- `comboboxes`：选项选择和过滤。
- `datepickers`：日期选择和范围限制。
- `form`：表单组合。
- `layout`：布局组件。
- `todo`：带 key 的重复 UI 和列表交互。
- `popup_shadow_example`：阴影绘制示例。

## 项目状态

FlowUI 仍在演进中。当前 API 会尽量保持小而明确，新增组件也会优先保持
强类型、MVU 和 Go 风格。
