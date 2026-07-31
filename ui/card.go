package ui

import "github.com/qianniancn/flowui/internal/components/card"

// CardVariant selects the card surface treatment.
type CardVariant = card.CardVariant

type CardWidget = card.CardWidget

const (
	CardDefault     = card.CardDefault
	CardSecondary   = card.CardSecondary
	CardTertiary    = card.CardTertiary
	CardTransparent = card.CardTransparent
)

// Card creates a themed surface for one or more child widgets.
func Card(children ...Widget) CardWidget {
	return card.Card(children...)
}
