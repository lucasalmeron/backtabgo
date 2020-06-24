package storage

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mgo *MongoDB

type MongoDB struct {
	connection *mongo.Database
}

func NewMongoDBConnection() error {
	mgo = &MongoDB{}
	return mgo.connect()
}

func GetMongoDBConnection() *MongoDB {
	return mgo
}

func (mongoDB *MongoDB) connect() error {
	host := "localhost"
	port := 27017

	clientOpts := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", host, port))
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
	mongoDB.connection = client.Database("taboogame")
	log.Println("MongoDB connection success")
	return nil
}

func (mongoDB *MongoDB) Aggregate(collection string, pipeline bson.D) ([]bson.M, error) {

	cursor, err := mongoDB.connection.Collection(collection).Aggregate(context.TODO(), mongo.Pipeline{pipeline})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cursor.Close(context.TODO())
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Println(err)
		return nil, err
	}
	return results, nil
}

func (mongoDB *MongoDB) FindAll(collection string) ([]bson.M, error) {

	cursor, err := mongoDB.connection.Collection(collection).Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cursor.Close(context.TODO())
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Println(err)
		return nil, err
	}
	return results, nil
}

func (mongoDB *MongoDB) FindOne(collection string, filter bson.D) (interface{}, error) {
	var s interface{}
	err := mongoDB.connection.Collection(collection).FindOne(context.TODO(), filter).Decode(&s)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return s, nil
}

func (mongoDB *MongoDB) InsertOne(collection string, entity interface{}) (interface{}, error) {
	insertedDocument, err := mongoDB.connection.Collection(collection).InsertOne(context.TODO(), entity)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return insertedDocument, nil
}
