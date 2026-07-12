package ui_test

import (
	"bytes"
	"testing"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide/lucide"
)

func TestNewIconDataRejectsInvalidIconVG(t *testing.T) {
	if _, err := ui.NewIconData([]byte("not iconvg")); err == nil {
		t.Fatal("invalid IconVG data was accepted")
	}
}

func TestNewIconDataCopiesInput(t *testing.T) {
	source := append([]byte(nil), lucide.Search...)
	data, err := ui.NewIconData(source)
	if err != nil {
		t.Fatal(err)
	}
	source[0] = 0
	if !bytes.Equal(data, lucide.Search) {
		t.Fatal("icon data changed with its source buffer")
	}
}

func TestMustIconDataPanicsForInvalidIconVG(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustIconData did not panic")
		}
	}()
	ui.MustIconData([]byte("not iconvg"))
}
