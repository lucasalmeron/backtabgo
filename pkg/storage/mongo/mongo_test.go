package mongostorage

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	mongoURI      = os.Getenv("MONGODB_URI")
	mongoDataBase = os.Getenv("MONGODB_DB")
)

func Test_MongoConnection(t *testing.T) {

	err := NewMongoDBConnection(mongoURI, mongoDataBase)
	if err != nil {
		t.Errorf("MongoDb Connection Error: %v", err)
	}
}

func Test_Cards_ABM(t *testing.T) {

	card := new(card.Card)
	card.Word = "Test"
	card.ForbiddenWords = []string{"asd1", "asd2", "asd3"}

	t.Run("GET CARDS", func(t *testing.T) {
		_, err := card.GetCards()
		if err != nil {
			t.Errorf("Error getting Cards: %v", err)
		}
	})

	t.Run("NEW CARD", func(t *testing.T) {
		newCard, err := card.NewCard(*card)
		if err != nil {
			t.Errorf("Creating Card Error: %v", err)
		}

		_, err = primitive.ObjectIDFromHex(newCard.ID)
		if err != nil {
			t.Errorf("Parsing ObjectID Error: %v", err)
		}
		if reflect.TypeOf(newCard.Word).Kind() != reflect.String {
			t.Errorf("Creating card field Error")
		}
		if reflect.TypeOf(newCard.ForbiddenWords).Kind() != reflect.Slice {
			t.Errorf("Creating card field Error")
		}
		card = newCard
	})

	t.Run("GET CARD", func(t *testing.T) {
		cardFound, err := card.GetCard(card.ID)
		if err != nil {
			t.Errorf("Card getting Error: %v", err)
		}

		if !cmp.Equal(cardFound, card) {
			t.Errorf("card found is not equal to expect")
		}
	})

	t.Run("UPDATE CARD", func(t *testing.T) {
		card.Word = "Testing"
		card.ForbiddenWords = append(card.ForbiddenWords, "asd5")
		updatedCard, err := card.UpdateCard(*card)
		if err != nil {
			t.Errorf("Updating Card Error : %v", err)
		}
		if !cmp.Equal(updatedCard, card) {
			t.Errorf("Updated card is not equal to request card")
		}
	})

	/*t.Run("DELETE CARD", func(t *testing.T) {
		err := card.
		if err != nil {
			t.Errorf("Removing Card Error : %v", err)
		}
	})*/

}

func Test_Decks_ABM(t *testing.T) {

	dk := new(deck.Deck)
	dk.Name = "Test"
	dk.Theme = "Test"
	dk.Cards = make(map[string]*card.Card)
	card := new(card.Card)
	cards, err := card.GetCards()
	if err != nil {
		t.Errorf("ABM deck -> getting cards Error: %v", err)
	}

	for i := 0; i < 10; i++ {
		dk.Cards[cards[i].ID] = &cards[i]
	}

	t.Run("GET DECKS", func(t *testing.T) {
		decks, err := dk.GetDecks()
		if err != nil {
			t.Errorf("Error getting Decks: %v", err)
		}

		for _, dbdeck := range decks {
			if reflect.TypeOf(dbdeck) != reflect.TypeOf(*dk) {
				t.Errorf("Getting Decks, Type error")
			}
		}
	})

	t.Run("GET DECKS WITH CARDS", func(t *testing.T) {
		decks, err := dk.GetDecksWithCards()
		if err != nil {
			t.Errorf("Error getting Decks With Cards: %v", err)
		}
		for _, dbdeck := range decks {
			if reflect.TypeOf(dbdeck) != reflect.TypeOf(*dk) {
				t.Errorf("Getting Decks with cards, Type error")
			}
		}
	})

	t.Run("NEW DECK", func(t *testing.T) {

		reqDeck := deck.RequestDeck{
			"",
			dk.Name,
			dk.Theme,
			0,
			[]string{},
		}
		for keyCard := range dk.Cards {
			reqDeck.Cards = append(reqDeck.Cards, keyCard)
		}

		newDeck, err := dk.NewDeck(reqDeck)
		if err != nil {
			t.Errorf("Creating deck Error: %v", err)
		}

		_, err = primitive.ObjectIDFromHex(newDeck.ID)
		if err != nil {
			t.Errorf("Parsing ObjectID Error: %v", err)
		}

		if reflect.TypeOf(newDeck.Name).Kind() != reflect.String {
			t.Errorf("Creating deck field Error")
		}
		if reflect.TypeOf(newDeck.Theme).Kind() != reflect.String {
			t.Errorf("Creating deck field Error")
		}
		for _, c := range newDeck.Cards {
			if reflect.TypeOf(c.Word).Kind() != reflect.String {
				t.Errorf("Creating deck, card field Error")
			}
			if reflect.TypeOf(c.ForbiddenWords).Kind() != reflect.Slice {
				t.Errorf("Creating deck, card field Error")
			}
		}

		dk = newDeck
	})

	t.Run("GET DECK", func(t *testing.T) {
		deckFound, err := dk.GetDeck(dk.ID)

		if err != nil {
			t.Errorf("deck getting Error: %v", err)
		}

		if reflect.TypeOf(deckFound.Name).Kind() != reflect.String {
			t.Errorf("get deck field Error")
		}
		if reflect.TypeOf(deckFound.Theme).Kind() != reflect.String {
			t.Errorf("get deck field Error")
		}
		for _, c := range deckFound.Cards {
			if reflect.TypeOf(c.Word).Kind() != reflect.String {
				t.Errorf("get deck, card field Error")
			}
			if reflect.TypeOf(c.ForbiddenWords).Kind() != reflect.Slice {
				t.Errorf("get deck, card field Error")
			}
		}

		if !cmp.Equal(deckFound, dk) {
			t.Errorf("deck found is not equal to expect")
		}
	})

	t.Run("UPDATE DECK", func(t *testing.T) {

		reqDeck := deck.RequestDeck{
			dk.ID,
			"Testing",
			"Testing",
			dk.CardsLength,
			[]string{},
		}
		for keyCard := range dk.Cards {
			reqDeck.Cards = append(reqDeck.Cards, keyCard)
		}

		updatedDeck, err := dk.UpdateDeck(reqDeck)
		if err != nil {
			t.Errorf("Updating deck Error : %v", err)
		}
		if reflect.TypeOf(updatedDeck.Name).Kind() != reflect.String {
			t.Errorf("get deck field Error")
		}
		if reflect.TypeOf(updatedDeck.Theme).Kind() != reflect.String {
			t.Errorf("get deck field Error")
		}
		for _, c := range updatedDeck.Cards {
			if reflect.TypeOf(c.Word).Kind() != reflect.String {
				t.Errorf("get deck, card field Error")
			}
			if reflect.TypeOf(c.ForbiddenWords).Kind() != reflect.Slice {
				t.Errorf("get deck, card field Error")
			}
		}
		dk.Name = "Testing"
		dk.Theme = "Testing"
		if !cmp.Equal(updatedDeck, dk) {
			t.Errorf("Updated deck is not equal to request deck")
		}
	})

	t.Run("DELETE DECK", func(t *testing.T) {
		err := dk.DeleteDeck(*dk)
		if err != nil {
			t.Errorf("Removing deck Error : %v", err)
		}
	})

}
