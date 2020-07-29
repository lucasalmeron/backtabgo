package player

func (c *Player) WriteMessage(message Message) {
	c.Socket.Write(message, c.ID)
}

func (c *Player) ReadMessages(reconnect bool, IncommingMessagesChannel chan Message) {

	var message Message

	if reconnect {
		message = Message{Action: "reconnected", Data: "reconnection success", PlayerID: c.ID}
	} else {
		message = Message{Action: "connected", Data: "connection success", PlayerID: c.ID}
	}

	IncommingMessagesChannel <- message

	c.Socket.Read(IncommingMessagesChannel, c.ID)

}

func (c *Player) CloseSocket() {
	c.Socket.CloseSocket()
}
