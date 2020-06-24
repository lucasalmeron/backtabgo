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
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	gameroom "github.com/lucasalmeron/backtabgo/pkg/gameRoom"
	storage "github.com/lucasalmeron/backtabgo/pkg/storage"
)

var addr = flag.String("addr", "127.0.0.1:3500", "address:port")

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var gameRooms = map[uuid.UUID]*gameroom.GameRoom{}

func main() {

	err := storage.NewMongoDBConnection()
	if err != nil {
		log.Fatal("MongoDb Connection Error: %v", err)
	}

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

	gameRoom := gameroom.CreateGameRoom()
	gameRooms[gameRoom.ID] = gameRoom
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
		return
	}
	if gameRoom, ok := gameRooms[key]; ok {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
		}

		//Here it will wait for incomming messages
		gameRoom.AddPlayer(conn)

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

			//Here it will wait for incomming messages
			gameRoom.ReconnectPlayer(conn, player)
		}

	}
}
