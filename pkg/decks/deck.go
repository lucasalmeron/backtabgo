package deck

import (
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
)

type Deck struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Theme       string                `json:"theme"`
	CardsLength int                   `json:"cardsLength"`
	Cards       map[string]*card.Card `json:"-"`
}
