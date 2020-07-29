package gameroom

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

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
		Team1Score:               0,
		Team2Score:               0,
		TeamTurn:                 1,
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
		Mutex:          sync.Mutex{},
	}
	gameRoom.StartListenSocketMessages()
	return gameRoom
}

func (gameRoom *GameRoom) AddPlayer(conn player.WebSocket) {

	gameRoom.Mutex.Lock()
	playerNumber := strconv.Itoa(len(gameRoom.Players) + 1)
	player := &player.Player{
		ID:     uuid.New(),
		Name:   "Player " + playerNumber,
		Status: "connected",
		Socket: conn,
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
		gameRoom.PlayersTeam2 = append(gameRoom.PlayersTeam2, player)
		gameRoom.Players[player.ID] = player
	} else {
		player.Team = 1
		gameRoom.PlayersTeam1 = append(gameRoom.PlayersTeam1, player)
		gameRoom.Players[player.ID] = player
	}

	gameRoom.Mutex.Unlock()
	gameRoom.Wg.Add(1)
	if gameRoom.GameStatus == "waitingMinPlayers" {
		gameRoom.PlayerConnectedChannel <- *player
	}

	player.ReadMessages(false, gameRoom.IncommingMessagesChannel)

}

func (gameRoom *GameRoom) ReconnectPlayer(conn player.WebSocket, player *player.Player) {
	gameRoom.Mutex.Lock()
	player.Socket = conn
	player.Status = "connected"
	gameRoom.Mutex.Unlock()
	gameRoom.Wg.Add(1)
	if gameRoom.GameStatus == "waitingMinPlayers" {
		gameRoom.PlayerConnectedChannel <- *player
	}
	player.ReadMessages(true, gameRoom.IncommingMessagesChannel)
}

//it check if are minimum 2 players in each team and it stay waiting for reconnects or connects
func (gameRoom *GameRoom) checkMinPlayersConnection() {
	lastState := gameRoom.GameStatus
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
			gameRoom.GameStatus = lastState
			return
		}

		//broadcast Waiting for players
		gameRoom.GameStatus = "waitingMinPlayers"
		gameRoom.sendMessage("waitingForPlayers", gameRoom, uuid.UUID{})

		<-gameRoom.PlayerConnectedChannel
	}
}

func (gameRoom *GameRoom) setNextPlayer(currentIndex1 *int, currentIndex2 *int) {
	if gameRoom.TeamTurn == 2 {
		for {
			if gameRoom.PlayersTeam1[*currentIndex1].Status != "connected" {
				if len(gameRoom.PlayersTeam1)-1 == *currentIndex1 {
					*currentIndex1 = 0
				} else {
					*currentIndex1++
				}
			} else {
				gameRoom.CurrentTurn = gameRoom.PlayersTeam1[*currentIndex1]
				gameRoom.TeamTurn = 1
				if len(gameRoom.PlayersTeam1)-1 == *currentIndex1 {
					*currentIndex1 = 0
				} else {
					*currentIndex1++
				}
				break
			}
			/*if gameRoom.PlayersTeam1[*currentIndex1].Status == "connected" {
				gameRoom.CurrentTurn = gameRoom.PlayersTeam1[*currentIndex1]
				gameRoom.TeamTurn = 1
				if len(gameRoom.PlayersTeam1)-1 == *currentIndex1 {
					*currentIndex1 = 0
				} else {
					*currentIndex1++
				}
				break
			}
			if len(gameRoom.PlayersTeam1)-1 == *currentIndex1 {
				*currentIndex1 = 0
			} else {
				*currentIndex1++
			}*/
		}
	} else {
		for {
			if gameRoom.PlayersTeam1[*currentIndex2].Status != "connected" {
				if len(gameRoom.PlayersTeam2)-1 == *currentIndex2 {
					*currentIndex2 = 0
				} else {
					*currentIndex2++
				}
			} else {
				gameRoom.CurrentTurn = gameRoom.PlayersTeam2[*currentIndex2]
				gameRoom.TeamTurn = 2
				if len(gameRoom.PlayersTeam2)-1 == *currentIndex2 {
					*currentIndex2 = 0
				} else {
					*currentIndex2++
				}
				break
			}
		}
	}
}

func (gameRoom *GameRoom) StartGame() {
	gameRoom.Wg.Add(1)
	lastPlayerTeam1Index := 0
	lastPlayerTeam2Index := 0
	//set game time
	gameTime := time.Now()
	gameRoom.GameTime = gameTime.Unix()

	//wait gametime and force end
	go func() {
		time.Sleep(time.Duration(gameRoom.GameTime) * time.Minute)
		gameRoom.Mutex.Lock()
		gameTimeOut = true
		gameRoom.Mutex.Unlock()
	}()
	defer func() {
		close(gameRoom.PlayerConnectedChannel)
		close(gameRoom.gameChannel)
		close(gameRoom.IncommingMessagesChannel)
		gameRoom.Wg.Done()
	}()
	for {
		//Check end game conditions
		gameRoom.Mutex.Lock()
		if gameRoom.Settings.MaxPoints <= gameRoom.Team1Score ||
			gameRoom.Settings.MaxPoints <= gameRoom.Team2Score ||
			gameRoom.TotalCards == 0 ||
			gameTimeOut {

			gameRoom.GameStatus = "gameEnded"

			gameRoom.Mutex.Unlock()
			//broadcast game is end
			gameRoom.sendMessage("gameEnded", gameRoom, uuid.UUID{})

			break
		}
		gameRoom.Mutex.Unlock()

		//check if are minimum 2 players in each team
		gameRoom.checkMinPlayersConnection()

		//set current player and index for next player
		gameRoom.Mutex.Lock()
		gameRoom.setNextPlayer(&lastPlayerTeam1Index, &lastPlayerTeam2Index)
		gameRoom.Mutex.Unlock()

		gameRoom.Mutex.Lock()
		gameRoom.GameStatus = "gameInCourse"
		gameRoom.Mutex.Unlock()
		//broadcast Next Player Turn
		gameRoom.sendMessage("broadcastNextPlayerTurn", gameRoom.CurrentTurn, uuid.UUID{})

		fmt.Println("waiting for take a card...")
		chanValue := <-gameRoom.gameChannel
		if chanValue {
			//TIME TO SEND ATTEMPS
			time.Sleep(time.Duration(gameRoom.Settings.TurnTime) * time.Minute)

			fmt.Println("END TURN")
		}
	}
	for _, player := range gameRoom.Players {
		gameRoom.closePlayersWg.Add(1)
		player.CloseSocket()
	}
	gameRoom.closePlayersWg.Wait()
	fmt.Println("game ended")
}

func (gameRoom *GameRoom) TakeCard() error {
	if gameRoom.TotalCards > 0 {
		var randKeyDeck string
		var randKeyCard string
		for {
			randKeyDeck = getRandomKeyOfMap(gameRoom.Settings.Decks).(string)
			if gameRoom.Settings.Decks[randKeyDeck].CardsLength > 0 {
				break
			}
		}
		randKeyCard = getRandomKeyOfMap(gameRoom.Settings.Decks[randKeyDeck].Cards).(string)
		card := gameRoom.Settings.Decks[randKeyDeck].Cards[randKeyCard]
		delete(gameRoom.Settings.Decks[randKeyDeck].Cards, randKeyCard)
		gameRoom.Settings.Decks[randKeyDeck].CardsLength--
		gameRoom.TotalCards--
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
		return nil
	}
	return fmt.Errorf("Empty Decks")

}

func (gameRoom *GameRoom) PlayTurn() {
	//check if are minimum 2 players in each team
	gameRoom.checkMinPlayersConnection()

	fmt.Println("taken card, turn in course")
	turnTime := time.Now()
	gameRoom.TurnTime = turnTime.Unix()

	gameRoom.GameStatus = "turnInCourse"

	err := gameRoom.TakeCard()
	if err != nil {
		gameRoom.gameChannel <- false
		return
	}
	gameRoom.gameChannel <- true
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
	gameRoom.Mutex.Lock()
	socketReq := SocketRequest{
		message: player.Message{
			Action:   action,
			Data:     message,
			PlayerID: triggerPlayer,
		},
		gameRoom: gameRoom,
	}
	socketReq.Route()
	gameRoom.Mutex.Unlock()
}

var (
	sctx          context.Context
	cancelTimeOut context.CancelFunc
)

func (gameRoom *GameRoom) messagesTimeOut() {
	if cancelTimeOut != nil {
		cancelTimeOut()
	}

	sctx, cancelTimeOut = context.WithTimeout(context.TODO(), 10*time.Minute)
	go func() {
		tStart := time.Now()
		<-sctx.Done() // will sit here until the timeout or cancelled
		tStop := time.Now()

		if tStop.Sub(tStart) >= time.Duration(1)*time.Minute {
			gameRoom.Mutex.Lock()
			gameTimeOut = true
			gameRoom.Mutex.Unlock()
			gameRoom.gameChannel <- false
		}
	}()

}

//StartListen channel and wait for player's incomming messages, then it call socket handler to classify
func (gameRoom *GameRoom) StartListenSocketMessages() {
	gameRoom.Wg.Add(1)
	go func() {
		defer gameRoom.Wg.Done()
		for message := range gameRoom.IncommingMessagesChannel {
			fmt.Println("message ", message)

			if message.Action != "keepAlive" {
				gameRoom.messagesTimeOut()
			}
			if message.Action == "closeConnection" {
				gameRoom.closePlayersWg.Done()
				gameRoom.Wg.Done()
			}
			if message.Action == "playerDisconnected" {
				gameRoom.Wg.Done()
			} else {
				gameRoom.sendMessage(message.Action, message.Data, message.PlayerID)
			}
		}
	}()
}
