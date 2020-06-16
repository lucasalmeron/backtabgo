package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lucasalmeron/backtabgo/pkg/gameRoom"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
)

var addr = flag.String("addr", "127.0.0.1:3500", "address:port")

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var gameRooms = map[uuid.UUID]gameRoom.GameRoom{}

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.Path("/createroom").HandlerFunc(createRoom).Methods("GET", "OPTIONS")
	router.Path("/joinroom/{gameroom}").HandlerFunc(joinRoom)
	router.Path("/reconnectroom/{gameroom}/{playerid}").HandlerFunc(reconnect)

	srv := &http.Server{
		Handler: router,
		Addr:    *addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}

	// Canal para señalar conexiones inactivas cerradas.
	conxCerradas := make(chan struct{})
	// Lanzamos goroutine para esperar señal y llamar Shutdown.
	go waitForShutdown(conxCerradas, srv)

	fmt.Printf("Server started on %s. CTRL+C for shutdown.\n", *addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe Error: %v", err)
	}
	// Esperamos a que el shut down termine al cerrar todas las conexiones.
	<-conxCerradas
	fmt.Println("Shutdown Success.")
}

// waitForShutdown para detectar señales de interrupción al proceso y hacer Shutdown.
func waitForShutdown(conxCerradas chan struct{}, srv *http.Server) {
	// Canal para recibir señal de interrupción.
	sigint := make(chan os.Signal, 1)
	// Escuchamos por una señal de interrupción del OS (SIGINT).
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	// Si llegamos aquí, recibimos la señal, iniciamos shut down.
	// Noten se puede usar un Context para posible límite de tiempo.
	fmt.Println("\nShutdown started...")
	// Límite de tiempo para el Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		// Error aquí tiene que ser cerrando conexiones.
		log.Printf("Shutdown Error: %v", err)
	}

	// Cerramos el canal, señalando conexiones ya cerradas.
	close(conxCerradas)
}

func createRoom(w http.ResponseWriter, r *http.Request) {

	gameRoom := gameRoom.CreateGameRoom()
	gameRooms[gameRoom.ID] = *gameRoom
	//fmt.Println(gameRooms)

	go gameRoom.StartListen()
	fmt.Println("Goroutines \t", runtime.NumGoroutine())

	var response struct {
		GameRoomID string `json:"gameRoomID"`
	}
	response.GameRoomID = gameRoom.ID.String()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

func joinRoom(w http.ResponseWriter, r *http.Request) {

	fmt.Println("WebSocket Endpoint Hit")
	gameRoomID := mux.Vars(r)["gameroom"]

	key, err := uuid.Parse(gameRoomID)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		//w.WriteHeader(400)
		//json.NewEncoder(w).Encode("{'status':400,'message':'Game room doesn't exist'") //esto hay q hacerlo bien
		return
	}
	if gameRoom, ok := gameRooms[key]; ok {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
		}

		playerNumber := strconv.Itoa(len(gameRoom.Players) + 1)
		player := &player.Player{
			ID:              uuid.New(),
			Name:            "Player " + playerNumber,
			Socket:          conn,
			GameRoomChannel: gameRoom.GameRoomChannel,
		}

		//balance teams
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
		} else {
			player.Team = 1
		}

		//set admin
		if len(gameRoom.Players) == 0 {
			player.Admin = true
		}
		gameRoom.Players[player.ID] = player

		fmt.Println("Goroutines \t", runtime.NumGoroutine())

		player.Read(false)

	}

}

func reconnect(w http.ResponseWriter, r *http.Request) {

	fmt.Println("WebSocket Reconnect")
	gameRoomID := mux.Vars(r)["gameroom"]
	playerID := mux.Vars(r)["playerid"]

	roomKey, err := uuid.Parse(gameRoomID)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}
	playerKey, err := uuid.Parse(playerID)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}

	if gameRoom, ok := gameRooms[roomKey]; ok {

		if player, ok := gameRoom.Players[playerKey]; ok {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				fmt.Fprintf(w, "%+v\n", err)
			}
			player.Socket = conn
			player.Read(true)
		}

	}
}
