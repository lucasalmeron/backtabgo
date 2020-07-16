package gameroom

import (
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
	Team1Score               int                          `json:"team1Score"`
	Team2Score               int                          `json:"team2Score"`
	PlayersTeam1             []*player.Player             `json:"team1"`
	PlayersTeam2             []*player.Player             `json:"team2"`
	TeamTurn                 int                          `json:"teamTurn"`
	TurnTime                 int64                        `json:"turnTime"`
	GameTime                 int64                        `json:"gameTime"`
	TurnMistakes             []*TurnMistakes              `json:"teamMistakes"`
	TotalCards               int                          `json:"totalCards"`
	CurrentTurn              *player.Player               `json:"currentTurn"`
	CurrentCard              *card.Card                   `json:"-"`
	GameStatus               string                       `json:"gameStatus"`
	Settings                 *GameSettings                `json:"settings"`
	gameChannel              chan bool
	PlayerConnectedChannel   chan player.Player  `json:"-"`
	IncommingMessagesChannel chan player.Message `json:"-"`
	Wg                       sync.WaitGroup      `json:"-"`
	closePlayersWg           sync.WaitGroup
	Mutex                    sync.Mutex `json:"-"`
}
