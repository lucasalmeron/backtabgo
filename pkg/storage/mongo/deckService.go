package mongostorage

import (
	"context"
	"log"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeckService struct {
	database   *mongo.Database
	collection *mongo.Collection
}

func NewDeckService(db *mongo.Database) deck.DeckRepository {
	return &DeckService{db, db.Collection("decks")}
}

func (service *DeckService) GetDecks(ctx context.Context) ([]deck.Deck, error) {

	cursor, err := service.collection.Find(ctx, bson.D{{}})
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
	cursor, err := service.collection.Aggregate(ctx, mongo.Pipeline{pipeline})
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

func (service *DeckService) GetDeck(ctx context.Context, deckID string) (*deck.Deck, error) {

	objectId, err := primitive.ObjectIDFromHex(deckID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"_id", objectId},
			}}},
		bson.D{
			{"$lookup", bson.D{
				{"from", "cards"},
				{"localField", "cards"},
				{"foreignField", "_id"},
				{"as", "cards"},
			}}},
	}

	cursor, err := service.collection.Aggregate(ctx, pipeline)
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
	return &decks[0], nil

}

func (service *DeckService) NewDeck(ctx context.Context, deck deck.Deck) (*deck.Deck, error) {

	var cards []primitive.ObjectID

	for cardKey := range deck.Cards {
		objectID, err := primitive.ObjectIDFromHex(cardKey)
		if err != nil {
			return nil, err
		}
		cards = append(cards, objectID)
	}

	newDeckID, err := service.collection.InsertOne(
		ctx,
		bson.M{
			"name":  deck.Name,
			"theme": deck.Theme,
			"cards": cards,
		},
	)
	if err != nil {
		return nil, err
	}
	dbDeck, err := service.GetDeck(ctx, newDeckID.InsertedID.(primitive.ObjectID).Hex())
	if err != nil {
		return nil, err
	}
	deck = *dbDeck
	return &deck, nil
}

func (service *DeckService) UpdateDeck(ctx context.Context, reqDeck deck.Deck) (*deck.Deck, error) {

	objectID, err := primitive.ObjectIDFromHex(reqDeck.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var cards []primitive.ObjectID

	for cardKey := range reqDeck.Cards {
		objectID, err := primitive.ObjectIDFromHex(cardKey)
		if err != nil {
			return nil, err
		}
		cards = append(cards, objectID)
	}

	var result bson.M

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	err = service.collection.FindOneAndUpdate(ctx, bson.D{{"_id", objectID}}, bson.M{"$set": bson.M{
		"name":  reqDeck.Name,
		"theme": reqDeck.Theme,
		"cards": cards,
	}}, &opt).Decode(&result)
	if err != nil {
		return nil, err
	}
	dbDeck, err := service.GetDeck(ctx, result["_id"].(primitive.ObjectID).Hex())
	if err != nil {
		return nil, err
	}

	return dbDeck, nil
}

func (service *DeckService) DeleteDeck(ctx context.Context, reqDeck deck.Deck) error {
	objectID, err := primitive.ObjectIDFromHex(reqDeck.ID)
	if err != nil {
		log.Println(err)
		return err
	}
	var deletedDocument bson.M
	err = service.collection.FindOneAndDelete(ctx, bson.D{{"_id", objectID}}).Decode(&deletedDocument)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
