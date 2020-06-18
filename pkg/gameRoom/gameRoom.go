package gameroom

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

type GameSettings struct {
	MaxTurnAttemps int                     `json:"maxTurnAttemps"`
	Decks          map[uuid.UUID]deck.Deck `json:"decks"`
	MaxPoints      int                     `json:"maxPoints"`
	TurnTime       int                     `json:"turnTime"`
	GameTime       int                     `json:"gameTurn"`
}

type GameRoom struct {
	ID                       uuid.UUID                    `json:"id"`
	Players                  map[uuid.UUID]*player.Player `json:"players"`
	Team1Score               int                          `json:"team1Score"`
	Team2Score               int                          `json:"team2Score"`
	CurrentTurn              *player.Player               `json:"currentTurn"`
	CurrentTurnTime          time.Time                    `json:"currentTurnTime"`
	CurrentCard              card.Card                    `json:"currentCard"`
	Settings                 *GameSettings                `json:"settings"`
	IncommingMessagesChannel chan player.Message          `json:"-"`
}

//CreateGameRoom is a constructor of GameRoom
func CreateGameRoom() *GameRoom {
	return &GameRoom{
		ID:                       uuid.New(),
		Players:                  map[uuid.UUID]*player.Player{},
		Team1Score:               0,
		Team2Score:               0,
		CurrentTurn:              &player.Player{},
		CurrentCard:              card.Card{},
		IncommingMessagesChannel: make(chan player.Message),
		Settings: &GameSettings{
			MaxPoints:      100,
			MaxTurnAttemps: 0,
			TurnTime:       1,
			GameTime:       20,
			Decks:          map[uuid.UUID]deck.Deck{},
		},
	}
}

func (gameRoom *GameRoom) StartGame() {

	//time.AfterFunc(time.Duration(gameRoom.Settings.TurnTime)*time.Minute, gameRoom.nextTurn())
}

func (gameRoom *GameRoom) nextTurn() {

}

//StartListen channel and wait for player's incomming messages, then it call socketRequest to classify
func (gameRoom *GameRoom) StartListen() {
	defer func() {
		//pop gameroom
	}()
	for message := range gameRoom.IncommingMessagesChannel {
		fmt.Println("message ", message)
		socketReq := SocketRequest{
			message:  message,
			gameRoom: gameRoom,
		}
		socketReq.Route()
	}
}
