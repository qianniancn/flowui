package flowui

// Send dispatches a message to Update.
type Send[Msg any] func(Msg)

// Update applies a message to the model.
type Update[M any, Msg any] func(*M, Msg)

// Cmd runs work outside Update and may send messages back later.
type Cmd[Msg any] func(Send[Msg])

// UpdateCmd applies a message and may return a command to run.
type UpdateCmd[M any, Msg any] func(*M, Msg) Cmd[Msg]

// Do turns fn into a command.
func Do[Msg any](fn func(Send[Msg])) Cmd[Msg] {
	if fn == nil {
		return nil
	}
	return fn
}
