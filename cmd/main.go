package main

import (
	server "github.com/lucasalmeron/backtabgo/internal/server"
)

func main() {

	server := new(server.Server)
	server.Init()
	server.ConnectMongoDB()
	server.StartAndListen()
}
