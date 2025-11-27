package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//  ----  Copied from stuff.go ----

type UserData struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
type PostData struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	Description string `json:"description"`
	// Category of the post
	Category string `json:"category"`
	// Imgur URL to the image
	Thumbnail string `json:"thumbnail"`
	// Link to the post
	Link string `json:"link"`
	// Tags of the post
	Tags   []string `json:"tags"`
	User   UserData `json:"user"`
	Status string   `json:"status"`
	// Sent with Digest or not
	Sent bool                   `json:"sent"`
	Info map[string]interface{} `json:"info"`
	// TO BE REMOVED: KEPT FOR REVERSE-COMPATIBILITY WITH OLD VERSION
	Name string `json:"name"`
}

// ---- End of stuff.go ----

// Local development helpers
func SetupInitialDatabase(client *mongo.Client) error {
	// Create core database
	ctx := context.Background()
	fmt.Println("[i] Creating core database...")
	client.Database("apps").Drop(ctx)
	client.Database("core").Drop(ctx)
	client.Database("core").CreateCollection(ctx, "users")
	client.Database("apps").CreateCollection(ctx, "stuff")
	client.Database("apps").CreateCollection(ctx, "mail")
	client.Database("core").Collection("users").InsertOne(ctx, map[string]interface{}{
		"name":  "Meatball Hoagie",
		"email": "meatball@princeton.edu",
	})

	veggie := UserData{Name: "Veggie Hoagie", Email: "veggie@princeton.edu"}
	buffalo := UserData{Name: "Buffalo Chicken", Email: "buffalo@princeton.edu"}
	tomato := UserData{Name: "Tomato Tomato", Email: "tomato@princeton.edu"}
	potato := UserData{Name: "Potato Tomato", Email: "potato@princeton.edu"}

	posts := []PostData{
		{Id: "1", Title: "Looking for uber sharing", Description: "1", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "2", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "2", User: buffalo},
		{Id: "3", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "3", User: tomato},
		{Id: "4", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "4", User: potato},
		{Id: "5", Title: "5", Category: "lost", Tags: []string{"lost"}, Description: "5", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "6", Title: "Looking for uber sharing", Description: "6", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "7", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "7", User: buffalo},
		{Id: "8", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "8", User: tomato},
		{Id: "9", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "9", User: potato},
		{Id: "10", Title: "10", Category: "lost", Tags: []string{"lost"}, Description: "10", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "11", Title: "Looking for uber sharing", Description: "11", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "12", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "12", User: buffalo},
		{Id: "13", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "13", User: tomato},
		{Id: "14", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "14", User: potato},
		{Id: "15", Title: "15", Category: "lost", Tags: []string{"lost"}, Description: "15", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "16", Title: "Looking for uber sharing", Description: "16", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "17", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "17", User: buffalo},
		{Id: "18", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "18", User: tomato},
		{Id: "19", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "19", User: potato},
		{Id: "20", Title: "20", Category: "lost", Tags: []string{"lost"}, Description: "20", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "21", Title: "Looking for uber sharing", Description: "21", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "22", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "22", User: buffalo},
		{Id: "23", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "23", User: tomato},
		{Id: "24", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "24", User: potato},
		{Id: "25", Title: "25", Category: "lost", Tags: []string{"lost"}, Description: "25", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "26", Title: "Looking for uber sharing", Description: "26", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "27", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "27", User: buffalo},
		{Id: "28", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "28", User: tomato},
		{Id: "29", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "29", User: potato},
		{Id: "30", Title: "30", Category: "lost", Tags: []string{"lost"}, Description: "30", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "31", Title: "Looking for uber sharing", Description: "31", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "32", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "32", User: buffalo},
		{Id: "33", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "33", User: tomato},
		{Id: "34", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "34", User: potato},
		{Id: "35", Title: "35", Category: "lost", Tags: []string{"lost"}, Description: "35", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
	}

	for i, post := range posts {
		InsertOne(client, "apps", "stuff", bson.D{
			{"email", post.Email},
			{"user", post.User},
			{"id", post.Id},
			{"title", post.Title},
			{"description", post.Description},
			{"thumbnail", post.Thumbnail},
			{"category", post.Category},
			{"link", post.Link},
			{"tags", post.Tags},
			{"sent", post.Sent},
			{"createdAt", time.Now().Add(time.Duration(i) * time.Millisecond)},
		})
	}

	return nil
}
