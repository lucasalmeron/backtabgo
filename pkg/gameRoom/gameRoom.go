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
	MaxTurnAttemps int                  `json:"maxTurnAttemps"`
	Decks          map[string]deck.Deck `json:"decks"`
	MaxPoints      int                  `json:"maxPoints"`
	TurnTime       int                  `json:"turnTime"`
	GameTime       int                  `json:"gameTurn"`
}

type GameRoom struct {
	ID                       uuid.UUID                    `json:"id"`
	Players                  map[uuid.UUID]*player.Player `json:"-"`
	Team1Score               int                          `json:"team1Score"`
	Team2Score               int                          `json:"team2Score"`
	PlayersTeam1             []*player.Player             `json:"team1"`
	PlayersTeam2             []*player.Player             `json:"team2"`
	TeamTurn                 int                          `json:"teamTurn"`
	CurrentTurn              *player.Player               `json:"currentTurn"`
	CurrentCard              card.Card                    `json:"currentCard"`
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
			Decks:          map[string]deck.Deck{},
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
			if len(gameRoom.PlayersTeam1)-1 == lastPlayerTeam1Index {
				lastPlayerTeam1Index = 0
			} else {
				lastPlayerTeam1Index++
			}
		} else {
			gameRoom.CurrentTurn = gameRoom.PlayersTeam2[lastPlayerTeam2Index]
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

		fmt.Println("wait for take card")
		<-gameRoom.GameChannel
		fmt.Println("end wait for take card")

		gameRoom.GameStatus = "turnInCourse"

		//TIME TO SEND ATTEMPS
		time.Sleep(time.Duration(gameRoom.Settings.TurnTime) * time.Minute)

		fmt.Println("END TURN")

		//add empty deck
		if gameRoom.Settings.MaxPoints <= gameRoom.Team1Score || gameRoom.Settings.MaxPoints <= gameRoom.Team2Score {
			break
		}
	}
	fmt.Println("game end")
}

func (gameRoom *GameRoom) TakeCard() {
	fmt.Println("ramdomear carta y setear")
}

func (gameRoom *GameRoom) PlayTurn() {

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
