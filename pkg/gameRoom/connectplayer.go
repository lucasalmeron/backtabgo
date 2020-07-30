package gameroom

import (
	"strconv"

	"github.com/google/uuid"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

func (gameRoom *GameRoom) AddPlayer(conn player.WebSocket) {

	gameRoom.Mutex.Lock()
	playerNumber := strconv.Itoa(len(gameRoom.Players) + 1)
	player := &player.Player{
		ID:     uuid.New(),
		Name:   "Player " + playerNumber,
		Status: "connected",
		Socket: conn,
	}

	//set admin
	if len(gameRoom.Players) == 0 {
		player.Admin = true
	}

	//balancing teams
	playerTeam1Count := 0
	playerTeam2Count := 0
	for _, player := range gameRoom.Players {
		if player.Team == 1 {
			playerTeam1Count++
		} else {
			playerTeam2Count++
		}
	}
	if playerTeam1Count > playerTeam2Count {
		player.Team = 2
		gameRoom.PlayersTeam2 = append(gameRoom.PlayersTeam2, player)
		gameRoom.Players[player.ID] = player
	} else {
		player.Team = 1
		gameRoom.PlayersTeam1 = append(gameRoom.PlayersTeam1, player)
		gameRoom.Players[player.ID] = player
	}

	gameRoom.Mutex.Unlock()
	gameRoom.Wg.Add(1)
	gameRoom.closePlayersWg.Add(1)
	if gameRoom.GameStatus == "waitingMinPlayers" {
		gameRoom.PlayerConnectedChannel <- *player
	}

	player.ReadMessages(false, gameRoom.IncommingMessagesChannel)

}

func (gameRoom *GameRoom) ReconnectPlayer(conn player.WebSocket, player *player.Player) {
	gameRoom.Mutex.Lock()
	player.Socket = conn
	player.Status = "connected"
	gameRoom.Mutex.Unlock()
	gameRoom.Wg.Add(1)
	gameRoom.closePlayersWg.Add(1)
	if gameRoom.GameStatus == "waitingMinPlayers" {
		gameRoom.PlayerConnectedChannel <- *player
	}
	player.ReadMessages(true, gameRoom.IncommingMessagesChannel)
}
