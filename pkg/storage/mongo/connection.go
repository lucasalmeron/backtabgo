package mongostorage

import (
	"context"
	"fmt"
	"log"
	"os"

	deck "github.com/lucasalmeron/backtabgo/pkg/decks"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mgo *MongoDB

type MongoDB struct {
	connection *mongo.Database
}

func NewMongoDBConnection() error {
	mgo = &MongoDB{}

	err := mgo.connect()
	if err != nil {
		return err
	}
	//set deck Repository
	deck.SetRepository(NewDeckService(mgo.connection))
	return nil
}

func GetMongoDBConnection() *MongoDB {
	return mgo
}

func (mongoDB *MongoDB) connect() error {

	var mongoURI string
	var mongoDataBase string
	if os.Getenv("MONGODB_URI") != "" {
		mongoURI = os.Getenv("MONGODB_URI")
	} else {
		mongoURI = fmt.Sprintf("mongodb://localhost:27017")
	}

	if os.Getenv("MONGODB_DB") != "" {
		mongoDataBase = os.Getenv("MONGODB_DB")
	} else {
		mongoDataBase = "taboogame"
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
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

	mongoDB.connection = client.Database(mongoDataBase)
	log.Println("MongoDB connection success")
	return nil
}
