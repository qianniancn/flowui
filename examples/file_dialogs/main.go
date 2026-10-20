package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/qianniancn/flowui/explorer"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

const maxFileSize = 4 << 20

type Model struct {
	Content   string
	Status    string
	Operation string
}

type Msg interface{ msg() }

type EditContent string
type OpenFile struct{}
type SaveFile struct{}

type DialogFinished struct {
	Status       string
	Content      string
	ApplyContent bool
}

func (EditContent) msg()    {}
func (OpenFile) msg()       {}
func (SaveFile) msg()       {}
func (DialogFinished) msg() {}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case EditContent:
		model.Content = string(msg)
	case OpenFile:
		if model.Operation != "" {
			return nil
		}
		model.Operation = "open"
		model.Status = "Opening native file dialog..."
		return openFile()
	case SaveFile:
		if model.Operation != "" {
			return nil
		}
		model.Operation = "save"
		model.Status = "Opening native save dialog..."
		content := model.Content
		return saveFile(content)
	case DialogFinished:
		model.Operation = ""
		model.Status = msg.Status
		if msg.ApplyContent {
			model.Content = msg.Content
		}
	}
	return nil
}

func openFile() ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		reader, err := explorer.ChooseFile(ctx, ".txt", ".md", ".json")
		if err != nil {
			return finishDialog(send, "Open", err)
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, maxFileSize+1))
		closeErr := reader.Close()
		if len(data) > maxFileSize {
			readErr = fmt.Errorf("file exceeds the %d MiB example limit", maxFileSize>>20)
		}
		if err := errors.Join(readErr, closeErr); err != nil {
			send(DialogFinished{Status: "Open failed: " + err.Error()})
			return nil
		}
		send(DialogFinished{
			Status:       fmt.Sprintf("Opened %d bytes", len(data)),
			Content:      string(data),
			ApplyContent: true,
		})
		return nil
	})
}

func saveFile(content string) ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		writer, err := explorer.CreateFile(ctx, "flowui-note.txt")
		if err != nil {
			return finishDialog(send, "Save", err)
		}
		written, writeErr := io.WriteString(writer, content)
		closeErr := writer.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			send(DialogFinished{Status: "Save failed: " + err.Error()})
			return nil
		}
		send(DialogFinished{Status: fmt.Sprintf("Saved %d bytes", written)})
		return nil
	})
}

func finishDialog(send ui.Send[Msg], operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, explorer.ErrCanceled) {
		send(DialogFinished{Status: operation + " canceled"})
		return nil
	}
	if errors.Is(err, explorer.ErrUnavailable) {
		send(DialogFinished{Status: operation + " dialog is unavailable"})
		return nil
	}
	send(DialogFinished{Status: operation + " failed: " + err.Error()})
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	busy := model.Operation != ""
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Native File Dialogs").Size(24),
				description.New(model.Status),
				input.NewTextArea(model.Content).Key("content").
					Rows(12).
					FullWidth().
					Placeholder("Document content").
					OnChange(func(value string) { send(EditContent(value)) }),
				layout.Row(
					button.New("Open file").
						Variant(button.Secondary).
						Disabled(busy).
						Loading(model.Operation == "open").
						OnClick(func() { send(OpenFile{}) }),
					button.New("Save as").
						Disabled(busy).
						Loading(model.Operation == "save").
						OnClick(func() { send(SaveFile{}) }),
				).Gap(10),
			).Gap(14),
		).Style(style.New().Width(620).Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Content: "FlowUI native file dialogs are bound to this window.",
		Status:  "Ready",
	}, Update,
		View), window.Title("FlowUI Explorer"),
		window.Size(760, 620),
		window.Theme(base.Light()),
	)
}
