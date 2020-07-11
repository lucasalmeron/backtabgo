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
