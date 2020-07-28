package httphandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	player "github.com/lucasalmeron/backtabgo/pkg/players"
	mongostorage "github.com/lucasalmeron/backtabgo/pkg/storage/mongo"
)

var (
	mongoURI      = os.Getenv("MONGODB_URI")
	mongoDataBase = os.Getenv("MONGODB_DB")
	gameRoomID    = ""
)

func Test_new_room(t *testing.T) {
	if err := mongostorage.NewMongoDBConnection(mongoURI, mongoDataBase); err != nil {
		t.Fatal(err)
	}
	defer mongostorage.GetMongoDBConnection().CloseConnection()

	req := httptest.NewRequest("GET", "/room/new", nil)

	w := httptest.NewRecorder()

	roomHandler := new(httpRoomHandler)
	roomHandler.Create(w, req)

	resp := w.Result()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Error("Request Error")
	}

	gameRoom := struct {
		GameRoomID string `json:"gameRoomID"`
	}{}

	json.Unmarshal(body, &gameRoom)
	if reflect.TypeOf(gameRoom.GameRoomID).Kind() != reflect.String {
		t.Errorf("Creating new room Error")
	}
	if gameRoom.GameRoomID == "" {
		t.Errorf("Creating new room Error")
	}
	gameRoomID = gameRoom.GameRoomID
}

var (
	numberOfPlayers = 4
	players         []player.Player
	playerAdmin     *player.Player
	decks           []deck.Deck
	Once            sync.Once
)

func Test_socket(t *testing.T) {
	if err := mongostorage.NewMongoDBConnection(mongoURI, mongoDataBase); err != nil {
		t.Fatal(err)
	}
	//defer mongostorage.GetMongoDBConnection().CloseConnection()
	router := mux.NewRouter().StrictSlash(true)
	InitRoomHandler(router)
	// Create test server.
	s := httptest.NewServer(router)
	//defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http") + "/room/join/" + gameRoomID

	for i := 1; i <= numberOfPlayers; i++ {

		t.Run(fmt.Sprintf("WEB SOCKET CONNECTION -> %v", i), func(t *testing.T) {

			t.Parallel()
			// Connect to the server
			playerSocket, _, err := websocket.DefaultDialer.Dial(u, nil)
			if err != nil {
				t.Fatalf("Error on connect player: %v", err)
			}

			defer playerSocket.Close()

			var message player.Message
			//go func() {

			for {
				_, msj, err := playerSocket.ReadMessage()
				if err != nil {
					t.Error("read:", err)
					return
				}
				json.Unmarshal(msj, &message)
				handleSocketMessages(&message, playerSocket, t)
			}
		})
	}

}

func handleSocketMessages(message *player.Message, playerSocket *websocket.Conn, t *testing.T) {
	switch message.Action {
	case "connected":
		var pl player.Player
		j, _ := json.Marshal(message.Data)
		json.Unmarshal(j, &pl)
		pl.Socket = playerSocket
		if pl.Admin {
			playerAdmin = &pl
		}
		if pl.Status != "connected" {
			t.Fatal("connect Error")
		}
		players = append(players, pl)
		t.Logf("CONNECTED: %v", len(players))
		if len(players) == numberOfPlayers {
			t.Run("GET DECKS", func(t *testing.T) {
				playerAdmin.Socket.WriteJSON(player.Message{"getDecks", nil, uuid.UUID{}})
			})
		}
	case "playerConnected":
		var player player.Player
		j, _ := json.Marshal(message.Data)
		json.Unmarshal(j, &player)

		//t.Logf("PLAYER CONNECTED: %v", player)
	case "gameStatus":

		//t.Logf("PLAYER CONNECTED: %v", player)
	case "getDecks":
		for _, d := range message.Data.([]interface{}) {
			var dk deck.Deck
			j, _ := json.Marshal(d)
			json.Unmarshal(j, &dk)
			decks = append(decks, dk)
		}
		if len(decks) == 0 {
			t.Fatal("Error getting decks")
		}
		t.Run("UPDATE GAME OPTIONS", func(t *testing.T) {
			var stringDecks []string
			for _, strDeck := range decks {
				stringDecks = append(stringDecks, strDeck.ID)
			}

			var opt = struct {
				MaxTurnAttemps int      `json:"maxTurnAttemps"`
				Decks          []string `json:"decks"`
				MaxPoints      int      `json:"maxPoints"`
				TurnTime       int      `json:"turnTime"`
				GameTime       int      `json:"gameTurn"`
			}{0, stringDecks, 20, 1, 60}
			playerAdmin.Socket.WriteJSON(player.Message{"updateRoomOptions", opt, uuid.UUID{}})
		})
	case "updateRoomOptions":

		Once.Do(func() {
			t.Logf("NUMBER OF PLAYERS: %v", len(players))
			t.Run("START GAME", func(t *testing.T) {
				for _, p := range players {
					t.Logf("status: %v", p.Status)
				}
				playerAdmin.Socket.WriteJSON(player.Message{"startGame", nil, uuid.UUID{}})
			})
		})
	case "nextPlayerTurn":
		t.Logf("nxtTurn: %v", message.Data)
	default:
		//t.Logf("OTHER EVENT: %v", message)
	}
}
