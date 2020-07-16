package deck

import (
	"context"
	"fmt"
	"reflect"
)

var repository Repository

func SetRepository(repo Repository) {
	repository = repo
}

func (deck *Deck) Validate() error {
	if deck.Name == "" {
		return fmt.Errorf("Field Name is missing")
	}
	if deck.Theme == "" {
		return fmt.Errorf("Field Theme is missing")
	}

	if reflect.TypeOf(deck.Name).Kind() != reflect.String {
		return fmt.Errorf("Name must be a string")
	}
	if reflect.TypeOf(deck.Theme).Kind() != reflect.String {
		return fmt.Errorf("Theme must be a string")
	}
	if len(deck.Name) <= 2 || len(deck.Name) >= 30 {
		return fmt.Errorf("Name must be at least 2 characters long and not exceed 30 characters")
	}
	if len(deck.Theme) <= 2 || len(deck.Theme) >= 30 {
		return fmt.Errorf("Theme must be at least 2 characters long and not exceed 30 characters")
	}
	if reflect.TypeOf(deck.Cards).Kind() != reflect.Slice {
		return fmt.Errorf("Cards must be an array")
	}
	for _, card := range deck.Cards {
		if reflect.TypeOf(card).Kind() != reflect.String {
			return fmt.Errorf("Card must be a string")
		}
	}
	return nil
}

func (deck *Deck) GetDecks() ([]Deck, error) {
	return repository.GetDecks(context.Background())
}

func (deck *Deck) GetDecksWithCards() ([]Deck, error) {
	return repository.GetDecksWithCards(context.Background())
}

func (Deck *Deck) GetDeck(cardID string) (*Deck, error) {
	return repository.GetDeck(context.Background(), cardID)
}

func (Deck *Deck) NewDeck(newDeck RequestDeck) (*Deck, error) {
	return repository.NewDeck(context.Background(), newDeck)
}

func (Deck *Deck) UpdateDeck(reqDeck RequestDeck) (*Deck, error) {
	return repository.UpdateDeck(context.Background(), reqDeck)
}

func (Deck *Deck) DeleteDeck(reqDeck Deck) error {
	return repository.DeleteDeck(context.Background(), reqDeck)
}
