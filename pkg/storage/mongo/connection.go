package mongostorage

import (
	"context"
	"log"

	card "github.com/lucasalmeron/backtabgo/pkg/cards"
	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mgo *MongoDB

type MongoDB struct {
	connection *mongo.Database
	mongoURI   string
	database   string
}

func NewMongoDBConnection(mongoURI string, database string) error {
	mgo = new(MongoDB)

	mgo.mongoURI = mongoURI
	mgo.database = database

	err := mgo.connect()
	if err != nil {
		return err
	}
	//set deck Repository
	deck.SetRepository(NewDeckService(mgo.connection))
	card.SetRepository(NewCardService(mgo.connection))
	return nil
}

func GetMongoDBConnection() *MongoDB {
	return mgo
}

func (mongoDB *MongoDB) connect() error {

	clientOpts := options.Client().ApplyURI(mgo.mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Check the connections
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	mongoDB.connection = client.Database(mgo.database)
	log.Println("MongoDB connection success")
	return nil
}
