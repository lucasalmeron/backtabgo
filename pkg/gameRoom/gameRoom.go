package gameroom

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	Wg                       sync.WaitGroup
	Mutex                    sync.Mutex
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
		Wg:    sync.WaitGroup{},
		Mutex: sync.Mutex{},
	}
}

func (gameRoom *GameRoom) AddPlayer(conn *websocket.Conn) {
	//Add new go routine to waitgroup per player
	gameRoom.Wg.Add(1)

	playerNumber := strconv.Itoa(len(gameRoom.Players) + 1)
	player := &player.Player{
		ID:                       uuid.New(),
		Name:                     "Player " + playerNumber,
		Socket:                   conn,
		IncommingMessagesChannel: gameRoom.IncommingMessagesChannel,
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
	} else {
		player.Team = 1
	}

	//set admin
	if len(gameRoom.Players) == 0 {
		player.Admin = true
	}
	gameRoom.Mutex.Lock()
	gameRoom.Players[player.ID] = player
	gameRoom.Mutex.Unlock()
	player.Read(false)
	defer gameRoom.Wg.Done()
}

func (gameRoom *GameRoom) ReconnectPlayer(conn *websocket.Conn, player *player.Player) {
	//Add new go routine to waitgroup per player
	gameRoom.Wg.Add(1)
	gameRoom.Mutex.Lock()
	player.Socket = conn
	gameRoom.Mutex.Unlock()
	player.Read(true)
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
		//could i lock each action on socket handler???
		gameRoom.Mutex.Lock()
		socketReq.Route()
		gameRoom.Mutex.Unlock()
	}
}
