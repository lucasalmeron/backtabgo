package gameRoom

import (
	"fmt"
	"time"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
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

type Pool struct {
	Register   chan *player.Player
	Unregister chan *player.Player
	Players    map[*player.Player]bool
	//Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *player.Player),
		Unregister: make(chan *player.Player),
		Players:    make(map[*player.Player]bool),
		//Broadcast:  make(chan Message),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case player := <-pool.Register:
			pool.Players[player] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Players))
			/*for client, _ := range pool.Players {
				fmt.Println(client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}*/
			break
		case player := <-pool.Unregister:
			delete(pool.Players, player)
			fmt.Println("Size of Connection Pool: ", len(pool.Players))
			/*for player, _ := range pool.Players {
				player.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}*/
			break
			/*case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}*/
		}
	}
}
