package card

import "context"

var repository Repository

func SetRepository(repo Repository) {
	repository = repo
}

func (card *Card) GetCards() ([]Card, error) {
	return repository.GetCards(context.Background())
}

func (card *Card) GetCard(idCard string) (*Card, error) {
	return repository.GetCard(context.Background(), idCard)
}

func (card *Card) NewCard(newCard Card) (*Card, error) {
	return repository.NewCard(context.Background(), newCard)
}

func (card *Card) UpdateCard(reqCard Card) (*Card, error) {
	return repository.UpdateCard(context.Background(), reqCard)
}
