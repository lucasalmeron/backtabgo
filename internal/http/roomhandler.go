package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	gameroom "github.com/lucasalmeron/backtabgo/pkg/gameRoom"
	gorillawebsocket "github.com/lucasalmeron/backtabgo/pkg/websocket/gorilla"
)

type httpError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type httpRoomHandler struct{}

var (
	gameRooms = map[uuid.UUID]*gameroom.GameRoom{}
)

func InitRoomHandler(router *mux.Router) {
	handler := new(httpRoomHandler)

	router.Path("/room/new").HandlerFunc(handler.Create).Methods(http.MethodGet, http.MethodOptions)
	router.Path("/room/join/{gameroom}").HandlerFunc(handler.Join)
	router.Path("/room/reconnect/{gameroom}/{playerid}").HandlerFunc(handler.Reconnect)
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

func (h httpRoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	gameRoom := gameroom.NewGameRoom()

	gameRoom.Mutex.Lock()
	gameRooms[gameRoom.ID] = gameRoom
	gameRoom.Mutex.Unlock()

	fmt.Println("Goroutines new room -> ", gameRoom.ID, " --> ", runtime.NumGoroutine())

	//this goroutine wait for gameEnded and pop gameRoom of array
	go func() {
		gameRoom.Wg.Wait()
		gameRoom.Mutex.Lock()
		delete(gameRooms, gameRoom.ID)
		gameRoom.Mutex.Unlock()
		fmt.Println("Goroutines closed room -> ", gameRoom.ID, " --> ", runtime.NumGoroutine())
	}()

	json.NewEncoder(w).Encode(struct {
		GameRoomID string `json:"gameRoomID"`
	}{gameRoom.ID.String()})

}

func (h httpRoomHandler) Join(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	fmt.Println("WebSocket Endpoint Hit")
	gameRoomID := mux.Vars(r)["gameroom"]

	key, err := uuid.Parse(gameRoomID)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
		return
	}
	if gameRoom, ok := gameRooms[key]; ok {
		socket, err := gorillawebsocket.Connect(w, r)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
			return
		}
		//Here it will wait for incomming messages
		gameRoom.AddPlayer(socket)

	} else {
		json.NewEncoder(w).Encode(&httpError{http.StatusBadRequest, "Game Room doesn't exist"})
		return
	}

}

func (h httpRoomHandler) Reconnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	fmt.Println("WebSocket Reconnect Hit")
	gameRoomID := mux.Vars(r)["gameroom"]
	playerID := mux.Vars(r)["playerid"]

	roomKey, err := uuid.Parse(gameRoomID)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
		return
	}
	playerKey, err := uuid.Parse(playerID)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
		return
	}

	if gameRoom, ok := gameRooms[roomKey]; ok {

		if player, ok := gameRoom.Players[playerKey]; ok {
			socket, err := gorillawebsocket.Connect(w, r)
			if err != nil {
				fmt.Fprintf(w, "%+v\n", err)
				json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
				return
			}

			//Here it will wait for incomming messages
			gameRoom.ReconnectPlayer(socket, player)
		}

	} else {
		json.NewEncoder(w).Encode(&httpError{http.StatusBadRequest, "Game Room doesn't exist"})
		return
	}
}
