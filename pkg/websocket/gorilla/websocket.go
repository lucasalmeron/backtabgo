package gorillawebsocket

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

type Socket struct {
	Connection *websocket.Conn
}

func Connect(w http.ResponseWriter, r *http.Request) (player.WebSocket, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Socket{conn}, nil
}

func (c *Socket) Write(message player.Message, playerTrigger uuid.UUID) error {
	message.PlayerID = playerTrigger
	return c.Connection.WriteJSON(message)
}

func (c *Socket) Read(IncommingMessagesChannel chan player.Message, playerTrigger uuid.UUID) {
	defer c.CloseSocket()
	var message player.Message

	c.Connection.SetReadDeadline(time.Now().Add(10 * time.Minute))

	for {

		err := c.Connection.ReadJSON(&message)
		if err != nil {
			if ok := strings.Contains(err.Error(), "timeout"); ok {
				message = player.Message{Action: "playerDisconnected", Data: "Player kicked due to connection timeout", PlayerID: playerTrigger}
				IncommingMessagesChannel <- message
				fmt.Println("TimeOut", err)
				break
			}
			if ok := strings.Contains(err.Error(), "websocket: close 1005 (no status)"); ok {
				message = player.Message{Action: "playerDisconnected", Data: "Player Disconnected", PlayerID: playerTrigger}
				IncommingMessagesChannel <- message
				fmt.Println("Disconnected", err)
				break
			}
			if ok := strings.Contains(err.Error(), "websocket: close 1001 (going away)"); ok {
				message = player.Message{Action: "playerDisconnected", Data: "Player Disconnected", PlayerID: playerTrigger}
				IncommingMessagesChannel <- message
				fmt.Println("Disconnected", err)
				break
			}
			if ok := strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure): unexpected EOF"); ok {
				message = player.Message{Action: "playerDisconnected", Data: "Player connection closed", PlayerID: playerTrigger}
				IncommingMessagesChannel <- message
				fmt.Println("Disconnected", err)
				break
			}
			if ok := strings.Contains(err.Error(), "use of closed network connection"); ok {
				message = player.Message{Action: "closeConnection", Data: "Close connection", PlayerID: playerTrigger}
				IncommingMessagesChannel <- message
				fmt.Println("closeConnection", err)
				break
			}
			fmt.Printf("unexpected type %T", err)
			fmt.Println("Error ", err)
		}

		message.PlayerID = playerTrigger

		IncommingMessagesChannel <- message

	}
}

func (c *Socket) CloseSocket() error {
	return c.Connection.Close()
}
