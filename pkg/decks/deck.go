package deck

import (
	"github.com/google/uuid"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
)

type Deck struct {
	ID    uuid.UUID                `json:"id"`
	Name  string                   `json:"name"`
	Theme string                   `json:"theme"`
	Cards map[uuid.UUID]*card.Card `json:"cards"`
}
