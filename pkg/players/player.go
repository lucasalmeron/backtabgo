package player

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	PlayerID uuid.UUID   `json:"playerID"`
	Name     string      `json:"name"`
	Team     int         `json:"team"`
}

type Player struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Team            int       `json:"team"`
	Socket          *websocket.Conn
	GameRoomChannel chan Message
}

func (c *Player) Write(message Message) {
	c.Socket.WriteJSON(message)
}

func (c *Player) Read() {
	defer func() {
		c.Socket.Close()
	}()

	message := Message{Type: "connected", Data: "connection success", PlayerID: c.ID, Name: c.Name, Team: c.Team}

	c.GameRoomChannel <- message

	for {
		c.Socket.SetReadDeadline(time.Now().Add(10 * time.Minute))
		m := Message{}
		err := c.Socket.ReadJSON(&m)
		if err != nil {
			message := Message{Type: "kickPlayerTimeOut", Data: "Time out", PlayerID: c.ID, Name: c.Name, Team: c.Team}
			c.GameRoomChannel <- message
			fmt.Println("TimeOut", err)
			break
		}

		m.PlayerID = c.ID
		m.Name = c.Name
		m.Team = c.Team

		c.GameRoomChannel <- m
		//fmt.Printf("player: %+v\n", c)
		//fmt.Printf("Got message: %#v\n", m)

	}
}
