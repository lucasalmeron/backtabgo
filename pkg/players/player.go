package player

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type    int     `json:"type"`
	Message string  `json:"message"`
	Player  *Player `json:"player"`
}

type Player struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Socket          *websocket.Conn `json:"socket"`
	GameRoomChannel chan Message
}

func (c *Player) Write(messageType int, message string) {
	c.Socket.WriteJSON(Message{Type: messageType, Message: message})
}

func (c *Player) Read() {
	defer func() {
		//c.Pool.Unregister <- c
		c.Socket.Close()
	}()

	for {
		messageType, p, err := c.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		message := Message{Type: messageType, Message: string(p), Player: c}
		//c.Pool.Broadcast <- message
		c.GameRoomChannel <- message
		fmt.Printf("player: %+v\n", c)
		fmt.Printf("Message Received: %+v\n", messageType)
		fmt.Printf("Message Received: %+v\n", string(p))

	}
}
