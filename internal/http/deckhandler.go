package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
)

type httpDeckHandler struct{}

func InitDeckHandler(router *mux.Router) {
	handler := new(httpDeckHandler)

	router.Path("/getdecks").HandlerFunc(handler.getDecks).Methods(http.MethodGet, http.MethodOptions)

}

func (h httpDeckHandler) getDecks(w http.ResponseWriter, r *http.Request) {
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
