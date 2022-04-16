package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"hoagie-profile/db"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserData struct {
	Email string
	Name  string
	Phone string
}

type PostData struct {
	Id          string
	Title       string
	Description string
	// Type of the post
	TypePost string
	// Imgur URL to the image
	Thumbnail string
	// Link to the post
	Link string
	// Tags of the post
	Tags   []string
	User   UserData
	Status string
}

var postTypes = map[string]bool{
	"sale":     true,
	"selling":  true,
	"lost":     true,
	"bulletin": true,
}

var tagTypes = map[string]bool{
	"tech":          true,
	"clothing":      true,
	"help":          true,
	"opportunities": true,
	"lost":          true,
	"found":         true,
}

var getCurrentDigest = func(user string) (PostData, error) {
	var response PostData
	err := db.FindOne(client, "apps", "mail", bson.D{{"email", user}}, &response)
	if err != nil {
		return PostData{}, fmt.Errorf("error getting digest: %s", err)
	}
	response.Status = "used"
	return response, nil
}

var getAllStuff = func() ([]PostData, error) {
	var responses []PostData
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})

	resultCursor, err := db.FindMany(client, "apps", "stuff", bson.D{}, findOptions)
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
var digestStatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))

	if !success {
		http.Error(w, "You do not have access to send digest.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	currentDigest, err := getCurrentDigest(user.Email)
	if err != nil || currentDigest.Title == "" {
		result, _ := json.Marshal(map[string]string{"Status": "unused"})
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
var stuffGeneralHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, success := getUser(r.Header.Get("authorization"))

	if !success {
		http.Error(w, "You do not have access to the Hoagie API.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	allStuff, err := getAllStuff()
	if err != nil {
		http.Error(w, "You do not have access to stuff.", http.StatusBadRequest)
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
var digestSendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		http.Error(w, "Message did not contain correct fields.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}
	// Validation here
	// Ensure type of post is valid
	if !postTypes[postReq.TypePost] {
		http.Error(w, "Invalid type, try again later.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Ensure that tags are valid
	var numTags int = len(postReq.Tags)
	for i := 0; i < numTags; i++ {
		if !tagTypes[postReq.Tags[i]] {
			http.Error(w, "Invalid tag, try again later.", http.StatusBadRequest)
			deleteVisitor(user.Email)
			return
		}
	}

	// Title length
	if utf8.RuneCountInString(postReq.Title) < 3 || utf8.RuneCountInString(postReq.Title) > 100 {
		http.Error(w, "Title needs to be between 3 and 100 characters inclusive.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Description Length
	if utf8.RuneCountInString(postReq.Description) < 3 || utf8.RuneCountInString(postReq.Description) > 300 {
		http.Error(w, "Description needs to be between 3 and 200 characters inclusive.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	// Link
	if len(postReq.Link) > 0 {
		if postReq.TypePost == "lost" {
			if !strings.HasPrefix(postReq.Link, "https://i.imgur.com/") {
				http.Error(w, "Link must be a valid Imgur URL.", http.StatusBadRequest)
				deleteVisitor(user.Email)
				return
			}
		} else if postReq.TypePost == "sale" {
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
		{"name", user.Name},
		{"id", postReq.Id},
		{"title", postReq.Title},
		{"description", postReq.Description},
		{"thumbnail", postReq.Thumbnail},
		{"typePost", postReq.TypePost},
		{"link", postReq.Link},
		{"tags", postReq.Tags},
		{"created_at", time.Now()},
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
	_, err := db.DeleteOne(client, "apps", "mail", bson.D{
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
