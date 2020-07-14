package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	gameroom "github.com/lucasalmeron/backtabgo/pkg/gameRoom"
)

type httpRoomHandler struct{}

var (
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	gameRooms = map[uuid.UUID]*gameroom.GameRoom{}
)

func InitRoomHandler(router *mux.Router) {
	handler := new(httpRoomHandler)

	router.Path("/room/new").HandlerFunc(handler.createRoom).Methods(http.MethodGet, http.MethodOptions)
	router.Path("/room/join/{gameroom}").HandlerFunc(handler.joinRoom)
	router.Path("/room/reconnect/{gameroom}/{playerid}").HandlerFunc(handler.reconnect)
}

///serve static for SPA code///
/*type spaHandler struct {
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
}*/

func (h httpRoomHandler) createRoom(w http.ResponseWriter, r *http.Request) {

	gameRoom := gameroom.NewGameRoom()

	gameRoom.Mutex.Lock()
	gameRooms[gameRoom.ID] = gameRoom
	gameRoom.Mutex.Unlock()

	var response struct {
		GameRoomID string `json:"gameRoomID"`
	}
	response.GameRoomID = gameRoom.ID.String()

	fmt.Println("Goroutines start room -> ", gameRoom.ID, " --> ", runtime.NumGoroutine())

	//this goroutine wait for gameEnded and pop gameRoom of array
	go func() {
		gameRoom.Wg.Wait()
		gameRoom.Mutex.Lock()
		delete(gameRooms, gameRoom.ID)
		gameRoom.Mutex.Unlock()
		fmt.Println("Goroutines close room -> ", gameRoom.ID, " --> ", runtime.NumGoroutine())
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)

}

func (h httpRoomHandler) joinRoom(w http.ResponseWriter, r *http.Request) {

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

func (h httpRoomHandler) reconnect(w http.ResponseWriter, r *http.Request) {

	fmt.Println("WebSocket Reconnect Hit")
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
