# FlowUI

[English](README.md)

FlowUI 是一个基于 [Gio](https://gioui.org/) 构建的 Go 桌面 UI 框架。
它通过公开的 `ui` 包提供 MVU 应用循环、消息驱动的状态更新、受控组件、
主题配置、布局原语、弹出界面和测试工具。

应用统一使用：

```go
import "github.com/qianniancn/FlowUI/ui"
```

应用状态放在自己的模型中。组件负责焦点、悬停、选择和动画进度等交互状态，
以及必要的派生绘制状态。

## 环境要求

- Go 1.26.2 或更高版本
- Gio 支持的桌面平台

## 安装

```bash
go get github.com/qianniancn/FlowUI/ui
```

## 快速开始

在一个 Go module 中创建 `main.go`：

```go
package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Count int
}

type Msg struct {
	Delta int
}

func Update(model *Model, msg Msg) {
	model.Count += msg.Delta
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.Text(fmt.Sprintf("Count: %d", model.Count)).Size(20),
		ui.Row(
			ui.Button("decrease", ui.Text("-1")).OnClick(func() {
				send(Msg{Delta: -1})
			}),
			ui.Button("increase", ui.Text("+1")).OnClick(func() {
				send(Msg{Delta: 1})
			}),
		).Gap(8),
	).Gap(12)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Counter"),
		ui.Size(640, 480),
	)
}
```

运行：

```bash
go run .
```

## 应用模型

根据应用需求选择入口：

- `ui.Run`：同步更新。
- `ui.RunCmd`：`Update` 返回异步 command。
- `ui.RunWithSubscriptions`：使用 command 和长期订阅。
- `ui.RunProgram`：需要启动任务或原生窗口状态消息时使用。

Command 在事件循环之外运行。捕获不可变的值，并通过 `ui.Send` 发送结果；
不要在 command 中保存模型指针或 `ui.Context`。

搜索、预览等只需要最新结果的任务使用 `ui.LatestCmd`，不再需要时使用
`ui.CancelLatestCmd` 取消对应任务。

多窗口应用可以使用 `ui.NewWindow`、`ui.NewWindowCmd` 或
`ui.NewProgramWindow` 创建 `ui.WindowSpec`，再交给 `ui.RunWindows` 或
`ui.Application.Run`。`NewWindow` 系列的初始化函数会为每个窗口实例执行，
并且必须返回彼此独立的模型状态。

## 组件

公共组件 API 覆盖桌面应用中常见的几类工作：

- **内容：** `Text`、`Label`、`Description`、`Image`、`Avatar`、`Badge`。
- **表单：** `Input`、`TextArea`、`Checkbox`、`Switch`、`RadioGroup`、
  `Select`、`ComboBox`、日期选择器和颜色选择器。
- **导航与弹出界面：** `Tabs`、`Sidebar`、`Tree`、`Menu`、`Menubar`、
  `Dropdown`、`Popover`、`Tooltip`、`Modal`、`AlertDialog`、`Toast`。
- **数据与反馈：** `Table`、`Pagination`、`Slider`、`ProgressBar`、
  `ProgressCircle`、`Meter`、`Spinner`，以及折线图、柱状图、饼图和
  K 线图。
- **布局：** `Surface`、`Card`、`Box`、`Row`、`Column`、`Grid`、`Scroll`、
  `SplitPane`、`Stack`、`Overlay` 和相关布局原语。

组件样式由 FlowUI 的主题配置统一管理，并在适合桌面控件的范围内参考
HeroUI 的变体、尺寸、状态和交互反馈。

## 主题与语言

可以从 `ui.DefaultTheme` 或 `ui.DarkTheme` 开始。使用 `WithTheme` 替换
主题，或使用 `CustomizeTheme` 修改指定配置。在上面的应用中可以这样写：

```go
package main

import (
	"image/color"

	"github.com/qianniancn/FlowUI/ui"
)

func main() {
	ui.Run(Model{}, Update, View,
		ui.CustomizeTheme(func(theme *ui.Theme) {
			theme.Palette.Accent = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}
			theme.Components.Button.Radius = 8
			theme.Components.Button.BorderWidth = 2
		}),
		ui.Locale(ui.LanguageEnglish),
	)
}
```

应用级主题覆盖作用于应用或窗口。组件实例使用可复用的 `Style` 快照。
默认样式、变体、尺寸、继承的 `StyleScope` 和实例样式按顺序级联：

```go
primary := ui.Background(ui.TokenAccent).
	TextColor(ui.TokenAccentForeground).
	Radius(8).
	Cursor(ui.CursorPointer).
	BoxShadow(0, 6, 18, 0, ui.RGBA(0x00000030)).
	When(ui.Hovered, ui.Background(ui.TokenAccentHover)).
	When(ui.Pressed, ui.Background(ui.TokenAccentPressed).
		Scale(0.96, 0.96))

ui.StyleScope(
	ui.FontSize(14),
	ui.Button("save", ui.Text("保存")).Style(primary),
)
```

`When` 可以使用 `Hovered`、`Pressed`、`Focused`、`Disabled`、`Selected` 和
`Invalid` 等运行时状态。`StyleScope` 作用于后代组件，实例 `Style` 拥有最终
优先级。主题颜色 Token 会在布局时解析，因此同一份 Style 可以跟随明暗主题。
MVU 模型里的布尔值使用 `When(ui.If(model.Highlighted), ...)` 接入同一条路径；
模型变化后由 View 重新构建声明。

`RGB` 接收 `0xRRGGBB`，`RGBA` 接收 `0xRRGGBBAA`，`Color` 接收标准库
`color.Color`，`WithAlpha` 可调整具体颜色或主题 Token 的透明度。几何能力包括
固定/最小/最大尺寸、填充、margin、padding、overflow 裁剪、cursor 和宽高比：

```go
square := ui.Width(40).AspectRatio(1).Background(ui.RGBA(0x9333eacc))
```

根属性始终作用于组件外层盒子；复合组件内部使用具名 Part。内置 Part 包括
`PartContent`、`PartLabel`、`PartDescription`、`PartIcon`、`PartTrack`、
`PartFill`、`PartThumb`、`PartIndicator`、`PartPanel`、`PartItem`、
`PartBackdrop`、`PartPlaceholder`、`PartSelection`、`PartPrefix` 和
`PartSuffix`：

```go
barStyle := ui.Background(ui.RGBA(0x111827cc)). // 组件外层
	Part(ui.PartTrack, ui.Height(6).Background(ui.TokenSurfaceRaised)).
	Part(ui.PartFill, ui.Background(ui.TokenAccent)).
	Part(ui.PartLabel, ui.TextColor(ui.TokenMutedForeground))

ui.ProgressBar("upload", 42).Label("上传").Style(barStyle)
```

Select、ComboBox 和日期控件这类复合字段使用 `PartContent` 表示字段表面；
根属性仍属于整个组件外层。

自定义组件通过 `ResolveStyle`、`ResolveStylePart`、`LayoutResolvedStyle` 和
`LayoutInteractiveResolvedStyle` 复用同一套级联与渲染器；完整写法见
`examples/custom_widgets`。

`Surface` 也使用同一套 Style API 设置实例几何和绘制属性：

```go
ui.Surface(content).Style(ui.Radius(12).
	BorderWidth(1).
	BorderColor(ui.RGB(0x9333ea)))
```

transition 需要稳定身份。交互组件已经拥有身份；每个同级的非交互 transition
组件应分别放入不同的 `ui.Key` 作用域；`Box` 也可以直接使用自己的 `.Key(...)`。

阴影几何参数属于应用级主题。每种阴影包含由近到远排列的三层，透明度为零的层
不会绘制：

```go
ui.Run(Model{}, Update, View, ui.CustomizeTheme(func(theme *ui.Theme) {
	theme.Palette.SurfaceShadow = color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xff}
	theme.Shadows.Surface.Layers = [ui.ShadowLayerCount]ui.ShadowLayerTheme{
		{OffsetY: 2, Blur: 4, Opacity: 0.65},
		{OffsetY: 7, Blur: 16, Spread: 2, Opacity: 0.4},
		{OffsetY: 16, Blur: 36, Spread: 6, Opacity: 0.3},
	}
}))
```

`Layers[0]`、`Layers[1]`、`Layers[2]` 分别表示近层、中层和远层；每层可控制
偏移、模糊、扩散和透明度。
可用配置包括 `Surface`、`Overlay`、`Menu`、`Control`、`Checkbox` 和
`SwitchThumb`。这些配置控制阴影几何参数和每层透明度；基础颜色仍由组件使用的
主题字段提供，例如 `Palette.SurfaceShadow` 和 `Components.Menu.ShadowColor`。

FlowUI 内置英文和中文组件文案。`ui.LanguageAuto` 会根据系统语言选择
默认语言。由应用管理窗口时，可以通过 `Application.SetTheme` 和
`Application.SetLanguage` 在运行期间切换指定窗口。

## 示例

可运行示例位于 [`examples/`](examples/)：

```bash
go run ./examples/components
go run ./examples/counter
go run ./examples/async
go run ./examples/tables
go run ./examples/multi_windows
```

目录中还包含表单、导航、弹出界面、图表、自定义组件、窗口标题栏和布局原语等
示例。

## 测试

`uitest` 提供确定性的帧、输入、时间和应用测试工具，用于组件和应用
测试。它不是应用运行时依赖。

```bash
go test ./...
go vet ./...
```

## 项目结构

- `ui/`：应用使用的公开包。
- `internal/components/`：组件实现。
- `internal/frame/`、`internal/state/`、`internal/theme/` 等：共享实现服务。
- `uitest/`：确定性的测试工具。
- [`docs/architecture.md`](docs/architecture.md)：依赖方向、状态归属和弹出界面行为说明。

应用应使用 `ui`，不要直接导入 `internal` 包。API 仍在演进中，示例和公开
包文档是当前用法的主要参考。

## 许可证

FlowUI 使用 [MIT 许可证](LICENSE)。

## 参与开发

开发、测试和提交规范见 [`CONTRIBUTING.md`](CONTRIBUTING.md)。
