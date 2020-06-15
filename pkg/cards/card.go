package card

import "github.com/google/uuid"

type Card struct {
	ID             uuid.UUID `json:"id"`
	Word           string    `json:"word"`
	ForbbidenWords []string  `json:"forbbidenWords"`
}
