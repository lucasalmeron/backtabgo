package deck

import (
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
)

type Deck struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Theme string      `json:"theme"`
	Cards []card.Card `json:"cards"`
}
