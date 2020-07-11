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
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
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
	router.Path("/createroom").HandlerFunc(handler.createRoom).Methods(http.MethodGet, http.MethodOptions)
	router.Path("/joinroom/{gameroom}").HandlerFunc(handler.joinRoom)
	router.Path("/reconnectroom/{gameroom}/{playerid}").HandlerFunc(handler.reconnect)

	router.Path("/getdecks").HandlerFunc(handler.getDecks).Methods(http.MethodGet, http.MethodOptions)
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

func (h httpHandler) getDecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	deckRepository := new(deck.Deck)
	dbDecks, err := deckRepository.GetDecksWithCards()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("db error")
		return
	}
	type deck struct {
		ID          string      `json:"id"`
		Name        string      `json:"name"`
		Theme       string      `json:"theme"`
		CardsLength int         `json:"cardsLength"`
		Cards       []card.Card `json:"cards"`
	}
	var decks []deck
	for _, dbDeck := range dbDecks {
		deck := deck{dbDeck.ID, dbDeck.Name, dbDeck.Theme, dbDeck.CardsLength, []card.Card{}}
		for _, card := range dbDeck.Cards {
			deck.Cards = append(deck.Cards, *card)
		}
		decks = append(decks, deck)
	}

	json.NewEncoder(w).Encode(decks)
}
