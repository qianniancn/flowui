package style

import "time"

func (s Style) Transition(property PropertyID, duration time.Duration, options ...TransitionOption) Style {
	// Soft-ignore properties outside the animation whitelist (spec 07).
	if !property.Animatable() {
		return s
	}
	s.transitions = append([]Transition(nil), s.transitions...)
	transition := transition(&s.transitions, property)
	transition.Duration = max(duration, 0)
	for _, option := range options {
		if option != nil {
			option(transition)
		}
	}
	return s
}

func transition(transitions *[]Transition, property PropertyID) *Transition {
	for index := range *transitions {
		if (*transitions)[index].Property == property {
			return &(*transitions)[index]
		}
	}
	*transitions = append(*transitions, Transition{Property: property})
	return &(*transitions)[len(*transitions)-1]
}
