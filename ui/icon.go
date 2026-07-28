package ui

import (
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/components/icon"
)

// NewIconData validates and copies encoded IconVG bytes. Load custom icon data
// once and reuse the returned slice across layouts.
func NewIconData(src []byte) ([]byte, error) {
	data := append([]byte(nil), src...)
	if _, err := widget.NewIcon(data); err != nil {
		return nil, err
	}
	return data, nil
}

// MustIconData is like NewIconData but panics when src is not valid IconVG data.
func MustIconData(src []byte) []byte {
	data, err := NewIconData(src)
	if err != nil {
		panic(err)
	}
	return data
}

type IconWidget = icon.Widget

// Icon creates an icon from IconVG bytes. The data must remain unchanged and
// should be reused after the icon is first laid out.
func Icon(data []byte) IconWidget {
	return icon.New(data)
}
