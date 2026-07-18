package runtime

import "sync"

type Queue[Msg any] struct {
	mu      sync.Mutex
	msgs    []Msg
	limit   int
	dropped int
}

func (q *Queue[Msg]) Push(msg Msg) bool {
	return q.push(msg, nil)
}

// PushOrReplace appends msg while there is room. Once full, it replaces the
// newest matching entry so high-frequency streams can retain their latest
// value without changing the ordering of ordinary messages.
func (q *Queue[Msg]) PushOrReplace(msg Msg, match func(Msg) bool) bool {
	return q.push(msg, match)
}

func (q *Queue[Msg]) push(msg Msg, match func(Msg) bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.limit <= 0 || len(q.msgs) < q.limit {
		q.msgs = append(q.msgs, msg)
		return true
	}
	if match != nil {
		for index := len(q.msgs) - 1; index >= 0; index-- {
			if match(q.msgs[index]) {
				q.msgs[index] = msg
				return true
			}
		}
	}
	q.dropped++
	return false
}

func (q *Queue[Msg]) Drain(fn func(Msg)) int {
	q.mu.Lock()
	msgs := q.msgs
	q.msgs = nil
	dropped := q.dropped
	q.dropped = 0
	q.mu.Unlock()
	for _, msg := range msgs {
		fn(msg)
	}
	return dropped
}
