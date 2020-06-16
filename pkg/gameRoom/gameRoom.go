package gameRoom

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
	"github.com/mitchellh/mapstructure"
)

type GameRoom struct {
	ID              uuid.UUID                    `json:"id"`
	Players         map[uuid.UUID]*player.Player `json:"players"`
	Team1Score      int                          `json:"team1Score"`
	Team2Score      int                          `json:"team2Score"`
	CurrentTurn     *player.Player               `json:"currentTurn"`
	TurnTime        time.Time                    `json:"turnTime"`
	GameTime        time.Time                    `json:"gameTurn"`
	MaxTurnAttemps  int                          `json:"maxTurnAttemps"`
	Decks           map[uuid.UUID]deck.Deck      `json:"decks"`
	CurrentCard     card.Card                    `json:"currentCard"`
	MaxPoints       int                          `json:"maxPoints"`
	GameRoomChannel chan player.Message          `json:"-"`
}

func CreateGameRoom() *GameRoom {
	return &GameRoom{
		ID:              uuid.New(),
		Players:         map[uuid.UUID]*player.Player{},
		Team1Score:      0,
		Team2Score:      0,
		CurrentTurn:     &player.Player{},
		TurnTime:        time.Time{},
		GameTime:        time.Time{},
		MaxTurnAttemps:  0,
		Decks:           map[uuid.UUID]deck.Deck{},
		CurrentCard:     card.Card{},
		MaxPoints:       100,
		GameRoomChannel: make(chan player.Message),
	}

}

//StartListen asdasd
func (gameRoom *GameRoom) StartListen() {
	defer func() {
		//pop gameroom
	}()
	for message := range gameRoom.GameRoomChannel {
		fmt.Println("message ", message)
		switch message.Action {
		case "updateRoomOptions":
			if gameRoom.Players[message.PlayerID].Admin {
				var output GameRoom
				cfg := &mapstructure.DecoderConfig{
					Metadata: nil,
					Result:   &output,
					TagName:  "data",
				}
				decoder, _ := mapstructure.NewDecoder(cfg)
				decoder.Decode(message.Data)

				gameRoom.MaxTurnAttemps = output.MaxTurnAttemps
				gameRoom.MaxPoints = output.MaxPoints

				message.Data = gameRoom
				gameRoom.Players[message.PlayerID].Write(message)
			}
		case "changeTeam":
			gameRoom.Players[message.PlayerID].Team = message.Data.(int)
			message.Data = gameRoom.Players[message.PlayerID]
			for _, player := range gameRoom.Players {
				player.Write(message)
			}
		case "getPlayerList":
			playerList := make([]player.Player, 0)

			for _, player := range gameRoom.Players {
				playerList = append(playerList, *player)
			}
			message.Data = playerList
			gameRoom.Players[message.PlayerID].Write(message)
		case "connected":
			//send PlayerList to new Player
			playerList := make([]player.Player, 0)
			for _, player := range gameRoom.Players {
				playerList = append(playerList, *player)
			}
			message.Data = playerList
			gameRoom.Players[message.PlayerID].Write(message)

			//broadcast new player
			message.Action = "joinPlayer"
			message.Data = gameRoom.Players[message.PlayerID]
			for _, player := range gameRoom.Players {
				if player.ID != message.PlayerID {
					player.Write(message)
				}
			}
		case "reconnected":
			//send PlayerList to new Player
			playerList := make([]player.Player, 0)
			for _, player := range gameRoom.Players {
				playerList = append(playerList, *player)
			}
			message.Data = playerList
			gameRoom.Players[message.PlayerID].Write(message)

			//broadcast reconnected player
			message.Action = "playerReconnected"
			message.Data = gameRoom.Players[message.PlayerID]
			for _, player := range gameRoom.Players {
				if player.ID != message.PlayerID {
					player.Write(message)
				}
			}
		case "kickPlayerTimeOut":
			delete(gameRoom.Players, message.PlayerID)
			for _, player := range gameRoom.Players {
				player.Write(message)
			}
		case "playerDisconnected":
			//maybe i should set new admin here
			//delete(gameRoom.Players, message.PlayerID)
			message.Data = gameRoom.Players[message.PlayerID]
			for _, player := range gameRoom.Players {
				player.Write(message)
			}
		case "changeName":
			gameRoom.Players[message.PlayerID].Name = fmt.Sprintf("%v", message.Data)
			message.Data = gameRoom.Players[message.PlayerID]
			for _, player := range gameRoom.Players {
				player.Write(message)
			}

		}
	}
}
