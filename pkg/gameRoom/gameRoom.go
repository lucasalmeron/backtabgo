package gameroom

import (
	"math/rand"
	"reflect"
	"sync"

	"github.com/google/uuid"
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
	mapMutex                 sync.RWMutex
	Team1Score               int              `json:"team1Score"`
	Team2Score               int              `json:"team2Score"`
	PlayersTeam1             []*player.Player `json:"team1"`
	PlayersTeam2             []*player.Player `json:"team2"`
	TeamTurn                 int              `json:"teamTurn"`
	TurnTime                 int64            `json:"turnTime"`
	GameTime                 int64            `json:"gameTime"`
	TurnMistakes             []*TurnMistakes  `json:"teamMistakes"`
	TotalCards               int              `json:"totalCards"`
	CurrentTurn              *player.Player   `json:"currentTurn"`
	CurrentCard              *card.Card       `json:"-"`
	GameStatus               string           `json:"gameStatus"`
	Settings                 *GameSettings    `json:"settings"`
	gameChannel              chan bool
	PlayerConnectedChannel   chan player.Player  `json:"-"`
	IncommingMessagesChannel chan player.Message `json:"-"`
	Wg                       sync.WaitGroup      `json:"-"`
	closePlayersWg           sync.WaitGroup
	Mutex                    sync.RWMutex `json:"-"`
}

var (
	gameTimeOut = false
)

//util
func getRandomKeyOfMap(mapI interface{}) interface{} {
	keys := reflect.ValueOf(mapI).MapKeys()

	return keys[rand.Intn(len(keys))].Interface()
}

//NewGameRoom is a "constructor" of GameRoom
func NewGameRoom() *GameRoom {
	gameRoom := &GameRoom{
		ID:                       uuid.New(),
		Players:                  map[uuid.UUID]*player.Player{},
		mapMutex:                 sync.RWMutex{},
		Team1Score:               0,
		Team2Score:               0,
		TeamTurn:                 2,
		CurrentTurn:              &player.Player{},
		GameStatus:               "roomPhase",
		gameChannel:              make(chan bool),
		PlayerConnectedChannel:   make(chan player.Player),
		IncommingMessagesChannel: make(chan player.Message),
		Settings: &GameSettings{
			MaxPoints:      50,
			MaxTurnAttemps: 0,
			TurnTime:       1,
			GameTime:       20,
			Decks:          map[string]*deck.Deck{},
		},
		Wg:             sync.WaitGroup{},
		closePlayersWg: sync.WaitGroup{},
		Mutex:          sync.RWMutex{},
	}
	gameRoom.StartListenSocketMessages()
	return gameRoom
}
