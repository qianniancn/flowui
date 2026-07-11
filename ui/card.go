package ui

import "github.com/qianniancn/FlowUI/internal/components/card"

type CardVariant = card.CardVariant
type CardWidget = card.CardWidget
type CardHeaderWidget = card.CardHeaderWidget
type CardTitleWidget = card.CardTitleWidget
type CardDescriptionWidget = card.CardDescriptionWidget
type CardContentWidget = card.CardContentWidget
type CardFooterWidget = card.CardFooterWidget

const (
	CardDefault     = card.CardDefault
	CardSecondary   = card.CardSecondary
	CardTertiary    = card.CardTertiary
	CardTransparent = card.CardTransparent
)

func Card(children ...Widget) CardWidget {
	return card.Card(children...)
}

func CardHeader(children ...Widget) CardHeaderWidget {
	return card.CardHeader(children...)
}

func CardTitle(value string) CardTitleWidget {
	return card.CardTitle(value)
}

func CardDescription(value string) CardDescriptionWidget {
	return card.CardDescription(value)
}

func CardContent(children ...Widget) CardContentWidget {
	return card.CardContent(children...)
}

func CardFooter(children ...Widget) CardFooterWidget {
	return card.CardFooter(children...)
}
