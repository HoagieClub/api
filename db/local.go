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
		{Id: "1", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "2", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "3", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "4", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/2OMYXEY.jpeg"},
		{Id: "5", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
		{Id: "6", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "7", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "8", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "9", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "10", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
		{Id: "11", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "12", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "13", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "14", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/vWCQQPb.png"},
		{Id: "15", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
		{Id: "16", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "17", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "18", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "19", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/2OMYXEY.jpeg"},
		{Id: "20", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
		{Id: "21", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "22", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "23", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "24", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/bbZ5Tmj.png"},
		{Id: "25", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
		{Id: "26", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "27", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "28", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "29", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/vWCQQPb.png"},
		{Id: "30", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
		{Id: "31", Title: "Looking for uber sharing", Description: "looking for someone to share an Uber back to campus from Newark airport on Wednesday. My flight is at 3AM. Willing to wait unti 4am.", Category: "bulletin", Tags: []string{"Request"}, User: veggie},
		{Id: "32", Title: "CLOTHING + TECH sale! Moving out!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"clothing", "technology"}, Description: "MOVEOUT SALE! I am currently having a small moveout sale for anyone who is interested! All prices negotiable as I'm hoping for buncha stuff to find new homes", User: buffalo},
		{Id: "33", Title: "Looking for a roommate", Category: "bulletin", Tags: []string{"announcement"}, Description: "I am looking for a roommate to move into my apartment in the fall. The apartment is a 2 bedroom 2 bathroom apartment in the 100 block of Nassau.", User: tomato},
		{Id: "34", Title: "My watch", Category: "lost", Tags: []string{"lost"}, Description: "Was studying in Firestone and noticed I lost my watch, it has a blue strap and red watchface.", User: potato, Thumbnail: "https://i.imgur.com/2OMYXEY.jpeg"},
		{Id: "35", Title: "SELLING MY PC!", Category: "sale", Link: "https://hoagie.io", Tags: []string{"technology"}, Description: "Selling my PC, built it myself 2 years ago, still running great. Email me for specs!", User: potato},
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
