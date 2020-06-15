package gameRoom

import (
	"fmt"
	"runtime"
	"time"

	"github.com/google/uuid"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

type GameRoom struct {
	ID              uuid.UUID                    `json:"id"`
	Players         map[uuid.UUID]*player.Player `json:"players"`
	Team1Score      int                          `json:"team1Score"`
	Team2Score      int                          `json:"team2Score"`
	CurrentTurn     player.Player                `json:"currentTurn"`
	TurnTime        time.Time                    `json:"turnTime"`
	GameTime        time.Time                    `json:"gameTurn"`
	MaxTurnAttemps  int                          `json:"maxTurnAttemps"`
	Decks           map[uuid.UUID]deck.Deck      `json:"decks"`
	CurrentCard     card.Card                    `json:"currentCard"`
	MaxPoints       int                          `json:"maxPoints"`
	GameRoomChannel chan player.Message
}

func CreateGameRoom() *GameRoom {
	return &GameRoom{
		ID:              uuid.New(),
		Players:         map[uuid.UUID]*player.Player{},
		Team1Score:      0,
		Team2Score:      0,
		CurrentTurn:     player.Player{},
		TurnTime:        time.Time{},
		GameTime:        time.Time{},
		MaxTurnAttemps:  0,
		Decks:           map[uuid.UUID]deck.Deck{},
		CurrentCard:     card.Card{},
		MaxPoints:       100,
		GameRoomChannel: make(chan player.Message),
	}

}

func (gameRoom *GameRoom) Start() {
	for message := range gameRoom.GameRoomChannel {
		fmt.Println("message ", message)
		switch message.Type {
		case "joinPlayer":
			for _, player := range gameRoom.Players {
				player.Write(message)
			}

		case "kickPlayerTimeOut":
			delete(gameRoom.Players, message.PlayerID)
			for _, player := range gameRoom.Players {
				player.Write(message)
			}
		case "changeName":
			var pl = gameRoom.Players[message.PlayerID]
			pl.Name = fmt.Sprintf("%v", message.Data)

			for _, player := range gameRoom.Players {
				message.Name = fmt.Sprintf("%v", message.Data)
				player.Write(message)
			}
		}

		/*if player.ID != message.PlayerID {
			player.Write(message)
		}*/

		fmt.Println("Goroutines \t", runtime.NumGoroutine())
	}
}
