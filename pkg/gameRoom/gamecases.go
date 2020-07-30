package gameroom

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (gameRoom *GameRoom) StartGame() {
	gameRoom.Wg.Add(1)
	lastPlayerTeam1Index := 0
	lastPlayerTeam2Index := 0
	//set game time
	gameTime := time.Now()
	gameRoom.GameTime = gameTime.Unix()

	//wait gametime and force end
	messagesTimeOut := gameRoom.messagesTimeOut(gameRoom.GameTime, "GAME TIMEOUT")

	defer func() {
		messagesTimeOut.Stop()
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
	//close player connections before close room
	for _, player := range gameRoom.Players {
		player.CloseSocket()
	}
	gameRoom.closePlayersWg.Wait()
	fmt.Println("game ended")
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
