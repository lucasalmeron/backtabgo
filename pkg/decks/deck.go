package deck

import (
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
)

type Deck struct {
	ID    string                `json:"id"`
	Name  string                `json:"name"`
	Theme string                `json:"theme"`
	Cards map[string]*card.Card `json:"-"`
}
