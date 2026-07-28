# FlowUI Notify

`notify` 为 Windows、macOS 和 Linux 提供原生桌面通知，并隐藏
`gioui.org/x/notify` 的不稳定类型。

## 快速开始

```go
notification, err := notify.Push("Build complete", "The application is ready.")
if err != nil {
	return err
}

if notification.Cancelable() {
	defer notification.Cancel()
}
```

需要复用后端或配置 Windows 通知图标时创建 `Notifier`：

```go
notifier, err := notify.New()
if err != nil {
	return err
}
if notifier.SupportsIcon() {
	_ = notifier.SetIcon("app.png")
}
notification, err := notifier.Push("FlowUI", "Notification body")
```

## 平台差异

| 能力 | Windows | macOS | Linux |
| --- | --- | --- | --- |
| 标题与正文 | 支持 | 支持 | 支持 |
| 投递后取消 | 不支持 | 支持 | 支持 |
| `SetIcon` | 支持文件路径 | 不支持 | 不支持 |

Linux 需要可用的会话 D-Bus 通知服务。macOS 使用系统通知中心。
当前上游没有可跨平台统一的点击回调、动作按钮或进度通知，因此 FlowUI 暂不暴露这些能力。

- `ErrUnavailable`：当前平台、权限或桌面服务导致通知无法初始化或投递；
  后端的具体错误仍会保留，可继续使用 `errors.Is` / `errors.As` 检查。
- `ErrUnsupported`：平台不支持指定操作，例如 Windows 上取消已投递通知。

完整示例位于 [`examples/notifications`](../examples/notifications)。
