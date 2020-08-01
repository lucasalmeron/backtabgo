package gameroom

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

//Send messages to handler
func (gameRoom *GameRoom) sendMessage(action string, message interface{}, triggerPlayer uuid.UUID) {
	//i should optimize mutex here
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

func (gameRoom *GameRoom) messagesTimeOut(minutes int64, name string) *time.Timer {
	return time.AfterFunc(time.Duration(minutes)*time.Minute, func() {
		fmt.Println("TimeOut >> ", gameRoom.ID, " event >> ", name)
		if gameRoom.GameStatus == "roomPhase" {
			for _, player := range gameRoom.Players {
				player.CloseSocket()
			}
			gameRoom.closePlayersWg.Wait()
			close(gameRoom.IncommingMessagesChannel)
		} else {
			gameRoom.Mutex.Lock()
			gameTimeOut = true
			gameRoom.Mutex.Unlock()
			gameRoom.gameChannel <- false
		}
	})
}

//StartListen channel and wait for player's incomming messages, then it call socket handler to classify
func (gameRoom *GameRoom) StartListenSocketMessages() {
	messagesTimeOut := gameRoom.messagesTimeOut(10, "first")
	gameRoom.Wg.Add(1)
	go func() {
		defer func() {
			messagesTimeOut.Stop()
			gameRoom.Wg.Done()
		}()

		for message := range gameRoom.IncommingMessagesChannel {
			fmt.Println("message ", message)
			//PLAYER'S MESSAGES PROXY
			switch message.Action {
			case "keepAlive":
			case "closeConnection":
				gameRoom.closePlayersWg.Done()
				gameRoom.Wg.Done()
			case "playerDisconnected":
				gameRoom.closePlayersWg.Done()
				gameRoom.Wg.Done()
				messagesTimeOut.Stop()
				messagesTimeOut = gameRoom.messagesTimeOut(10, message.Action)
			}
			gameRoom.sendMessage(message.Action, message.Data, message.PlayerID)
		}
	}()
}
