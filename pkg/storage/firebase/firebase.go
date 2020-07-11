package storage

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type FirebaseDB struct {
	Connection *firebase.App
}

func (firebaseDB *FirebaseDB) Init() {
	opt := option.WithCredentialsFile("/home/lucas/GOAPPS/gametaboo-dbec8-firebase-adminsdk-9xva5-664385bb16.json")
	config := &firebase.Config{
		ProjectID:   "gametaboo-dbec8",
		DatabaseURL: "https://gametaboo-dbec8.firebaseio.com/",
	}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		fmt.Errorf("error initializing app: %v", err)
		return
	}
	firebaseDB.Connection = app
}

func (firebaseDB *FirebaseDB) FindAll(collection string) ([]interface{}, error) {
	ctx := context.Background()
	client, err := firebaseDB.Connection.Firestore(ctx)
	defer client.Close()
	if err != nil {
		fmt.Errorf("error initializing db: %v", err)
		return nil, err
	}
	var data []interface{}
	iterator := client.Collection(collection).Documents(ctx)
	docs, err := iterator.GetAll()
	if err != nil {
		fmt.Errorf("error getting collection: %v", err)
		return nil, err
	}
	for _, doc := range docs {
		docWID := doc.Data()
		docWID["ID"] = doc.Ref.ID
		data = append(data, docWID)
	}
	return data, nil
}
