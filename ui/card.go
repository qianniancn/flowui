package ui

import "github.com/qianniancn/flowui/internal/components/card"

type CardVariant = card.CardVariant
type CardWidget = card.CardWidget

const (
	CardDefault     = card.CardDefault
	CardSecondary   = card.CardSecondary
	CardTertiary    = card.CardTertiary
	CardTransparent = card.CardTransparent
)

func Card(children ...Widget) CardWidget {
	return card.Card(children...)
}
