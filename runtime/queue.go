package runtime

import "sync"

type Queue[Msg any] struct {
	mu   sync.Mutex
	msgs []Msg
}

func (q *Queue[Msg]) Push(msg Msg) {
	q.mu.Lock()
	q.msgs = append(q.msgs, msg)
	q.mu.Unlock()
}

func (q *Queue[Msg]) Drain(fn func(Msg)) {
	q.mu.Lock()
	msgs := q.msgs
	q.msgs = nil
	q.mu.Unlock()
	for _, msg := range msgs {
		fn(msg)
	}
}
