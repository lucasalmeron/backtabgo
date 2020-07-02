package player

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Action   string      `json:"action"`
	Data     interface{} `json:"data"`
	PlayerID uuid.UUID   `json:"triggerPlayer"`
}

type Player struct {
	ID                       uuid.UUID       `json:"id"`
	Name                     string          `json:"name"`
	Team                     int             `json:"team"`
	Admin                    bool            `json:"admin"`
	Status                   string          `json:"status"`
	Socket                   *websocket.Conn `json:"-"`
	IncommingMessagesChannel chan Message    `json:"-"`
}

func (c *Player) Write(message Message) {
	c.Socket.WriteJSON(message)
}

func (c *Player) Read(reconnect bool) {
	defer c.Socket.Close()
	var message Message

	if reconnect {
		message = Message{Action: "reconnected", Data: "reconnection success", PlayerID: c.ID}
	} else {
		message = Message{Action: "connected", Data: "connection success", PlayerID: c.ID}
	}

	c.IncommingMessagesChannel <- message

	c.Socket.SetReadDeadline(time.Now().Add(10 * time.Minute))

	for {

		m := Message{}
		err := c.Socket.ReadJSON(&m)
		if err != nil {
			if ok := strings.Contains(err.Error(), "timeout"); ok {
				message = Message{Action: "playerDisconnected", Data: "Player kicked due to connection timeout", PlayerID: c.ID}
				c.Status = "disconnected"
				c.IncommingMessagesChannel <- message
				fmt.Println("TimeOut", err)
				break
			}
			if ok := strings.Contains(err.Error(), "websocket: close 1005 (no status)"); ok {
				message = Message{Action: "playerDisconnected", Data: "Player Disconnected", PlayerID: c.ID}
				c.Status = "disconnected"
				c.IncommingMessagesChannel <- message
				fmt.Println("Disconnected", err)
				break
			}
			if ok := strings.Contains(err.Error(), "websocket: close 1001 (going away)"); ok {
				message = Message{Action: "playerDisconnected", Data: "Player Disconnected", PlayerID: c.ID}
				c.Status = "disconnected"
				c.IncommingMessagesChannel <- message
				fmt.Println("Disconnected", err)
				break
			}
			if ok := strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure): unexpected EOF"); ok {
				message = Message{Action: "playerDisconnected", Data: "Player connection closed", PlayerID: c.ID}
				c.Status = "disconnected"
				c.IncommingMessagesChannel <- message
				fmt.Println("Disconnected", err)
				break
			}
			fmt.Printf("unexpected type %T", err)
			fmt.Println("Error ", err)
		}

		m.PlayerID = c.ID

		c.IncommingMessagesChannel <- m
		//fmt.Printf("player: %+v\n", c)
		fmt.Printf("Got message: %#v\n", m)

	}
}
