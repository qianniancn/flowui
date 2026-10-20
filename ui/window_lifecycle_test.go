package ui

import (
	"fmt"
	"strings"
	"testing"

	"gioui.org/app"
	"github.com/qianniancn/flowui/internal/runtime"
	flowwindow "github.com/qianniancn/flowui/window"
)

type recordingClosingEvent struct {
	aborted bool
}

func (e *recordingClosingEvent) Abort()           { e.aborted = true }
func (e *recordingClosingEvent) ImplementsEvent() {}

func TestHandleClosingRunsNativeCloseHandler(t *testing.T) {
	w := new(app.Window)
	called := 0
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision {
			called++
			return flowwindow.CloseProceed
		},
	}

	a.handleClosing("main", w, &app.ClosingEvent{})
	if called != 1 {
		t.Fatalf("close handler calls = %d, want 1", called)
	}
}

func TestHandleClosingSkipsApprovedAndForcedRequests(t *testing.T) {
	w := new(app.Window)
	called := 0
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision {
			called++
			return flowwindow.CloseProceed
		},
	}

	a.windows.markCloseIntent(w, closeIntentApproved)
	a.handleClosing("main", w, &app.ClosingEvent{})
	a.handleClosing("main", w, &app.ClosingEvent{})
	if called != 1 {
		t.Fatalf("close handler calls after approval = %d, want 1", called)
	}

	a.windows.markCloseIntent(w, closeIntentForced)
	a.handleClosing("main", w, &app.ClosingEvent{})
	if called != 1 {
		t.Fatalf("close handler calls after force close = %d, want 1", called)
	}
}

func TestHandleClosingKeepAliveEnablesApplicationKeepAlive(t *testing.T) {
	w := new(app.Window)
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision {
			return flowwindow.CloseKeepAlive
		},
	}

	a.handleClosing("main", w, &app.ClosingEvent{})
	if !a.windows.keepAlive {
		t.Fatal("keep-alive was not enabled")
	}
}

func TestRunRejectsCloseRequestMessageWithoutApplication(t *testing.T) {
	defer func() {
		value := fmt.Sprint(recover())
		if !strings.Contains(value, "CloseRequestMessage") {
			t.Fatalf("ui.Run panic = %q, want CloseRequestMessage guidance", value)
		}
	}()
	program := NewProgram(0,
		func(*int, struct{}) Cmd[struct{}] { return nil },
		func(*Context, int, Send[struct{}]) Widget { return nil },
	)
	program.CloseRequestMessage = func() struct{} { return struct{}{} }
	Run(program)
}

func TestCloseMarkersAreOneShot(t *testing.T) {
	w := new(app.Window)
	var windows windowSet

	windows.markCloseIntent(w, closeIntentForced)
	if _, ok := windows.consumeCloseIntent(w); !ok {
		t.Fatal("close intent was not consumed")
	}
	if _, ok := windows.consumeCloseIntent(w); ok {
		t.Fatal("close intent was not consumed exactly once")
	}
}

func TestHandleClosingIgnoresRequestsDuringShutdown(t *testing.T) {
	w := new(app.Window)
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision { return flowwindow.CloseCancel },
	}
	a.windows.quitting = true

	event := new(recordingClosingEvent)
	if handled := a.handleClosing("main", w, event); !handled {
		t.Fatal("shutdown request was not reported as handled")
	}
	if event.aborted {
		t.Fatal("shutdown request was aborted")
	}
}

func TestHandleClosingCancelAbortsNativeClose(t *testing.T) {
	w := new(app.Window)
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision { return flowwindow.CloseCancel },
	}

	event := new(recordingClosingEvent)
	if handled := a.handleClosing("main", w, event); !handled {
		t.Fatal("cancel decision was not reported as handled")
	}
	if !event.aborted {
		t.Fatal("cancel decision did not abort the native close")
	}
}

func TestHandleClosingUnknownDecisionAbortsNativeClose(t *testing.T) {
	w := new(app.Window)
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision { return flowwindow.CloseDecision(99) },
	}

	event := new(recordingClosingEvent)
	if handled := a.handleClosing("main", w, event); !handled {
		t.Fatal("unknown decision was not reported as handled")
	}
	if !event.aborted {
		t.Fatal("unknown decision did not abort the native close")
	}
}

func TestHandleClosingProceedDoesNotAbortNativeClose(t *testing.T) {
	w := new(app.Window)
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}
	a.windows.closeRequests = map[string]func() flowwindow.CloseDecision{
		"main": func() flowwindow.CloseDecision { return flowwindow.CloseProceed },
	}

	event := new(recordingClosingEvent)
	if handled := a.handleClosing("main", w, event); !handled {
		t.Fatal("proceed decision was not reported as handled")
	}
	if event.aborted {
		t.Fatal("proceed decision aborted the native close")
	}
}

func TestHandleClosingWithoutHandlerLeavesRequestUnhandled(t *testing.T) {
	w := new(app.Window)
	a := &Application{}
	a.windows.active = map[string]*app.Window{"main": w}

	event := new(recordingClosingEvent)
	if handled := a.handleClosing("main", w, event); handled {
		t.Fatal("missing handler was reported as handled")
	}
	if event.aborted {
		t.Fatal("missing handler aborted the native close")
	}
}

func TestNativeCloseHandlerPrefersExplicitHandler(t *testing.T) {
	event := new(recordingClosingEvent)
	sent := 0
	handler := nativeCloseHandler(
		func(runtime.ClosingEvent) bool { return true },
		func() int { return 1 },
	)
	handler(event, func(int) { sent++ })
	if event.aborted || sent != 0 {
		t.Fatalf("explicit handler fallback ran: aborted=%v sent=%d", event.aborted, sent)
	}
}

func TestNativeCloseHandlerAbortsAndSendsMessage(t *testing.T) {
	event := new(recordingClosingEvent)
	var got []int
	handler := nativeCloseHandler[int](
		func(runtime.ClosingEvent) bool { return false },
		func() int { return 7 },
	)
	handler(event, func(msg int) { got = append(got, msg) })
	if !event.aborted {
		t.Fatal("native close was not aborted")
	}
	if len(got) != 1 || got[0] != 7 {
		t.Fatalf("sent messages = %v, want [7]", got)
	}
}

func TestNativeCloseHandlerWithoutMessageLeavesCloseAlone(t *testing.T) {
	event := new(recordingClosingEvent)
	handler := nativeCloseHandler[int](nil, nil)
	handler(event, func(int) { t.Fatal("nil message hook sent a message") })
	if event.aborted {
		t.Fatal("nil message hook aborted the native close")
	}
}
