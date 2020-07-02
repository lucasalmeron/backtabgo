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

type httpHandler struct{}

var (
	router    *mux.Router
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	gameRooms = map[uuid.UUID]*gameroom.GameRoom{}
)

func Init() *mux.Router {
	handler := new(httpHandler)
	router = mux.NewRouter().StrictSlash(true)
	router.Path("/createroom").HandlerFunc(handler.createRoom).Methods("GET", "OPTIONS")
	router.Path("/joinroom/{gameroom}").HandlerFunc(handler.joinRoom)
	router.Path("/reconnectroom/{gameroom}/{playerid}").HandlerFunc(handler.reconnect)
	spa := spaHandler{StaticPath: "static", IndexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)
	return router
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

func (h httpHandler) createRoom(w http.ResponseWriter, r *http.Request) {

	gameRoom := gameroom.CreateGameRoom()
	gameRooms[gameRoom.ID] = gameRoom
	//fmt.Println(gameRooms)

	gameRoom.Wg.Add(1)
	go gameRoom.StartListenSocketMessages()

	var response struct {
		GameRoomID string `json:"gameRoomID"`
	}
	response.GameRoomID = gameRoom.ID.String()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
	gameRoom.Wg.Wait()
	fmt.Println("Goroutines start room -> ", gameRoom.ID, " --> ", runtime.NumGoroutine())
	delete(gameRooms, gameRoom.ID)
	fmt.Println("Goroutines close room -> ", gameRoom.ID, " --> ", runtime.NumGoroutine())
}

func (h httpHandler) joinRoom(w http.ResponseWriter, r *http.Request) {

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

func (h httpHandler) reconnect(w http.ResponseWriter, r *http.Request) {

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
