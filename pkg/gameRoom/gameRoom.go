package gameroom

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

//util
func mapRandomKeyGet(mapI interface{}) interface{} {
	keys := reflect.ValueOf(mapI).MapKeys()

	return keys[rand.Intn(len(keys))].Interface()
}

type TurnMistakes struct {
	Word    string           `json:"word"`
	Players []*player.Player `json:"players"`
}

type GameSettings struct {
	MaxTurnAttemps int                   `json:"maxTurnAttemps"`
	Decks          map[string]*deck.Deck `json:"decks"`
	MaxPoints      int                   `json:"maxPoints"`
	TurnTime       int                   `json:"turnTime"`
	GameTime       int                   `json:"gameTurn"`
}

type GameRoom struct {
	ID                       uuid.UUID                    `json:"id"`
	Players                  map[uuid.UUID]*player.Player `json:"-"`
	Team1Score               int                          `json:"team1Score"`
	Team2Score               int                          `json:"team2Score"`
	PlayersTeam1             []*player.Player             `json:"team1"`
	PlayersTeam2             []*player.Player             `json:"team2"`
	TeamTurn                 int                          `json:"teamTurn"`
	TurnMistakes             []*TurnMistakes              `json:"teamMistakes"`
	TotalCards               int                          `json:"totalCards"`
	CurrentTurn              *player.Player               `json:"currentTurn"`
	CurrentCard              *card.Card                   `json:"-"`
	GameStatus               string                       `json:"gameStatus"`
	Settings                 *GameSettings                `json:"settings"`
	GameChannel              chan interface{}             `json:"-"`
	PlayerConnectedChannel   chan player.Player           `json:"-"`
	IncommingMessagesChannel chan player.Message          `json:"-"`
	Wg                       sync.WaitGroup               `json:"-"`
	Mutex                    sync.Mutex                   `json:"-"`
}

//CreateGameRoom is a constructor of GameRoom
func CreateGameRoom() *GameRoom {
	return &GameRoom{
		ID:                       uuid.New(),
		Players:                  map[uuid.UUID]*player.Player{},
		Team1Score:               0,
		Team2Score:               0,
		TeamTurn:                 1,
		GameStatus:               "roomPhase",
		GameChannel:              make(chan interface{}),
		PlayerConnectedChannel:   make(chan player.Player),
		IncommingMessagesChannel: make(chan player.Message),
		Settings: &GameSettings{
			MaxPoints:      50,
			MaxTurnAttemps: 0,
			TurnTime:       1,
			GameTime:       20,
			Decks:          map[string]*deck.Deck{},
		},
		Wg:    sync.WaitGroup{},
		Mutex: sync.Mutex{},
	}
}

func (gameRoom *GameRoom) AddPlayer(conn *websocket.Conn) {

	playerNumber := strconv.Itoa(len(gameRoom.Players) + 1)
	player := &player.Player{
		ID:                       uuid.New(),
		Name:                     "Player " + playerNumber,
		Status:                   "connected",
		Socket:                   conn,
		IncommingMessagesChannel: gameRoom.IncommingMessagesChannel,
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
		gameRoom.Mutex.Lock()
		gameRoom.PlayersTeam2 = append(gameRoom.PlayersTeam2, player)
		gameRoom.Players[player.ID] = player
		gameRoom.Mutex.Unlock()
	} else {
		player.Team = 1
		gameRoom.Mutex.Lock()
		gameRoom.PlayersTeam1 = append(gameRoom.PlayersTeam1, player)
		gameRoom.Players[player.ID] = player
		gameRoom.Mutex.Unlock()
	}

	gameRoom.Wg.Add(1)
	if gameRoom.GameStatus != "roomPhase" {
		gameRoom.PlayerConnectedChannel <- *player
	}

	player.Read(false)

}

func (gameRoom *GameRoom) ReconnectPlayer(conn *websocket.Conn, player *player.Player) {
	gameRoom.Mutex.Lock()
	player.Socket = conn
	player.Status = "connected"
	gameRoom.Mutex.Unlock()
	gameRoom.Wg.Add(1)
	if gameRoom.GameStatus != "roomPhase" {
		gameRoom.PlayerConnectedChannel <- *player
	}
	player.Read(true)
}

//it check if are minimum 2 players in each team and it stay waiting for reconnects or connects
func (gameRoom *GameRoom) checkMinPlayersConnection() {
	for {
		disconnectedCountT1 := 0
		for _, player := range gameRoom.PlayersTeam1 {
			if player.Status == "disconnected" {
				disconnectedCountT1++
			}
		}
		disconnectedCountT2 := 0
		for _, player := range gameRoom.PlayersTeam2 {
			if player.Status == "disconnected" {
				disconnectedCountT2++
			}
		}
		if len(gameRoom.PlayersTeam1)-disconnectedCountT1 >= 2 && len(gameRoom.PlayersTeam2)-disconnectedCountT2 >= 2 {
			break
		}
		//broadcast Waiting for players
		gameRoom.sendMessage("waitingForPlayers", gameRoom, uuid.UUID{})

		<-gameRoom.PlayerConnectedChannel
	}
}

func (gameRoom *GameRoom) setNextPlayer(currentIndex1 *int, currentIndex2 *int) {
	if gameRoom.TeamTurn == 1 {
		for {
			if gameRoom.PlayersTeam1[*currentIndex1].Status == "connected" &&
				gameRoom.PlayersTeam1[*currentIndex1].ID != gameRoom.CurrentTurn.ID {
				break
			}
			if len(gameRoom.PlayersTeam1)-1 == *currentIndex1 {
				*currentIndex1 = 0
			} else {
				*currentIndex1++
			}

		}
		//maybe datarace here...
		gameRoom.CurrentTurn = gameRoom.PlayersTeam1[*currentIndex1]
		gameRoom.TeamTurn = 2

	} else {
		for {
			if gameRoom.PlayersTeam1[*currentIndex2].Status == "connected" &&
				gameRoom.PlayersTeam2[*currentIndex2].ID != gameRoom.CurrentTurn.ID {
				break
			}
			if len(gameRoom.PlayersTeam2)-1 == *currentIndex2 {
				*currentIndex2 = 0
			} else {
				*currentIndex2++
			}
		}
		//maybe datarace here...
		gameRoom.CurrentTurn = gameRoom.PlayersTeam2[*currentIndex2]
		gameRoom.TeamTurn = 1
	}
}

func (gameRoom *GameRoom) StartGame() {
	lastPlayerTeam1Index := 0
	lastPlayerTeam2Index := 0
	defer func() {
		close(gameRoom.PlayerConnectedChannel)
		close(gameRoom.GameChannel)
		close(gameRoom.IncommingMessagesChannel)
	}()
	for {
		//check if are minimum 2 players in each team
		gameRoom.checkMinPlayersConnection()

		//set current player and index for next player
		gameRoom.setNextPlayer(&lastPlayerTeam1Index, &lastPlayerTeam2Index)

		gameRoom.GameStatus = "gameInCourse"

		//broadcast Next Player Turn
		gameRoom.sendMessage("broadcastNextPlayerTurn", gameRoom.CurrentTurn, uuid.UUID{})

		fmt.Println("waiting for take a card...")
		<-gameRoom.GameChannel
		fmt.Println("taken card, turn in course")

		gameRoom.Mutex.Lock()
		gameRoom.GameStatus = "turnInCourse"
		gameRoom.Mutex.Unlock()
		//TIME TO SEND ATTEMPS
		time.Sleep(time.Duration(gameRoom.Settings.TurnTime) * time.Minute)

		fmt.Println("END TURN")

		if gameRoom.Settings.MaxPoints <= gameRoom.Team1Score || gameRoom.Settings.MaxPoints <= gameRoom.Team2Score || gameRoom.TotalCards == 0 {

			gameRoom.GameStatus = "gameEnded"
			//broadcast game is end
			gameRoom.sendMessage("gameEnded", gameRoom, uuid.UUID{})

			break
		}
	}
	fmt.Println("game end")
}

func (gameRoom *GameRoom) TakeCard() {
	//{"action":"updateRoomOptions","data":{"turnTime":1,"maxPoints":50,"decks":["5eeead0fcc4d1e8c5f635a18"]}}
	var randKeyDeck string
	var randKeyCard string
	for {
		randKeyDeck = mapRandomKeyGet(gameRoom.Settings.Decks).(string)
		if gameRoom.Settings.Decks[randKeyDeck].CardsLength > 0 {
			break
		}
	}
	randKeyCard = mapRandomKeyGet(gameRoom.Settings.Decks[randKeyDeck].Cards).(string)
	card := gameRoom.Settings.Decks[randKeyDeck].Cards[randKeyCard]
	delete(gameRoom.Settings.Decks[randKeyDeck].Cards, randKeyCard)
	gameRoom.Settings.Decks[randKeyDeck].CardsLength--
	gameRoom.CurrentCard = card
	gameRoom.TurnMistakes = nil
	gameRoom.TurnMistakes = append(gameRoom.TurnMistakes, &TurnMistakes{
		Word: card.Word,
	})
	for _, word := range card.ForbiddenWords {
		gameRoom.TurnMistakes = append(gameRoom.TurnMistakes, &TurnMistakes{
			Word: word,
		})
	}

}

func (gameRoom *GameRoom) PlayTurn() {
	//check if are minimum 2 players in each team
	gameRoom.checkMinPlayersConnection()

	gameRoom.TakeCard()
	gameRoom.GameChannel <- true
}

func (gameRoom *GameRoom) SubmitPlayerAttemp(attemp string) bool {
	currentCardWord := strings.ToUpper(gameRoom.CurrentCard.Word)
	attempWord := strings.ToUpper(attemp)
	if currentCardWord == attempWord {
		if gameRoom.CurrentTurn.Team == 1 {
			gameRoom.Team1Score++
		} else {
			gameRoom.Team2Score++
		}
		return true
	}
	return false
}

//Send messages to handler
func (gameRoom *GameRoom) sendMessage(action string, message interface{}, triggerPlayer uuid.UUID) {
	socketReq := SocketRequest{
		message: player.Message{
			Action:   action,
			Data:     message,
			PlayerID: triggerPlayer,
		},
		gameRoom: gameRoom,
	}
	//could i lock each action on socket handler???
	gameRoom.Mutex.Lock()
	socketReq.Route()
	gameRoom.Mutex.Unlock()
}

//StartListen channel and wait for player's incomming messages, then it call socketRequest to classify
func (gameRoom *GameRoom) StartListenSocketMessages() {
	defer gameRoom.Wg.Done()
	for message := range gameRoom.IncommingMessagesChannel {
		fmt.Println("message ", message)
		if message.Action == "playerDisconnected" {
			gameRoom.Wg.Done()
		}
		gameRoom.sendMessage(message.Action, message.Data, message.PlayerID)
	}
}
