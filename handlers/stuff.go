package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"hoagie-profile/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

// All valid post categories
var postTypes = map[string]bool{
	"sale":        true,
	"selling":     true,
	"lost":        true,
	"bulletin":    true,
	"marketplace": true,
}

// All valid types
var tagTypes = map[string]bool{
	// Bulletin
	"announcement": true,
	"request":      true,
	"opportunity":  true,
	// Lost & Found
	"lost":  true,
	"found": true,
	// Sale
	"accessories": true,
	"clothing":    true,
	"tech":        true,
	"furniture":   true,
	"school":      true,
	"tickets":     true,
	"other":       true,
}

var setupStuffIndex = func() error {
	stuff := client.Database("apps").Collection("stuff")

	model := mongo.IndexModel{
		Keys:    bson.M{"createdAt": 1},
		Options: options.Index().SetExpireAfterSeconds(int32(EXPIRATION_DURATION)),
	}
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()
	stuff.Indexes().DropAll(ctx)
	_, err := stuff.Indexes().CreateOne(ctx, model)
	if err != nil {
		// TODO: handle this better, errors are not necessarily bad
		// index usually exists so this is necessary only once.
		return err
	}
	return nil
}

var getCurrentDigest = func(user string) (PostData, error) {
	var response PostData
	err := db.FindOne(client, "apps", "stuff", bson.D{{"email", user}}, &response)
	if err != nil {
		return PostData{}, fmt.Errorf("error getting digest: %s", err)
	}
	response.Status = "used"
	return response, nil
}

var getAllStuff = func(limit int64, skip int64, category string) ([]PostData, error) {
	var responses []PostData

	// Setup options for database search
	findOptions := options.Find()
	findOptions.SetSort(bson.D{
		{"createdAt", -1},
		// {"category", category},
	})
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)

	query := bson.D{}
	if category != "" {
		if category == "marketplace" {
			query = bson.D{{"category", bson.D{{"$in", []string{"sale", "selling"}}}}}
		} else {
			query = bson.D{{"category", category}}
		}
	}
	// Perform database search
	resultCursor, err := db.FindMany(client, "apps", "stuff", query, findOptions)
	if err != nil {
		return []PostData{}, fmt.Errorf("Error getting stuff from database: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	// Convert from Cursor to PostData
	for resultCursor.Next(ctx) {
		var decodedResult PostData
		err := resultCursor.Decode(&decodedResult)

		if err != nil {
			return []PostData{}, fmt.Errorf("Error getting stuff: %s", err)
		}

		responses = append(responses, decodedResult)
	}

	// Close the Cursor
	defer resultCursor.Close(ctx)

	return responses, nil
}

// GET /stuff/user
var stuffUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))

	if !success {
		http.Error(w, "You do not have access to send digest.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	currentDigest, err := getCurrentDigest(user.Email)
	if err != nil {
		result, _ := json.Marshal(map[string]string{"status": "unused"})
		w.Write(result)
		return
	}

	// Construct user data
	var userData UserData
	userData.Email = user.Email
	userData.Name = user.Name

	// Construct currentDigest
	currentDigest.User = userData
	currentDigest.Status = "used"

	// Create JSON response
	jsonResp, err := json.Marshal(currentDigest)

	if err != nil {
		log.Fatalf("Error happened in response marshalling. %s", err)
	}
	w.Write(jsonResp)
})

// GET /stuff
var stuffAllHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, success := getUser(r.Header.Get("authorization"))

	if !success {
		http.Error(w, "You do not have access to the Hoagie API.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// vars := mux.Vars(r)
	// fmt.Println(vars)
	// limit, limitErr := strconv.ParseInt(vars["limit"], 10, 64)
	// println(limit, limitErr)
	// skip, skipErr := strconv.ParseInt(vars["offset"], 10, 64)
	// println(skip, skipErr)
	limit, limitErr := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	skip, skipErr := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	category := r.URL.Query().Get("category")
	categoryErr := false

	// Ensure selected category, if present, is valid
	if !postTypes[category] && len(category) > 0 {
		categoryErr = true
	}

	if limitErr != nil || skipErr != nil || categoryErr {
		http.Error(w, "Error parsing query parameters.", http.StatusBadRequest)
		return
	}

	// Retrieve relevant data
	allStuff, err := getAllStuff(limit, skip, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create JSON response
	jsonResp, err := json.Marshal(allStuff)

	if err != nil {
		log.Fatalf("Error happened in response marshalling. %s", err)
	}
	w.Write(jsonResp)
})

// POST /stuff
var stuffSendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))
	if !success {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	if len(user.Name) == 0 {
		http.Error(w, `Hoagie has been updated. Please log-out and log-in again.`, http.StatusBadRequest)
	}

	var postReq PostData
	err := json.NewDecoder(r.Body).Decode(&postReq)

	postReq.User.Name = user.Name
	postReq.User.Email = user.Email
	postReq.Sent = false

	if err != nil {
		http.Error(w, "Message did not contain correct fields.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Validation here
	// Ensure type of post is valid
	if !postTypes[postReq.Category] {
		http.Error(w, "Invalid category of the post.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Ensure that tags are valid
	var numTags int = len(postReq.Tags)
	for i := 0; i < numTags; i++ {
		if !tagTypes[postReq.Tags[i]] {
			http.Error(w, "Invalid tag used.", http.StatusBadRequest)
			deleteVisitor(user.Email)
			return
		}
	}

	// Title length
	if postReq.Category == "bulletin" || postReq.Category == "lost" {
		if utf8.RuneCountInString(postReq.Title) < 3 || utf8.RuneCountInString(postReq.Title) > 100 {
			http.Error(w, "Title needs to be between 3 and 100 characters inclusive.", http.StatusBadRequest)
			deleteVisitor(user.Email)
			return
		}
	}

	// Description Length
	if utf8.RuneCountInString(postReq.Description) < 3 || utf8.RuneCountInString(postReq.Description) > 300 {
		http.Error(w, "Description needs to be between 3 and 200 characters inclusive.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Link
	if len(postReq.Link) > 0 {
		if postReq.Category == "lost" {
			if !strings.HasPrefix(postReq.Link, "https://i.imgur.com/") {
				http.Error(w, "Link must be a valid Imgur URL.", http.StatusBadRequest)
				deleteVisitor(user.Email)
				return
			}
		} else if postReq.Category == "sale" {
			if !strings.HasPrefix(postReq.Link, "https://docs.google.com/") {
				http.Error(w, "Link must be a valid Google Slides URL.", http.StatusBadRequest)
				deleteVisitor(user.Email)
				return
			}
		} else {
			http.Error(w, "You cannot include links in this category.", http.StatusBadRequest)
			deleteVisitor(user.Email)
			return
		}
	}

	current, _ := getCurrentDigest(user.Email)
	if current.Title != "" {
		http.Error(w, "You have already an existing digest. Try deleting it and send again.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Add the digest request to the user's digest queue; the MongoDB document decomposes PostData and UserData
	// into their constitutent elements
	db.InsertOne(client, "apps", "stuff", bson.D{
		{"email", user.Email},
		{"user", postReq.User},
		{"id", postReq.Id},
		{"title", postReq.Title},
		{"description", postReq.Description},
		{"thumbnail", postReq.Thumbnail},
		{"category", postReq.Category},
		{"link", postReq.Link},
		{"tags", postReq.Tags},
		{"sent", postReq.Sent},
		{"createdAt", time.Now()},
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}"))
})

// DELETE /stuff/user
var digestDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))
	if !success {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	current, _ := getCurrentDigest(user.Email)
	if current.Title == "" {
		http.Error(w, "You do not have an existing digest message. Please create one first.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Remove the digest request from the user's digest queue
	_, err := db.DeleteOne(client, "apps", "stuff", bson.D{
		{"email", user.Email},
	})
	if err != nil {
		http.Error(w, "You do not have an existing digest message.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}"))
})
