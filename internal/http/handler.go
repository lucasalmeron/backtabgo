package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	gameroom "github.com/lucasalmeron/backtabgo/pkg/gameRoom"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var gameRooms = map[uuid.UUID]*gameroom.GameRoom{}

type httpHandler struct {
	router *mux.Router
}

func InitHttpHandler() *mux.Router {
	httpRouter := &httpHandler{
		router: mux.NewRouter().StrictSlash(true),
	}
	httpRouter.router.Path("/createroom").HandlerFunc(createRoom).Methods("GET", "OPTIONS")
	httpRouter.router.Path("/joinroom/{gameroom}").HandlerFunc(joinRoom)
	httpRouter.router.Path("/reconnectroom/{gameroom}/{playerid}").HandlerFunc(reconnect)
	spa := spaHandler{StaticPath: "static", IndexPath: "index.html"}
	httpRouter.router.PathPrefix("/").Handler(spa)
	return httpRouter.router
}

type spaHandler struct {
	StaticPath string
	IndexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = filepath.Join(h.StaticPath, path)

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.StaticPath, h.IndexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.StaticPath)).ServeHTTP(w, r)
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
