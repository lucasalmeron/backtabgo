package httphandler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
)

type httpCardHandler struct{}

func InitCardHandler(router *mux.Router) {
	handler := new(httpDeckHandler)

	router.Path("/cards").HandlerFunc(handler.getCards).Methods(http.MethodGet, http.MethodOptions)
	router.Path("/cards/{cardID:[0-9a-fA-F]{24}}").HandlerFunc(handler.getCard).Methods(http.MethodGet, http.MethodOptions)

	router.Path("/cards").HandlerFunc(handler.newCard).Methods(http.MethodPost, http.MethodOptions)
	router.Path("/cards").HandlerFunc(handler.updateCard).Methods(http.MethodPut, http.MethodOptions)

}

func (h httpDeckHandler) getCards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	cardRepository := new(card.Card)
	dbCards, err := cardRepository.GetCards()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("db error")
		return
	}

	json.NewEncoder(w).Encode(dbCards)
}

func (h httpDeckHandler) getCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	cardID := mux.Vars(r)["cardID"]

	cardRepository := new(card.Card)
	dbCard, err := cardRepository.GetCard(cardID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("db error")
		return
	}

	json.NewEncoder(w).Encode(dbCard)
}

func (h httpDeckHandler) newCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(r.Body)

	var req card.Card
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error unmarshalling request body")
		return
	}

	card := new(card.Card)
	newCard, err := card.NewCard(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error creating new card")
		return
	}

	json.NewEncoder(w).Encode(newCard)
}

func (h httpDeckHandler) updateCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(r.Body)

	var req card.Card
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error unmarshalling request body")
		return
	}

	card := new(card.Card)
	newCard, err := card.UpdateCard(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error updating new deck")
		return
	}

	json.NewEncoder(w).Encode(newCard)
}
