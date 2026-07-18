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
`ui.Application.Run`。

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
		}),
		ui.Locale(ui.LanguageEnglish),
	)
}
```

FlowUI 内置英文和中文组件文案。`ui.LanguageAuto` 会根据系统语言选择
默认语言。由应用管理窗口时，可以通过 `Application.SetTheme` 和
`Application.SetLanguage` 在运行期间切换指定窗口。

## 示例

可运行示例位于 [`examples/`](examples/)：

```bash
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
