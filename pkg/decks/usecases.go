package deck

import "context"

var repository Repository

func SetRepository(repo Repository) {
	repository = repo
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
