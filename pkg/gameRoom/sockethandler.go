package gameroom

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

type SocketRequest struct {
	message  player.Message
	gameRoom *GameRoom
}

//Route classifies the incomming message and call the appropiate function
func (req *SocketRequest) Route() {
	switch req.message.Action {
	case "startGame":
		req.startGame()
	case "playTurn":
		req.playTurn()
	case "broadcastNextPlayerTurn":
		req.broadcastNextPlayerTurn()
	case "waitingForPlayers":
		req.waitingForPlayers()
	case "skipCard":
		req.skipCard()
	case "submitAttemp":
		req.submitAttemp()
	case "submitMistake":
		req.submitMistake()
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
	case "kickPlayer":
		req.kickPlayer()
	case "playerDisconnected":
		req.playerDisconnected()
	case "changeName":
		req.changeName()
	case "gameEnded":
		req.gameEnded()
	case "keepAlive":
		req.keepAlive()
	default:
		fmt.Println("doesn't match any socket endpoint")
	}

}

func (req *SocketRequest) startGame() {
	if req.gameRoom.Players[req.message.PlayerID].Admin &&
		req.gameRoom.GameStatus == "roomPhase" &&
		len(req.gameRoom.Settings.Decks) > 0 &&
		len(req.gameRoom.Players) >= 4 {

		req.message.Data = "Starting game..."
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
		go req.gameRoom.StartGame()
		req.message.Action = "gameStarted"
		req.message.Data = "Game started"
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
	}
}

func (req *SocketRequest) playTurn() {
	if req.message.PlayerID == req.gameRoom.CurrentTurn.ID && req.gameRoom.GameStatus == "gameInCourse" {
		req.gameRoom.PlayTurn()

		//broadcast gamestatus and give a new card to player
		req.message.Action = "startTurn"
		req.message.Data = req.gameRoom
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}

		req.message.Action = "yourCard"
		req.message.Data = req.gameRoom.CurrentCard
		req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Write(req.message)

		//send currentCard to controller players
		req.message.Action = "currentCard"
		req.message.Data = req.gameRoom.CurrentCard
		if req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Team == 1 {
			for _, player := range req.gameRoom.PlayersTeam2 {
				player.Write(req.message)
			}
		} else {
			for _, player := range req.gameRoom.PlayersTeam1 {
				player.Write(req.message)
			}
		}
	}
}

func (req *SocketRequest) broadcastNextPlayerTurn() {
	req.message.Action = "nextPlayerTurn"
	req.message.Data = req.gameRoom
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}

func (req *SocketRequest) waitingForPlayers() {
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}

func (req *SocketRequest) skipCard() {
	if req.gameRoom.GameStatus == "turnInCourse" && req.gameRoom.CurrentTurn.ID == req.message.PlayerID {
		if req.gameRoom.TotalCards == 0 {
			req.message.Action = "emptyDecks"
			req.message.Data = "there aren't more cards"
			for _, player := range req.gameRoom.Players {
				player.Write(req.message)
			}
			return
		}
		req.message.Action = "cardSkipped"
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
		if req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Team == 1 {
			req.gameRoom.Team1Score--
		} else {
			req.gameRoom.Team2Score--
		}

		//send gameStatus
		req.message.Action = "gameStatus"
		req.message.Data = req.gameRoom
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}

		//send card to controller players
		err := req.gameRoom.TakeCard()
		if err != nil {
			req.message.Action = "emptyDecks"
			req.message.Data = "there aren't more cards"
			for _, player := range req.gameRoom.Players {
				player.Write(req.message)
			}
			return
		}
		req.message.Action = "yourCard"
		req.message.Data = req.gameRoom.CurrentCard
		req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Write(req.message)

		//send currentCard to controller players
		req.message.Action = "currentCard"
		req.message.Data = req.gameRoom.CurrentCard
		if req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Team == 1 {
			for _, player := range req.gameRoom.PlayersTeam2 {
				player.Write(req.message)
			}
		} else {
			for _, player := range req.gameRoom.PlayersTeam1 {
				player.Write(req.message)
			}
		}
	}
}

func (req *SocketRequest) submitAttemp() {
	if req.gameRoom.GameStatus == "turnInCourse" && req.gameRoom.CurrentTurn.Team == req.gameRoom.Players[req.message.PlayerID].Team {
		attempMessage := struct {
			Player        *player.Player
			SuccessAttemp bool
			Attemp        string
		}{
			Player: req.gameRoom.Players[req.message.PlayerID],
			Attemp: req.message.Data.(string),
		}
		if req.gameRoom.SubmitPlayerAttemp(req.message.Data.(string)) {
			//player boolean word

			attempMessage.SuccessAttemp = true
			req.message.Data = attempMessage
			for _, player := range req.gameRoom.Players {
				player.Write(req.message)
			}

			//send gameStatus
			req.message.Action = "gameStatus"
			req.message.Data = req.gameRoom
			for _, player := range req.gameRoom.Players {
				player.Write(req.message)
			}

			//send card to controller players
			err := req.gameRoom.TakeCard()
			if err != nil {
				req.message.Action = "emptyDecks"
				req.message.Data = "there aren't more cards"
				for _, player := range req.gameRoom.Players {
					player.Write(req.message)
				}
				return
			}
			req.message.Action = "yourCard"
			req.message.Data = req.gameRoom.CurrentCard
			req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Write(req.message)
			//ta = req.gameRoom.CurrentCard
			if req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Team == 1 {
				for _, player := range req.gameRoom.PlayersTeam2 {
					player.Write(req.message)
				}
			} else {
				for _, player := range req.gameRoom.PlayersTeam1 {
					player.Write(req.message)
				}
			}
		} else {
			attempMessage.SuccessAttemp = false
			req.message.Data = attempMessage
			for _, player := range req.gameRoom.Players {
				player.Write(req.message)
			}
		}
	}

}

func (req *SocketRequest) submitMistake() {
	if req.gameRoom.GameStatus == "turnInCourse" && req.gameRoom.CurrentTurn.Team != req.gameRoom.Players[req.message.PlayerID].Team {
		for _, mistake := range req.gameRoom.TurnMistakes {

			if strings.ToUpper(mistake.Word) == strings.ToUpper(req.message.Data.(string)) {
				var lengthPlayers int
				if req.gameRoom.Players[req.message.PlayerID].Team == 1 {
					lengthPlayers = len(req.gameRoom.PlayersTeam1)
				} else {
					lengthPlayers = len(req.gameRoom.PlayersTeam2)
				}
				for _, player := range mistake.Players {
					if player.ID == req.message.PlayerID {
						return
					}
				}

				mistake.Players = append(mistake.Players, req.gameRoom.Players[req.message.PlayerID])
				if len(mistake.Players) > (lengthPlayers / 2) {
					req.message.Action = "playerMistake"
					for _, player := range req.gameRoom.Players {
						player.Write(req.message)
					}
					if req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Team == 1 {
						req.gameRoom.Team1Score--
					} else {
						req.gameRoom.Team2Score--
					}

					//send gameStatus
					req.message.Action = "gameStatus"
					req.message.Data = req.gameRoom
					for _, player := range req.gameRoom.Players {
						player.Write(req.message)
					}

					//send card to controller players
					err := req.gameRoom.TakeCard()
					if err != nil {
						req.message.Action = "emptyDecks"
						req.message.Data = "there aren't more cards"
						for _, player := range req.gameRoom.Players {
							player.Write(req.message)
						}
						return
					}
					req.message.Action = "yourCard"
					req.message.Data = req.gameRoom.CurrentCard
					req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Write(req.message)

					//send currentCard to controller players
					req.message.Action = "currentCard"
					req.message.Data = req.gameRoom.CurrentCard
					if req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Team == 1 {
						for _, player := range req.gameRoom.PlayersTeam2 {
							player.Write(req.message)
						}
					} else {
						for _, player := range req.gameRoom.PlayersTeam1 {
							player.Write(req.message)
						}
					}
					break
				} else {
					req.message.Action = "mistakeSubmitted"
					req.message.Data = req.gameRoom.TurnMistakes
					for _, player := range req.gameRoom.Players {
						player.Write(req.message)
					}
					break
				}
			}
		}
	}
}

func (req *SocketRequest) getDecks() {

	deckService := new(deck.Deck)
	dbDecks, err := deckService.GetDecks()
	if err != nil {
		req.message.Data = "db error"
	}

	req.message.Data = dbDecks
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)
}

func (req *SocketRequest) updateRoomOptions() {
	if req.gameRoom.Players[req.message.PlayerID].Admin {
		//parsing map[string] interface{} to struct

		var output struct {
			MaxTurnAttemps int      `json:"maxTurnAttemps"`
			Decks          []string `json:"decks"`
			MaxPoints      int      `json:"maxPoints"`
			TurnTime       int      `json:"turnTime"`
			GameTime       int      `json:"gameTurn"`
		}

		j, _ := json.Marshal(req.message.Data)
		json.Unmarshal(j, &output)
		//parsing map[string] interface{} to struct

		//req.gameRoom.Settings.GameTime = output.GameTime
		if output.TurnTime > 0 {
			req.gameRoom.Settings.TurnTime = output.TurnTime
		}
		if output.MaxTurnAttemps > 5 {
			req.gameRoom.Settings.MaxTurnAttemps = output.MaxTurnAttemps
		}
		if output.MaxPoints > 1 {
			req.gameRoom.Settings.MaxPoints = output.MaxPoints
		}

		deckService := new(deck.Deck)
		dbDecks, err := deckService.GetDecksWithCards()
		if err != nil {
			req.message.Data = "db error"
		}

		totalCards := 0
		for _, reqDeckID := range output.Decks {
			for _, dbDeck := range dbDecks {
				var deck = dbDeck
				if deck.ID == reqDeckID {
					totalCards += dbDeck.CardsLength
					req.gameRoom.Settings.Decks[reqDeckID] = &deck
				}
			}
		}

		req.gameRoom.TotalCards = totalCards

		req.message.Data = req.gameRoom
		//broadcast Options
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)

		}
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

	req.message.Data = req.gameRoom
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
	//send reconected to player
	req.message.Data = req.gameRoom
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)

	//broadcast reconnected player
	req.message.Action = "playerReconnected"
	req.message.Data = req.gameRoom.Players[req.message.PlayerID]
	for _, player := range req.gameRoom.Players {
		if player.ID != req.message.PlayerID {
			player.Write(req.message)
		}
	}

	if req.gameRoom.GameStatus == "turnInCourse" {
		if req.gameRoom.CurrentTurn.ID == req.message.PlayerID {
			req.message.Action = "yourCard"
			req.message.Data = req.gameRoom.CurrentCard
			req.gameRoom.Players[req.gameRoom.CurrentTurn.ID].Write(req.message)
		} else {
			if req.gameRoom.CurrentTurn.Team != req.gameRoom.Players[req.message.PlayerID].Team {
				req.message.Action = "currentCard"
				req.message.Data = req.gameRoom.CurrentCard
				req.gameRoom.Players[req.message.PlayerID].Write(req.message)
			}
		}

	}

}

func (req *SocketRequest) kickPlayer() {
	if req.gameRoom.Players[req.message.PlayerID].Admin {
		req.message.Action = "playerKicked"
		req.message.Data = req.gameRoom.Players[req.message.Data.(uuid.UUID)]
		delete(req.gameRoom.Players, req.message.Data.(uuid.UUID))
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
	}

}

func (req *SocketRequest) playerDisconnected() {
	//maybe i should set new admin here
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

func (req *SocketRequest) gameEnded() {
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}

func (req *SocketRequest) keepAlive() {
	req.message.Data = "OK"
	req.gameRoom.Players[req.message.PlayerID].Write(req.message)
}
