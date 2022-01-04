package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var CONNECTION_TIMEOUT = 10 * time.Second
var REQUEST_TIMEOUT = 5 * time.Second

// Connect to the Mongo database and return a client
func MongoClient() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), CONNECTION_TIMEOUT)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Find one document in a collection
func FindOne(
	client *mongo.Client,
	databaseName string,
	collectionName string,
	filter bson.D,
) (bson.M, error) {
	coll := client.Database(databaseName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()
	var result bson.M
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Find many documents in a collection
func FindMany(
	client *mongo.Client,
	databaseName string,
	collectionName string,
	filter bson.D,
) (*mongo.Cursor, error) {
	coll := client.Database(databaseName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()
	result, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Insert one new document into a collection
func InsertOne(
	client *mongo.Client,
	databaseName string,
	collectionName string,
	newDocument bson.D,
) (*mongo.InsertOneResult, error) {
	coll := client.Database(databaseName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()
	var result *mongo.InsertOneResult
	result, err := coll.InsertOne(ctx, newDocument)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update one existing document in a collection
func UpdateOne(
	client *mongo.Client,
	databaseName string,
	collectionName string,
	filter bson.D,
	updateOperation bson.D,
) (*mongo.UpdateResult, error) {
	coll := client.Database(databaseName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()
	var result *mongo.UpdateResult
	result, err := coll.UpdateOne(ctx, filter, updateOperation)
	if err != nil {
		return nil, err
	}
	return result, nil
}
