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

type RequestDeck struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Theme       string   `json:"theme"`
	CardsLength int      `json:"cardsLength"`
	Cards       []string `json:"cards"`
}

type Repository interface {
	GetDecks(ctx context.Context) ([]Deck, error)
	GetDecksWithCards(ctx context.Context) ([]Deck, error)
	GetDeck(ctx context.Context, deckID string) (*Deck, error)
	NewDeck(ctc context.Context, deck RequestDeck) (*Deck, error)
	UpdateDeck(ctc context.Context, deck RequestDeck) (*Deck, error)
	DeleteDeck(ctc context.Context, deck Deck) error
}
