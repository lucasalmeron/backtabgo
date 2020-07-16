package card

import (
	"context"
	"fmt"
	"reflect"
)

var repository Repository

func SetRepository(repo Repository) {
	repository = repo
}

func (card *Card) Validate() error {
	if card.Word == "" {
		return fmt.Errorf("Field word is missing")
	}
	if len(card.ForbiddenWords) != 5 {
		return fmt.Errorf("Forbidden Words must be 5")
	}
	if reflect.TypeOf(card.Word).Kind() != reflect.String {
		return fmt.Errorf("Word must be a string")
	}
	if len(card.Word) <= 2 || len(card.Word) >= 30 {
		return fmt.Errorf("Word must be at least 2 characters long and not exceed 30 characters")
	}
	if reflect.TypeOf(card.ForbiddenWords).Kind() != reflect.Slice {
		return fmt.Errorf("ForbiddenWords must be an array")
	}
	for _, fw := range card.ForbiddenWords {
		if reflect.TypeOf(fw).Kind() != reflect.String {
			return fmt.Errorf("ForbiddenWord must be a string")
		}
		if len(fw) <= 2 || len(fw) >= 30 {
			return fmt.Errorf("ForbiddenWord must be at least 2 characters long and not exceed 30 characters")
		}
	}
	return nil
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
