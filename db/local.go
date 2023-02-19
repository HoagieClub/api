package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// Local development helpers
func SetupInitialDatabase(client *mongo.Client) error {
	// Create core database
	ctx := context.Background()
	fmt.Println("[i] Creating core database...")
	client.Database("core").CreateCollection(ctx, "users")
	client.Database("apps").CreateCollection(ctx, "stuff")
	client.Database("core").Collection("users").InsertOne(ctx, map[string]interface{}{
		"name":  "Meatball Hoagie",
		"email": "meatball@princeton.edu",
	})

	return nil
}
