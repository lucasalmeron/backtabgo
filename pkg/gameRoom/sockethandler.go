package gameroom

import (
	"encoding/json"
	"fmt"
	"time"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
	"github.com/lucasalmeron/backtabgo/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SocketRequest struct {
	message  player.Message
	gameRoom *GameRoom
}

//Route classify the incomming message and call the appropiate function
func (req *SocketRequest) Route() {
	switch req.message.Action {
	case "startGame":
		req.startGame()
	case "playTurn": //nextTurn
		req.playTurn()
	case "broadcastNextPlayerTurn": //nextTurn
		req.broadcastNextPlayerTurn()
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

func (req *SocketRequest) startGame() {
	//check min players
	if req.gameRoom.Players[req.message.PlayerID].Admin && req.gameRoom.GameStatus == "waitingPlayers" && len(req.gameRoom.Settings.Decks) != 0 {
		req.message.Data = "Starting game..."
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
		req.gameRoom.Wg.Add(1)
		time.Sleep(5 * time.Second)
		go req.gameRoom.StartGame()
		req.message.Data = "Game started"
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
	}
}

func (req *SocketRequest) playTurn() {
	if req.message.PlayerID == req.gameRoom.CurrentTurn.ID && req.gameRoom.GameStatus == "gameInCourse" {
		req.gameRoom.TakeCard()
		req.gameRoom.GameChannel <- true
		//retornar carta al player y game data a los demás
		req.message.Action = "startTurn"
		req.message.Data = req.gameRoom.CurrentTurn
		for _, player := range req.gameRoom.Players {
			player.Write(req.message)
		}
	}
}

func (req *SocketRequest) broadcastNextPlayerTurn() {
	req.message.Action = "nextPlayerTurn"
	//add game data
	req.message.Data = req.gameRoom
	for _, player := range req.gameRoom.Players {
		player.Write(req.message)
	}
}

func (req *SocketRequest) getDecks() {

	/*groupStage := bson.D{
		{"$lookup", bson.D{
			{"from", "cards"},
			{"localField", "cards"},
			{"foreignField", "_id"},
			{"as", "cards"},
		}},
	}
	db := storage.GetMongoDBConnection()
	decks, err := db.Aggregate("decks", groupStage)
	if err != nil {
		req.message.Data = "db error"
	}*/
	db := storage.GetMongoDBConnection()
	dbDecks, err := db.FindAll("decks")
	if err != nil {
		req.message.Data = "db error"
	}

	for _, dbDeck := range dbDecks {
		delete(dbDeck, "cards")
	}

	/*data, err := db.InsertOne("cards", card.Card{
		Word:           "Radiador",
		ForbbidenWords: []string{"Agua", "Motor", "Regrigerante", "Enfriar", "Auto"},
	})
	if err != nil {
		req.message.Data = err
	}*/
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

		//req.gameRoom.Settings.TurnTime = output.GameTime
		req.gameRoom.Settings.TurnTime = output.TurnTime
		req.gameRoom.Settings.MaxTurnAttemps = output.MaxTurnAttemps
		req.gameRoom.Settings.MaxPoints = output.MaxPoints

		groupStage := bson.D{
			{"$lookup", bson.D{
				{"from", "cards"},
				{"localField", "cards"},
				{"foreignField", "_id"},
				{"as", "cards"},
			}},
		}
		db := storage.GetMongoDBConnection()
		dbDecks, err := db.Aggregate("decks", groupStage)
		if err != nil {
			req.message.Data = "db error"
		}
		//{"action":"updateRoomOptions","data":{"turnTime":20,"decks":["5eeead0fcc4d1e8c5f635a18"]}}

		// PARSE PRIMITIVES FROM MONGO TO STRUCT
		for _, reqDeckID := range output.Decks {
			for _, dbDeck := range dbDecks {
				stringID := dbDeck["_id"].(primitive.ObjectID).Hex()
				if stringID == reqDeckID {
					parseDeck := deck.Deck{
						ID:    reqDeckID,
						Name:  dbDeck["name"].(string),
						Theme: dbDeck["theme"].(string),
						Cards: map[string]*card.Card{},
					}
					for _, dbCard := range dbDeck["cards"].(primitive.A) {
						primitiveCard := dbCard.(primitive.M)
						parseCard := card.Card{
							ID:   primitiveCard["_id"].(primitive.ObjectID).Hex(),
							Word: primitiveCard["word"].(string),
						}
						for _, fword := range primitiveCard["forbbidenWords"].(primitive.A) {
							parseCard.ForbbidenWords = append(parseCard.ForbbidenWords, fword.(string))
						}
						parseDeck.Cards[parseCard.ID] = &parseCard
					}
					req.gameRoom.Settings.Decks[reqDeckID] = parseDeck
				}
			}
		}
		// PARSE PRIMITIVES FROM MONGO TO STRUCT

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
	//send PlayerList to new Player
	/*playerList := make([]player.Player, 0) //review declaration
	for _, player := range req.gameRoom.Players {
		playerList = append(playerList, *player)
	}*/

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
	//send PlayerList to new Player
	/*playerList := make([]player.Player, 0)
	for _, player := range req.gameRoom.Players {
		playerList = append(playerList, *player)
	}*/

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
}

func (req *SocketRequest) kickPlayerTimeOut() {
	//delete(req.gameRoom.Players, req.message.PlayerID)
	req.message.Data = req.gameRoom.Players[req.message.PlayerID]
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
