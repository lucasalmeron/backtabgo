package card

import "context"

type Card struct {
	ID             string   `json:"id" bson:"_id,omitempty"`
	Word           string   `json:"word" bson:"word"`
	ForbiddenWords []string `json:"forbiddenWords" bson:"forbiddenWords"`
}

type Repository interface {
	GetCards(ctx context.Context) ([]Card, error)
	GetCard(ctx context.Context, cardID string) (*Card, error)
	NewCard(ctc context.Context, card Card) (*Card, error)
	UpdateCard(ctc context.Context, card Card) (*Card, error)
	DeleteCard(ctc context.Context, card Card) error
}
