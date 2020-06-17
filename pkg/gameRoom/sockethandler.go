package gameroom

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

type SocketRequest struct {
	message  player.Message
	gameRoom *GameRoom
}

//Route classify the incomming message and call the appropiate function
func (req *SocketRequest) Route() {
	switch req.message.Action {
	case "getDecks":
		req.getDecks()
	case "updateRoomOptions":
		req.updateRoomOptions()
	case "changeTeam":
		req.changeTeam()
	case "getPlayerList":
		req.getPlayerList()
	case "connected":
		req.connected()
	case "reconnected":
		req.reconnected()
	case "kickPlayerTimeOut":
		req.kickPlayerTimeOut()
	case "playerDisconnected":
		req.playerDisconnected()
	case "changeName":
		req.changeName()
	default:
		fmt.Println("doesn't match any socket endpoint")
	}

}

func (req *SocketRequest) getDecks() {
	deck1 := &deck.Deck{
		ID:    uuid.New(),
		Name:  "Totoro",
		Theme: "Caca",
		Cards: map[uuid.UUID]*card.Card{},
	}
	deck2 := &deck.Deck{
		ID:    uuid.New(),
		Name:  "Rebeca",
		Theme: "Tolueno",
		Cards: map[uuid.UUID]*card.Card{},
	}
	card1 := &card.Card{
		ID:             uuid.New(),
		Word:           "Caca",
		ForbbidenWords: []string{"Culo", "Materia Fecal", "Toto", "Baño"},
	}
	card2 := &card.Card{
		ID:             uuid.New(),
		Word:           "Culo",
		ForbbidenWords: []string{"Caca", "Materia Fecal", "Toto", "Baño"},
	}
	deck1.Cards[card1.ID] = card1
	deck1.Cards[card2.ID] = card2

	deck2.Cards[card1.ID] = card1
	deck2.Cards[card2.ID] = card2

	decks := []deck.Deck{*deck1, *deck2}

	req.message.Data = decks
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)
}

func (req *SocketRequest) updateRoomOptions() {
	if req.gameRoom.Players[req.message.PlayerID].Admin {
		//parsing map[string] interface{} to struct
		output := &GameRoom{}
		j, _ := json.Marshal(req.message.Data)
		json.Unmarshal(j, output)
		//parsing map[string] interface{} to struct

		req.gameRoom.MaxTurnAttemps = output.MaxTurnAttemps
		req.gameRoom.MaxPoints = output.MaxPoints

		req.message.Data = req.gameRoom
		req.gameRoom.Players[req.message.PlayerID].Write(req.message)
	}
}

func (req *SocketRequest) changeTeam() {

	//parsing map[string] interface{} to struct
	output := &player.Player{}
	j, _ := json.Marshal(req.message.Data)
	json.Unmarshal(j, output)
	//parsing map[string] interface{} to struct

	//Change team other player if is admin
	if output.ID != req.message.PlayerID {
		if req.gameRoom.Players[req.message.PlayerID].Admin {
			if req.gameRoom.Players[output.ID].Team == 1 {
				req.gameRoom.Players[output.ID].Team = 2
			} else {
				req.gameRoom.Players[output.ID].Team = 1
			}
			req.message.Data = req.gameRoom.Players[output.ID]
		} else {
			req.message.Data = "don't have permissions"
		}
	} else {
		if req.gameRoom.Players[req.message.PlayerID].Team == 1 {
			req.gameRoom.Players[req.message.PlayerID].Team = 2
		} else {
			req.gameRoom.Players[req.message.PlayerID].Team = 1
		}
		req.message.Data = req.gameRoom.Players[req.message.PlayerID]
	}
	if req.message.Data != nil {
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
	}

}

func (req *SocketRequest) getPlayerList() {
	playerList := make([]player.Player, 0)

	for _, player := range req.gameRoom.Players {
		playerList = append(playerList, *player)
	}
	req.message.Data = playerList
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)
}

func (req *SocketRequest) connected() {
	//send PlayerList to new Player
	playerList := make([]player.Player, 0) //review declaration
	for _, player := range req.gameRoom.Players {
		playerList = append(playerList, *player)
	}
	req.message.Data = playerList
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)

	//broadcast new player
	req.message.Action = "joinPlayer"
	req.message.Data = req.gameRoom.Players[req.message.PlayerID]
	for _, player := range req.gameRoom.Players {
		if player.ID != req.message.PlayerID {
			player.Write(req.message)
		}
	}
}

func (req *SocketRequest) reconnected() {
	//send PlayerList to new Player
	playerList := make([]player.Player, 0)
	for _, player := range req.gameRoom.Players {
		playerList = append(playerList, *player)
	}
	req.message.Data = playerList
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)

	//broadcast reconnected player
	req.message.Action = "playerReconnected"
	req.message.Data = req.gameRoom.Players[req.message.PlayerID]
	for _, player := range req.gameRoom.Players {
		if player.ID != req.message.PlayerID {
			player.Write(req.message)
		}
	}
}

func (req *SocketRequest) kickPlayerTimeOut() {
	delete(req.gameRoom.Players, req.message.PlayerID)
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}

func (req *SocketRequest) playerDisconnected() {
	//maybe i should set new admin here
	//delete(gameRoom.Players, message.PlayerID)
	req.message.Data = req.gameRoom.Players[req.message.PlayerID]
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}

func (req *SocketRequest) changeName() {
	req.gameRoom.Players[req.message.PlayerID].Name = fmt.Sprintf("%v", req.message.Data)
	req.message.Data = req.gameRoom.Players[req.message.PlayerID]
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}
