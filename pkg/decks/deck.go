package deck

import (
	"context"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
)

type Deck struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Theme       string                `json:"theme"`
	CardsLength int                   `json:"cardsLength"`
	Cards       map[string]*card.Card `json:"-"`
}

type DeckRepository interface {
	GetDecks(ctx context.Context) ([]Deck, error)
	GetDecksWithCards(ctx context.Context) ([]Deck, error)
	GetDeck(ctx context.Context, deckID string) (*Deck, error)
	NewDeck(ctc context.Context, deck Deck) (*Deck, error)
	UpdateDeck(ctc context.Context, deck Deck) (*Deck, error)
	DeleteDeck(ctc context.Context, deck Deck) error
}
