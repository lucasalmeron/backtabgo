package httphandler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
)

type httpDeckHandler struct{}

func InitDeckHandler(router *mux.Router) {
	handler := new(httpDeckHandler)

	router.Path("/decks").HandlerFunc(handler.GetDecks).Methods(http.MethodGet, http.MethodOptions)
	router.Path("/decks/{deckID:[0-9a-fA-F]{24}}").HandlerFunc(handler.GetDeck).Methods(http.MethodGet, http.MethodOptions)

	router.Path("/decks").HandlerFunc(handler.NewDeck).Methods(http.MethodPost, http.MethodOptions)
	router.Path("/decks").HandlerFunc(handler.UpdateDeck).Methods(http.MethodPut, http.MethodOptions)

}

func (h httpDeckHandler) GetDecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	deckRepository := new(deck.Deck)
	dbDecks, err := deckRepository.GetDecksWithCards()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "db error"})
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

func (h httpDeckHandler) GetDeck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	deckID := mux.Vars(r)["deckID"]

	deckRepository := new(deck.Deck)
	dbDeck, err := deckRepository.GetDeck(deckID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "db error"})
		return
	}

	json.NewEncoder(w).Encode(dbDeck)
}

func (h httpDeckHandler) NewDeck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(r.Body)

	var req deck.RequestDeck
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error unmarshalling request body"})
		return
	}

	deck := new(deck.Deck)
	newDeck, err := deck.NewDeck(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error creating new deck"})
		return
	}

	response := struct {
		ID          string      `json:"id"`
		Name        string      `json:"name"`
		Theme       string      `json:"theme"`
		CardsLength int         `json:"cardsLength"`
		Cards       []card.Card `json:"cards"`
	}{
		ID:          newDeck.ID,
		Name:        newDeck.Name,
		Theme:       newDeck.Theme,
		CardsLength: newDeck.CardsLength,
		Cards:       []card.Card{},
	}

	for _, card := range newDeck.Cards {
		response.Cards = append(response.Cards, *card)
	}

	json.NewEncoder(w).Encode(response)
}

func (h httpDeckHandler) UpdateDeck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(r.Body)

	var req deck.RequestDeck
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error unmarshalling request body"})
		return
	}

	deck := new(deck.Deck)
	newDeck, err := deck.UpdateDeck(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error updating deck"})
		return
	}

	response := struct {
		ID          string      `json:"id"`
		Name        string      `json:"name"`
		Theme       string      `json:"theme"`
		CardsLength int         `json:"cardsLength"`
		Cards       []card.Card `json:"cards"`
	}{
		ID:          newDeck.ID,
		Name:        newDeck.Name,
		Theme:       newDeck.Theme,
		CardsLength: newDeck.CardsLength,
		Cards:       []card.Card{},
	}

	for _, card := range newDeck.Cards {
		response.Cards = append(response.Cards, *card)
	}

	json.NewEncoder(w).Encode(response)
}
