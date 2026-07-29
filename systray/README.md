# FlowUI System Tray

`systray` 提供 Windows、macOS 和 Linux 的系统托盘图标与原生菜单封装。

## 当前支持

| 平台 | 实现 | 运行前提 |
| --- | --- | --- |
| Windows | Win32 `Shell_NotifyIcon`、原生 Popup Menu | 无额外依赖 |
| macOS | Cocoa `NSStatusItem`、`NSMenu` | macOS、CGo、Cocoa/Xcode 工具链 |
| Linux | D-Bus StatusNotifierItem、DBusMenu | 会话 D-Bus、StatusNotifierWatcher 和菜单宿主 |

平台代码使用构建标签选择；Android 和 iOS 不提供系统托盘实现。

## 快速开始

图标参数是图片文件的字节内容，不是文件路径。跨平台统一支持 PNG；Windows 和 Linux 会明确按 PNG 解码，因此不要依赖某个平台原生层对 JPEG 等其他格式的额外支持。目前不支持直接传入 ICO、SVG、WebP 或文件路径。

建议使用带透明背景的正方形 PNG，常规应用可从 `32x32` 或 `48x48` 开始。当前 API 每次接收一张图片，不包含 ICO 那样的多尺寸图层；系统可能根据 DPI 和桌面面板尺寸缩放图片，因此应避免尺寸过小。macOS 模板图标适合使用透明背景的单色 PNG，并通过 `SetTemplateIcon` 设置，以自动适配浅色和深色菜单栏。

下面的示例假定当前目录有 `icon.png`：

```go
package main

import (
	"log"
	"os"

	"github.com/qianniancn/flowui/systray"
)

func main() {
	icon, err := os.ReadFile("icon.png")
	if err != nil {
		log.Fatal(err)
	}

	tray := systray.New()
	tray.SetIcon(icon).SetTooltip("My App")

	menu := systray.NewMenu()
	menu.Add("显示窗口").OnClick(func() {
		// 显示应用窗口
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func() {
		tray.Destroy()
		os.Exit(0)
	})

	tray.SetMenu(menu)
	go tray.Run()

	select {
	case <-tray.Ready():
		log.Println("系统托盘已就绪")
	case err := <-tray.Errors():
		log.Printf("系统托盘启动失败: %v", err)
		return
	}

	for {
		select {
		case err := <-tray.Errors():
			log.Printf("系统托盘运行错误: %v", err)
		case <-tray.Done():
			return
		}
	}
}
```

仓库中的完整 FlowUI 示例位于 [`examples/systray_ui`](../examples/systray_ui)。Windows 发布版本建议用 `gogio` 打包；它会自动嵌入 DPI manifest，使用高于 100% 的显示缩放时，原生托盘菜单仍能保持清晰。

```powershell
go run gioui.org/cmd/gogio@v0.10.0 -target windows -arch amd64 -o flowui-systray.exe ./examples/systray_ui
```

FlowUI 默认在最后一个窗口关闭后退出。始终使用托盘的应用应在 `Run` 前启用保活，并只从托盘退出命令结束应用：

```go
application := ui.NewApplication()
application.SetKeepAlive(true)

// 托盘退出回调中调用。
application.Quit()
```

如果托盘由用户选择是否启用，不要在启动时调用 `SetKeepAlive(true)`。启用托盘时先开启保活，再创建并运行托盘；停用时先销毁托盘，再关闭保活：

```go
application.SetKeepAlive(true)
tray := systray.New()
go tray.Run()

select {
case <-tray.Ready():
	// 此时才应把界面状态标记为“托盘已启用”。
case err := <-tray.Errors():
	tray.Destroy()
	application.SetKeepAlive(false)
	log.Printf("托盘启动失败: %v", err)
}

// 用户停用托盘时调用。当前窗口仍保持打开，之后关闭窗口会正常退出。
tray.Destroy()
application.SetKeepAlive(false)
```

`SystemTray` 是一次性对象；`Destroy` 后重新启用托盘时必须再次调用 `systray.New()`。完整的可启用/停用控制器见上述示例。

窗口关闭后可用 `application.Open(windowSpec)` 重新创建窗口。需要保留 MVU Model 时，在创建窗口时添加 `ui.RetainModelOnClose()`：

```go
window := ui.NewWindow("main", program,
	ui.RetainModelOnClose(),
)
```

该选项通过赋值保留 Model，不会进行深拷贝。窗口重开时不会再次执行 `Init` 和初始命令，但会根据保留的 Model 重新建立订阅。组件内部的临时交互状态仍随原生窗口销毁；应保存的业务状态需要放在 Model 中。完整的控制器写法见上述示例。

需要让 FlowUI 发起的关闭请求支持“取消”或“关闭到托盘”时，可以为窗口配置关闭决策：

```go
window := ui.NewWindow("main", program,
	ui.OnWindowCloseRequest(func() ui.WindowCloseDecision {
		if hasUnsavedChanges {
			return ui.WindowCloseCancel
		}
		if trayEnabled {
			return ui.WindowCloseKeepAlive
		}
		return ui.WindowCloseProceed
	}),
	ui.RetainModelOnClose(),
)

// 应用内的关闭命令。
application.RequestClose("main")
```

`WindowCloseKeepAlive` 会销毁当前原生窗口并自动开启应用保活；它不是隐藏并保留同一个原生窗口句柄。之后用 `application.Open(window)` 重建窗口，用 `application.Quit()` 从托盘彻底退出。`WindowTitleBar` 的默认关闭按钮会调用这个生命周期入口；组件显式设置 `OnClose` 后则由该回调自行处理。

`application.Close`、`CloseAll` 和 `Quit` 是强制关闭路径，不会调用关闭决策，因此托盘的“退出”不会被取消。

Gio `v0.10.1` 没有公开可取消的原生关闭请求事件。系统标题栏关闭按钮、`Alt+F4` 和窗口管理器关闭命令可能直接销毁窗口，`OnWindowCloseRequest` 无法拦截它们。使用原生标题栏的托盘应用应在托盘启用期间调用 `SetKeepAlive(true)`；这样原生窗口关闭后进程仍会保留，但不能取消该次关闭。

## 菜单

```go
menu := systray.NewMenu()

menu.Add("普通项目").OnClick(func() {})
menu.AddSeparator()
menu.AddCheckbox("启用功能", true).OnClick(func() {})

// 连续的 Radio 项会组成一个单选组；分隔符或其他类型会结束当前分组。
menu.AddRadio("选项 1", true)
menu.AddRadio("选项 2", false)

submenu := menu.AddSubmenu("更多选项")
submenu.Add("子项").OnClick(func() {})
```

Checkbox 和 Radio 项在 `Click` 时会自动更新选中状态；Radio 同组的其他项目会自动取消选中。

## 动态更新

```go
tray.SetTooltip("同步中...")
tray.Hide()
tray.Show()

item := menu.AddCheckbox("状态", false)
item.SetLabel("已完成").SetEnabled(true).SetChecked(false)
menu.Update()
```

菜单动态更新由平台实现决定。Windows 和 macOS 会更新已创建的原生菜单；Linux 当前 `Menu.Update` 和菜单项原生 setter 受限，建议在 `SetMenu` 前完成菜单配置。

## API 概览

### `SystemTray`

- `New() *SystemTray`
- `SetIcon([]byte) *SystemTray`
- `SetTooltip(string) *SystemTray`
- `SetLabel(string) *SystemTray`（macOS 状态栏标签）
- `SetMenu(*Menu) *SystemTray`
- `OnClick(func()) *SystemTray`
- `OnRightClick(func()) *SystemTray`
- `OnDoubleClick(func()) *SystemTray`
- `OnMouseEnter(func()) *SystemTray`（平台相关）
- `OnMouseLeave(func()) *SystemTray`（平台相关）
- `Run()`
- `Ready() <-chan struct{}`
- `Errors() <-chan error`
- `Done() <-chan struct{}`
- `Show()` / `Hide()`
- `ShowMenu()`
- `Destroy()`
- `ID() uint`
- `SetTemplateIcon([]byte) *SystemTray`（仅 macOS，适配浅色/深色外观）

设置 `OnRightClick` 后，自定义回调会替代自动菜单弹出；需要同时显示菜单时，在回调中调用 `tray.ShowMenu()`。

`Run` 应在完成图标、菜单和回调配置后调用。`Ready` 在原生托盘注册成功后关闭；启动失败和运行期错误通过 `Errors` 报告；`Done` 在托盘销毁或无法继续运行时关闭。`Errors` 在对象生命周期内保持打开，因此应结合 `Done` 结束监听。Windows/Linux 实现会运行消息或 D-Bus 循环；macOS 的 `Run` 不会单独启动 Cocoa 事件循环，必须嵌入已有的 macOS GUI 事件循环。

### `Menu`

- `NewMenu() *Menu`
- `Add(string) *MenuItem`
- `AddSeparator() *MenuItem`
- `AddCheckbox(string, bool) *MenuItem`
- `AddRadio(string, bool) *MenuItem`
- `AddSubmenu(string) *Menu`
- `Items() []*MenuItem`
- `Update()` / `Destroy()`

### `MenuItem`

- `OnClick(func()) *MenuItem`
- `SetLabel` / `SetTooltip`
- `SetDisabled` / `SetEnabled`
- `SetChecked` / `SetHidden`
- `Label` / `Tooltip`
- `IsDisabled` / `IsEnabled` / `IsChecked` / `IsHidden`
- `Type()` / `Submenu()` / `ID()`
- `Click()`（手动触发回调）

可用的 `MenuItemType`：`MenuItemText`、`MenuItemSeparator`、`MenuItemCheckbox`、`MenuItemRadio` 和 `MenuItemSubmenu`。

## 运行示例

```bash
go run ./examples/systray_ui
```

Linux 需要当前桌面会话提供 StatusNotifierWatcher；没有托盘宿主时，`Errors` 会报告注册失败并关闭 `Done`。macOS 独立示例需要由 Cocoa 事件循环托管。

## 包结构

```text
systray/
├── systray.go
├── menu.go
├── menuitem.go
├── systray_windows.go / menu_windows.go
├── systray_darwin.go / menu_darwin.go
└── systray_linux.go / menu_linux.go

internal/sys/
├── windows/       # Win32 封装
├── darwin/        # Cocoa/CGo 封装
└── linux/dbus/    # StatusNotifierItem 与 DBusMenu
```

## 线程与生命周期

配置字段和菜单项状态由锁保护，`Menu.Items()` 返回菜单项指针的浅拷贝。不要在一个 goroutine 添加菜单项的同时从另一个 goroutine 更新或遍历同一菜单。原生平台调用应在完成初始化后进行；回调可能在平台事件 goroutine 中执行，不要假设它运行在 UI 主线程。

调用 `Destroy()` 会移除托盘图标、销毁原生菜单并从全局 ID 映射中删除对象。
`SystemTray` 是一次性生命周期对象，调用 `Destroy()` 后不要再次调用 `Run()`；需要重新创建托盘时使用 `systray.New()`。

Windows 实现会把隐藏窗口、消息循环、菜单更新和销毁固定在同一个系统线程。调用方可以从 UI、命令或菜单回调 goroutine 安全调用 `SetMenu`、`Show`、`Hide` 和 `Destroy`，原生操作会被投递回托盘消息线程。
Explorer 重启后，Windows 会响应系统的 `TaskbarCreated` 消息并自动重新注册托盘图标。

macOS 的 `NSStatusItem`、`NSMenu` 和 `NSMenuItem` 操作统一在 AppKit 主线程执行。Linux 会把 PNG 解码为 StatusNotifierItem 规范要求的 ARGB `IconPixmap`，读取 `Run` 前设置的菜单，并在第一次 D-Bus `GetLayout` 请求中返回完整菜单项属性；是否显示托盘和菜单仍取决于桌面会话中的 StatusNotifierWatcher 与 DBusMenu 宿主。

## 后续工作

- 支持 ICO、SVG 等更多图标格式
- 支持菜单项图标和键盘快捷键
- 补充跨平台自动化测试与更多示例

## 参考

- [Windows Shell_NotifyIcon](https://learn.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shell_notifyiconw)
- [macOS NSStatusBar](https://developer.apple.com/documentation/appkit/nsstatusbar)
- [Linux StatusNotifierItem](https://www.freedesktop.org/wiki/Specifications/StatusNotifierItem/)
