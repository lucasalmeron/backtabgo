package player

import (
	"github.com/google/uuid"
)

type Message struct {
	Action   string      `json:"action"`
	Data     interface{} `json:"data"`
	PlayerID uuid.UUID   `json:"triggerPlayer"`
}

type Player struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Team   int       `json:"team"`
	Admin  bool      `json:"admin"`
	Status string    `json:"status"`
	Socket WebSocket `json:"-"`
}

type WebSocket interface {
	Write(message Message, playerTrigger uuid.UUID) error
	Read(IncommingMessagesChannel chan Message, playerTrigger uuid.UUID)
	CloseSocket() error
}
