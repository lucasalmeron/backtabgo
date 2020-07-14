package deck

import "context"

var repository DeckRepository

func SetRepository(repo DeckRepository) {
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

func (Deck *Deck) NewDeck(newCard Deck) (*Deck, error) {
	return repository.NewDeck(context.Background(), newCard)
}

func (Deck *Deck) UpdateDeck(reqCard Deck) (*Deck, error) {
	return repository.UpdateDeck(context.Background(), reqCard)
}

func (Deck *Deck) DeleteDeck(reqCard Deck) error {
	return repository.DeleteDeck(context.Background(), reqCard)
}
