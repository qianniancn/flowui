# FlowUI

[English](README.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/qianniancn/flowui/ui.svg)](https://pkg.go.dev/github.com/qianniancn/flowui/ui)
[![Go version](https://img.shields.io/github/go-mod/go-version/qianniancn/flowui)](go.mod)
[![License](https://img.shields.io/github/license/qianniancn/flowui)](LICENSE)

FlowUI 是一个基于 [Gio](https://gioui.org/) 构建的 Go 桌面 UI 框架。
它通过统一的公开 `ui` 包，提供类型化 MVU 应用模型、完整组件集、声明式样式、
窗口管理、异步副作用以及可重复的 UI 测试能力。

```go
import "github.com/qianniancn/flowui/ui"
```

## 核心能力

- **类型化 MVU：** 业务状态放在 Model 中，用消息描述变化，View 只负责读取和渲染。
- **桌面组件：** 覆盖表单、导航、浮层、数据展示、图表、窗口外壳和布局原语。
- **声明式样式：** 支持主题令牌、明暗主题、组件部件、运行时状态、过渡和作用域样式。
- **应用运行时：** 支持 Command、订阅、多窗口、Model 保留和本地化。
- **平台服务：** 可选包提供原生文件对话框、桌面通知和系统托盘集成。
- **易于测试：** `uitest` 提供确定性的帧、输入、时间和应用测试工具。

## 环境要求

- Go 1.26.2 或更高版本
- Gio 支持的桌面平台

## 安装

```bash
go get github.com/qianniancn/flowui/ui
```

## 快速开始

在一个 Go module 中创建 `main.go`：

```go
package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Count int
}

type Msg interface{ msg() }

type Inc struct{}
type Dec struct{}

func (Inc) msg() {}
func (Dec) msg() {}

func Update(model *Model, msg Msg) {
	switch msg.(type) {
	case Inc:
		model.Count++
	case Dec:
		model.Count--
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Column(
			ui.Text("FlowUI Counter").Size(24),
			ui.Text(fmt.Sprintf("Count: %d", model.Count)),
			ui.Row(
				ui.Button("decrement", ui.Text("-1")).OnClick(func() {
					send(Dec{})
				}),
				ui.Button("increment", ui.Text("+1")).OnClick(func() {
					send(Inc{})
				}),
			).Gap(8),
		).Gap(12),
	)
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

仓库内的同款应用位于 [`examples/counter`](examples/counter)。
需要分步讲解时，请阅读[快速开始](https://github.com/qianniancn/flowui/wiki/01-快速开始)。

## 核心模型

FlowUI 遵循三条状态与布局归属规则：

1. **Model、消息和 Update 管理业务状态。** View 读取 Model 的值并发送类型化消息，
   不直接修改捕获的状态。
2. **Style 管理组件自身的盒子和外观。** 布局容器负责测量和排列子项。
3. **Key 提供跨帧身份。** 稳定的 Key 会为重复或移动的组件保留交互与动画状态。

根据应用需要的生命周期选择入口：

| API | 适用场景 |
| --- | --- |
| `ui.Run` | 同步 `Update` 循环，大多数应用从这里开始 |
| `ui.RunCmd` / `ui.RunProgram` | 异步 Command、初始化、订阅或运行时错误处理 |
| `ui.Application` / `ui.RunWindows` | 多窗口和由应用管理的窗口生命周期 |

Command 在事件循环外执行，并通过 `ui.Send` 返回结果。订阅用于计时器、外部事件流等
长期输入。完整契约见 [MVU 与消息](https://github.com/qianniancn/flowui/wiki/03-MVU与消息)
和[命令与订阅](https://github.com/qianniancn/flowui/wiki/08-命令与订阅)。

## 组件

| 分类 | 包含的 API |
| --- | --- |
| 内容 | `Text`、`Label`、`Description`、`Image`、`Avatar`、`Badge`、`Chip` |
| 操作 | `Button`、`ButtonGroup`、`ToggleButton`、`CloseButton`、`Action` |
| 表单 | `Input`、`TextArea`、`Checkbox`、`Switch`、`RadioGroup`、`Select`、`ComboBox`、日期和颜色选择器 |
| 导航 | `Tabs`、`Sidebar`、`Tree`、`Menu`、`Menubar`、`Pagination`、`Toolbar` |
| 浮层 | `Dropdown`、`ContextMenu`、`Popover`、`Tooltip`、`Modal`、`AlertDialog`、`Toast` |
| 数据与反馈 | `Table`、`ProgressBar`、`ProgressCircle`、`Meter`、`Spinner`、`Slider`、图表、`Heatmap`、`GanttChart` |
| 布局 | `Box`、`Surface`、`Card`、`Row`、`Column`、`Grid`、`Scroll`、`SplitPane`、`Stack`、`Overlay` |

可以阅读[组件一览](https://github.com/qianniancn/flowui/wiki/06-组件一览)，
或直接运行组件总览：

```bash
go run ./examples/components
```

## 样式与主题

Style 是不可变声明，可以使用主题令牌并响应运行时状态。同一份声明会自动适配当前的
浅色或深色主题：

```go
primary := ui.Background(ui.TokenAccent).
	TextColor(ui.TokenAccentForeground).
	Radius(8).
	Cursor(ui.CursorPointer).
	When(ui.Hovered, ui.Background(ui.TokenAccentHover)).
	When(ui.Pressed, ui.Background(ui.TokenAccentPressed).Scale(0.96, 0.96))

save := ui.Button("save", ui.Text("Save")).Style(primary)
```

运行时默认使用 `ui.DefaultTheme()`。通过 `ui.WithTheme(ui.DarkTheme())` 替换主题，
或用 `ui.CustomizeTheme` 做局部调整。复合控件通过 `PartContent`、`PartTrack`、
`PartIndicator`、`PartPanel` 等命名部件提供精确样式入口。FlowUI 内置中英文组件
文案，`ui.LanguageAuto` 会跟随系统语言。

[样式与主题](https://github.com/qianniancn/flowui/wiki/05-样式与主题)介绍了优先级、
命名部件、颜色、几何和过渡动画。

## 示例

下列目录都是可以直接运行的程序：

| 示例 | 内容 |
| --- | --- |
| [`examples/counter`](examples/counter) | 最小的类型化 MVU 应用 |
| [`examples/form`](examples/form) | 表单控件与验证 |
| [`examples/async`](examples/async) | Command 和异步结果 |
| [`examples/components`](examples/components) | 组件总览 |
| [`examples/custom_widgets`](examples/custom_widgets) | 自定义复合组件和画布组件 |
| [`examples/multi_windows`](examples/multi_windows) | 由应用管理的多窗口 |
| [`examples/file_dialogs`](examples/file_dialogs) | 原生打开和保存文件对话框 |
| [`examples/notifications`](examples/notifications) | 原生桌面通知 |
| [`examples/systray_ui`](examples/systray_ui) | FlowUI 窗口与原生系统托盘 |

[`examples/`](examples/) 中还有图表、动画、菜单、浮层、布局、窗口外壳以及单个控件的
专项示例。

## 文档

| 资源 | 用途 |
| --- | --- |
| [使用教程](https://github.com/qianniancn/flowui/wiki) | 从第一个应用到高级功能的任务式教程 |
| [Package Reference](https://pkg.go.dev/github.com/qianniancn/flowui/ui) | 公开 Go API |
| [`docs/architecture.md`](docs/architecture.md) | 依赖方向、状态归属、浮层和副作用 |
| [`explorer/README.md`](explorer/README.md) | 每窗口原生打开和保存对话框 |
| [`notify/README.md`](notify/README.md) | 跨平台原生通知 |
| [`systray/README.md`](systray/README.md) | 跨平台托盘生命周期与原生菜单 |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | 开发流程和贡献规范 |

应用使用 `github.com/qianniancn/flowui/ui` 构建界面和 MVU 运行时；
`explorer`、`notify`、`systray` 是可选平台服务。应用不得直接使用 `internal`
下的包。仓库根目录有意不提供 Go package。

## 测试

在仓库根目录运行完整检查：

```bash
go test ./...
go vet ./...
```

`uitest` 用于组件和应用测试，普通应用运行时不需要依赖它。详见
[测试教程](https://github.com/qianniancn/flowui/wiki/13-测试)。

## 参与贡献

欢迎参与贡献。提交修改前请阅读 [`CONTRIBUTING.md`](CONTRIBUTING.md)。

## 许可证

FlowUI 使用 [MIT License](LICENSE)。
