package runtime

import (
	"context"
	"fmt"
	"time"
)

// Subscription is long-lived asynchronous work identified by Key.
type Subscription[Msg any] struct {
	Key string
	Run Cmd[Msg]
}

// Subscriptions derives the active subscription set from the current model.
type Subscriptions[M any, Msg any] func(M) []Subscription[Msg]

type subscriptionToken struct {
	key        string
	generation uint64
}

type activeSubscription struct {
	token     subscriptionToken
	cancel    context.CancelFunc
	done      <-chan struct{}
	accepting bool
	stopping  bool
	deadline  time.Time
	timer     *time.Timer
}

type subscriptionSet[Msg any] struct {
	active          map[string]*activeSubscription
	seen            map[string]struct{}
	nextGeneration  uint64
	stopGracePeriod time.Duration
}

const defaultSubscriptionStopGracePeriod = 250 * time.Millisecond

func (s *subscriptionSet[Msg]) reconcile(
	root context.Context,
	desired []Subscription[Msg],
	group *effectGroup,
	send func(subscriptionToken, Msg),
	report func(error),
	wake func(),
) {
	if s.seen == nil {
		s.seen = make(map[string]struct{}, len(desired))
	} else {
		clear(s.seen)
	}
	// Validate the complete desired set before changing the active set.
	for _, subscription := range desired {
		if subscription.Key == "" {
			panic("flowui: empty subscription key")
		}
		if subscription.Run == nil {
			panic(fmt.Sprintf("flowui: nil subscription %q", subscription.Key))
		}
		if _, duplicate := s.seen[subscription.Key]; duplicate {
			panic(fmt.Sprintf("flowui: duplicate subscription key %q", subscription.Key))
		}
		s.seen[subscription.Key] = struct{}{}
	}
	// Stop removed work before replacements start. Completion and timeout wake
	// the event loop; no event-thread blocking is required.
	for key, active := range s.active {
		if _, keep := s.seen[key]; keep {
			continue
		}
		s.stop(active, wake)
	}
	now := time.Now()
	// Sweep finished or grace-expired stopping entries. Only the same key is
	// blocked from restarting while still stopping; other keys may start.
	for key, active := range s.active {
		if !active.stopping {
			continue
		}
		if effectDone(active.done) || !now.Before(active.deadline) {
			if active.timer != nil {
				active.timer.Stop()
			}
			delete(s.active, key)
		}
	}
	for _, subscription := range desired {
		if root.Err() != nil {
			return
		}
		if active, ok := s.active[subscription.Key]; ok {
			// Running or still in stop grace for this key — do not start another.
			_ = active
			continue
		}
		if s.active == nil {
			s.active = make(map[string]*activeSubscription)
		}
		ctx, cancel := context.WithCancel(root)
		s.nextGeneration++
		token := subscriptionToken{key: subscription.Key, generation: s.nextGeneration}
		active := &activeSubscription{
			token:     token,
			cancel:    cancel,
			accepting: true,
		}
		s.active[subscription.Key] = active
		active.done = startEffect(group, ctx, EffectSubscription, subscription.Key, subscription.Run, func(msg Msg) {
			send(token, msg)
		}, report, wake)
	}
}

func (s *subscriptionSet[Msg]) close() {
	for key, active := range s.active {
		active.accepting = false
		active.cancel()
		if active.timer != nil {
			active.timer.Stop()
		}
		delete(s.active, key)
	}
}

func (s *subscriptionSet[Msg]) accepts(token subscriptionToken) bool {
	active, ok := s.active[token.key]
	return ok && active.accepting && active.token == token
}

func (s *subscriptionSet[Msg]) stop(active *activeSubscription, wake func()) {
	if active.stopping {
		return
	}
	active.accepting = false
	active.stopping = true
	active.cancel()
	gracePeriod := s.stopGracePeriod
	if gracePeriod <= 0 {
		gracePeriod = defaultSubscriptionStopGracePeriod
	}
	active.deadline = time.Now().Add(gracePeriod)
	if !effectDone(active.done) && wake != nil {
		active.timer = time.AfterFunc(gracePeriod, wake)
	}
}

func effectDone(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}
