package main

import (
	"testing"
	"time"

	"github.com/qianniancn/FlowUI/examples/modules/counter"
	"github.com/qianniancn/FlowUI/uitest"
)

func TestChildCommandReturnsThroughParentUpdate(t *testing.T) {
	harness := uitest.NewAppWithConfig(uitest.AppConfig[Model, Msg]{Update: Update})
	defer func() {
		if err := harness.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}()

	harness.Send(CounterMsg{Value: counter.Load{}})
	loading := harness.Frame()
	if !loading.Counter.Loading || loading.Counter.Count != 0 {
		t.Fatalf("loading model = %#v", loading.Counter)
	}
	if !harness.Wait(time.Second) {
		t.Fatal("child command did not invalidate the parent")
	}
	loaded := harness.Frame()
	if loaded.Counter.Loading || loaded.Counter.Count != 10 {
		t.Fatalf("loaded model = %#v", loaded.Counter)
	}

	harness.Send(CounterMsg{Value: counter.Increment{}})
	if got := harness.Frame().Counter.Count; got != 11 {
		t.Fatalf("incremented count = %d, want 11", got)
	}
}
