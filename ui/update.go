package ui

// Send dispatches a message to Update. It is safe for concurrent commands to
// call Send; the messages are queued for a later frame.
type Send[Msg any] func(Msg)

// Update applies a message to the model.
type Update[M any, Msg any] func(*M, Msg)

// A Cmd returned by UpdateCmd runs in its own goroutine after UpdateCmd returns
// and may send messages back later.
//
// A command may overlap later frames. It must capture only immutable value
// snapshots prepared by UpdateCmd, must not retain or access the model pointer
// or a Context, and must report results through Send. Copy slices, maps, and
// other reference-backed model data before capturing them.
type Cmd[Msg any] func(Send[Msg])

// UpdateCmd applies a message and may return a command to run. It must finish
// all model mutation before returning; a returned Cmd must follow the Cmd
// capture rules.
type UpdateCmd[M any, Msg any] func(*M, Msg) Cmd[Msg]

// Do turns fn into a command. Do does not copy or synchronize values captured
// by fn; callers must prepare immutable snapshots before calling Do.
func Do[Msg any](fn func(Send[Msg])) Cmd[Msg] {
	if fn == nil {
		return nil
	}
	return fn
}
