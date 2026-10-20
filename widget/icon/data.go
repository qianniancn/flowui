package icon

import "gioui.org/widget"

// NewData validates and copies encoded IconVG bytes so the result is safe to
// mutate and reuse as an icon source.
func NewData(source []byte) ([]byte, error) {
	data := append([]byte(nil), source...)
	if _, err := widget.NewIcon(data); err != nil {
		return nil, err
	}
	return data, nil
}

// MustData is like NewData but panics when source is not valid IconVG data.
// Use it for icon literals that are known good at build time.
func MustData(source []byte) []byte {
	data, err := NewData(source)
	if err != nil {
		panic(err)
	}
	return data
}
