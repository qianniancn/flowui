# FlowUI Explorer

`explorer` 将 `gioui.org/x/explorer` 适配为 FlowUI 每窗口原生文件对话框。
公开 API 不暴露 Gio 窗口或上游类型。

## 使用方式

文件对话框必须在 `ui.Cmd` 收到的 `context.Context` 中调用：

```go
func openFile() ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		reader, err := explorer.ChooseFile(ctx, ".txt", ".md")
		if errors.Is(err, explorer.ErrCanceled) {
			send(OpenCanceled{})
			return nil
		}
		if err != nil {
			return err
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		send(FileOpened{Data: data})
		return nil
	})
}
```

不要使用 `context.Background()` 调用文件对话框，它没有绑定 FlowUI 窗口，
会返回 `ErrUnavailable`。

## API

- `ChooseFile(ctx, extensions...)`：选择一个文件，调用者负责关闭 reader。
- `ChooseFiles(ctx, extensions...)`：选择多个文件，调用者负责关闭所有 reader。
- `CreateFile(ctx, name)`：选择保存位置，调用者必须关闭 writer；部分平台在关闭时才提交文件。
- `ErrCanceled`：用户关闭或取消了对话框，是正常交互结果。
- `ErrUnavailable`：当前 context、平台或桌面会话不支持该操作。

`ChooseFiles` 的平台能力取决于上游实现；例如当前 macOS 实现不支持多选。
原生对话框 API 不支持通过 context 强制关闭已经显示的系统窗口。窗口退出会取消
FlowUI Command，并在对话框最终返回时关闭迟到的文件句柄。

完整示例位于 [`examples/file_dialogs`](../examples/file_dialogs)。
