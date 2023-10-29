package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

  var database = "tarea2"

  func GetCollection(collection string) *mongo.Collection {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://localhost:27017/").SetServerAPIOptions(serverAPI)
  
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
	  panic(err)
	}
  
	// defer func() {
	//   if err = client.Disconnect(context.TODO()); err != nil {
	// 	panic(err)
	//   }
	// }()
  
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
	  panic(err)
	}
	//fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client.Database(database).Collection(collection)
  }
  