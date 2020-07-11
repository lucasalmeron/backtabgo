package mongostorage

import (
	"context"
	"log"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeckService struct {
	connection *mongo.Database
}

func NewDeckService(con *mongo.Database) deck.DeckRepository {
	return &DeckService{con}
}

func (service *DeckService) GetDecks(ctx context.Context) ([]deck.Deck, error) {

	cursor, err := service.connection.Collection("decks").Find(ctx, bson.D{{}})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Println(err)
		return nil, err
	}
	var decks []deck.Deck
	for _, dbDeck := range results {
		parseDeck := deck.Deck{
			ID:          dbDeck["_id"].(primitive.ObjectID).Hex(),
			Name:        dbDeck["name"].(string),
			Theme:       dbDeck["theme"].(string),
			CardsLength: len(dbDeck["cards"].(primitive.A)),
		}
		decks = append(decks, parseDeck)
	}
	return decks, nil
}

func (service *DeckService) GetDecksWithCards(ctx context.Context) ([]deck.Deck, error) {
	pipeline := bson.D{
		{"$lookup", bson.D{
			{"from", "cards"},
			{"localField", "cards"},
			{"foreignField", "_id"},
			{"as", "cards"},
		}},
	}
	cursor, err := service.connection.Collection("decks").Aggregate(ctx, mongo.Pipeline{pipeline})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Println(err)
		return nil, err
	}
	var decks []deck.Deck
	for _, dbDeck := range results {
		parseDeck := deck.Deck{
			ID:          dbDeck["_id"].(primitive.ObjectID).Hex(),
			Name:        dbDeck["name"].(string),
			Theme:       dbDeck["theme"].(string),
			CardsLength: len(dbDeck["cards"].(primitive.A)),
			Cards:       map[string]*card.Card{},
		}
		for _, dbCard := range dbDeck["cards"].(primitive.A) {
			primitiveCard := dbCard.(primitive.M)
			parseCard := card.Card{
				ID:   primitiveCard["_id"].(primitive.ObjectID).Hex(),
				Word: primitiveCard["word"].(string),
			}
			for _, fword := range primitiveCard["forbiddenWords"].(primitive.A) {
				parseCard.ForbiddenWords = append(parseCard.ForbiddenWords, fword.(string))
			}
			parseDeck.Cards[parseCard.ID] = &parseCard
		}
		decks = append(decks, parseDeck)
	}
	return decks, nil
}
