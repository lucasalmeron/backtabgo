package mongostorage

import (
	"context"
	"log"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CardService struct {
	database   *mongo.Database
	collection *mongo.Collection
}

func NewCardService(db *mongo.Database) card.Repository {
	return &CardService{db, db.Collection("cards")}
}

func (service *CardService) GetCards(ctx context.Context) ([]card.Card, error) {

	cursor, err := service.collection.Find(ctx, bson.D{{}})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []card.Card
	if err = cursor.All(ctx, &results); err != nil {
		log.Println(err)
		return nil, err
	}

	return results, nil
}

func (service *CardService) GetCard(ctx context.Context, cardID string) (*card.Card, error) {

	objectId, err := primitive.ObjectIDFromHex(cardID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var card card.Card
	err = service.collection.FindOne(ctx, bson.D{{"_id", objectId}}).Decode(&card)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &card, nil
}

func (service *CardService) NewCard(ctx context.Context, card card.Card) (*card.Card, error) {

	newCardID, err := service.collection.InsertOne(
		ctx,
		bson.M{
			"word":           card.Word,
			"forbiddenWords": card.ForbiddenWords,
		},
	)
	if err != nil {
		return nil, err
	}
	card.ID = newCardID.InsertedID.(primitive.ObjectID).Hex()
	return &card, nil
}

func (service *CardService) UpdateCard(ctx context.Context, reqCard card.Card) (*card.Card, error) {

	objectID, err := primitive.ObjectIDFromHex(reqCard.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var card card.Card

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	err = service.collection.FindOneAndUpdate(ctx, bson.D{{"_id", objectID}}, bson.M{"$set": bson.M{
		"word":           reqCard.Word,
		"forbiddenWords": reqCard.ForbiddenWords,
	}}, &opt).Decode(&card)
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (service *CardService) DeleteCard(ctx context.Context, reqCard card.Card) error {
	return nil
}
