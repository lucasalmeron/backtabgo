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
		GameStatus:               "waitingPlayers",
		GameChannel:              make(chan interface{}),
		IncommingMessagesChannel: make(chan player.Message),
		Settings: &GameSettings{
			MaxPoints:      100,
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

	//balance teams
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

	player.Read(false)

}

func (gameRoom *GameRoom) ReconnectPlayer(conn *websocket.Conn, player *player.Player) {
	gameRoom.Mutex.Lock()
	player.Socket = conn
	player.Status = "connected"
	gameRoom.Mutex.Unlock()
	player.Read(true)
}

func (gameRoom *GameRoom) StartGame() {
	lastPlayerTeam1Index := 0
	lastPlayerTeam2Index := 0

	for {
		//set current player and index for next player
		if gameRoom.TeamTurn == 1 {
			gameRoom.CurrentTurn = gameRoom.PlayersTeam1[lastPlayerTeam1Index]
			gameRoom.TeamTurn = 2
			if len(gameRoom.PlayersTeam1)-1 == lastPlayerTeam1Index {
				lastPlayerTeam1Index = 0
			} else {
				lastPlayerTeam1Index++
			}
		} else {
			gameRoom.CurrentTurn = gameRoom.PlayersTeam2[lastPlayerTeam2Index]
			gameRoom.TeamTurn = 1
			if len(gameRoom.PlayersTeam2)-1 == lastPlayerTeam2Index {
				lastPlayerTeam2Index = 0
			} else {
				lastPlayerTeam2Index++
			}
		}
		//set current player and index for next player
		gameRoom.GameStatus = "gameInCourse"

		//broadcast Next Player Turn
		socketReq := SocketRequest{
			message: player.Message{
				Action: "broadcastNextPlayerTurn",
				Data:   gameRoom.CurrentTurn,
			},
			gameRoom: gameRoom,
		}
		gameRoom.Mutex.Lock()
		socketReq.Route()
		gameRoom.Mutex.Unlock()
		//broadcast Next Player Turn

		fmt.Println("waiting for take a card...")
		<-gameRoom.GameChannel
		fmt.Println("taken card, turn in course")

		gameRoom.Mutex.Lock()
		gameRoom.GameStatus = "turnInCourse"
		gameRoom.Mutex.Unlock()
		//TIME TO SEND ATTEMPS
		time.Sleep(time.Duration(gameRoom.Settings.TurnTime) * time.Minute)

		fmt.Println("END TURN")

		//add empty deck
		if gameRoom.Settings.MaxPoints <= gameRoom.Team1Score || gameRoom.Settings.MaxPoints <= gameRoom.Team2Score || gameRoom.TotalCards == 0 {
			break
		}
	}
	fmt.Println("game end")
}

func mapRandomKeyGet(mapI interface{}) interface{} {
	keys := reflect.ValueOf(mapI).MapKeys()

	return keys[rand.Intn(len(keys))].Interface()
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
	for _, word := range card.ForbbidenWords {
		gameRoom.TurnMistakes = append(gameRoom.TurnMistakes, &TurnMistakes{
			Word: word,
		})
	}

}

func (gameRoom *GameRoom) PlayTurn() {
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
	} else {
		return false
	}
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
		//could i lock each action on socket handler???
		gameRoom.Mutex.Lock()
		socketReq.Route()
		gameRoom.Mutex.Unlock()
	}
}
