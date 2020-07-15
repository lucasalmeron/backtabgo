package httphandler

import (
	"encoding/json"
	"fmt"
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
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "db error"})
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
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "db error"})
		return
	}

	json.NewEncoder(w).Encode(dbCard)
}

func (h httpDeckHandler) newCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(r.Body)

	var card card.Card

	if err := decoder.Decode(&card); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error unmarshalling request body"})
		return
	}

	if err := card.Validate(); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
		return
	}

	newCard, err := card.NewCard(card)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error creating new card"})
		return
	}

	json.NewEncoder(w).Encode(newCard)
}

func (h httpDeckHandler) updateCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(r.Body)

	var card card.Card

	if err := decoder.Decode(&card); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error unmarshalling request body"})
		return
	}

	if err := card.Validate(); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, err.Error()})
		return
	}

	updatedCard, err := card.UpdateCard(card)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&httpError{http.StatusInternalServerError, "Error updating deck"})
		return
	}

	json.NewEncoder(w).Encode(updatedCard)
}
