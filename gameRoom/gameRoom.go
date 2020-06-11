package gameRoom

import (
	"time"

	card "github.com/lucasalmeron/backtabgo/cards"
	deck "github.com/lucasalmeron/backtabgo/decks"
	player "github.com/lucasalmeron/backtabgo/players"
)

type GameRoom struct {
	ID             int             `json:"id"`
	Team1          []player.Player `json:"team1"`
	Team2          []player.Player `json:"team2"`
	Team1Score     int             `json:"team1Score"`
	Team2Score     int             `json:"team2Score"`
	CurrentTurn    player.Player   `json:"currentTurn"`
	TurnTime       time.Time       `json:"turnTime"`
	GameTime       time.Time       `json:"gameTurn"`
	MaxTurnAttemps int             `json:"maxTurnAttemps"`
	Decks          []deck.Deck     `json:"decks"`
	CurrentCard    card.Card       `json:"currentCard"`
	JoinLink       string          `json:"joinLink"`
	MaxPoints      int             `json:"maxPoints"`
}
