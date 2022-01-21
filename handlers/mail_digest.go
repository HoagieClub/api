package handlers

import (
	"encoding/json"
	"fmt"
	"hoagie-profile/db"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson"
)

type DigestRequest struct {
	Title       string
	Category    string
	Description string
	Email       string
}

type DigestResponse struct {
	Title       string
	Category    string
	Description string
	Email       string
	Status      string
}

var getCurrentDigest = func(user string) (DigestResponse, error) {
	var response DigestResponse
	err := db.FindOne(client, "apps", "mail", bson.D{{"email", user}}, &response)
	if err != nil {
		return DigestResponse{}, fmt.Errorf("Error getting digest: %s", err)
	}
	return response, nil
}

// GET /mail/digest
var digestStatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))

	if !success {
		http.Error(w, "You do not have access to send digest.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	currentDigest, err := getCurrentDigest(user)
	if err != nil || currentDigest.Title == "" {
		result, _ := json.Marshal(map[string]string{"status": "unused"})
		w.Write(result)
		return
	}
	jsonResp, err := json.Marshal(currentDigest)
	currentDigest.Status = "used"
	if err != nil {
		log.Fatalf("Error happened in response marshalling. %s", err)
	}
	w.Write(jsonResp)
})

// POST /mail/digest
var digestSendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))

	if !success {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	var digestReq DigestRequest
	err := json.NewDecoder(r.Body).Decode(&digestReq)
	if err != nil {
		http.Error(w, "Message did not contain correct fields.", http.StatusBadRequest)
		deleteVisitor(user)
		return
	}
	// Validation here
	// Category
	if digestReq.Category != "Lost and found" && digestReq.Category != "Sale" {
		http.Error(w, "Category must be lost and found or sale", http.StatusBadRequest)
		return
	}

	// Title length
	if utf8.RuneCountInString(digestReq.Title) < 5 || utf8.RuneCountInString(digestReq.Title) > 100 {
		http.Error(w, "Title needs to be between 5 and 200 characters inclusive.", http.StatusBadRequest)
		return
	}

	// Description Length
	if utf8.RuneCountInString(digestReq.Description) < 5 || utf8.RuneCountInString(digestReq.Description) > 200 {
		http.Error(w, "Description needs to be between 5 and 200 characters inclusive.", http.StatusBadRequest)
	}
	current, err := getCurrentDigest(user)
	if err != nil {
		http.Error(w, "The Digest service is temporarily unavaialble.", http.StatusInternalServerError)
	}
	if current.Title != "" {
		http.Error(w, "You have already an existing digest. Try deleting it and send again.", http.StatusBadRequest)
		deleteVisitor(user)
		return
	}

	// Add the digest request to the user's digest queue
	db.InsertOne(client, "apps", "mail", bson.D{
		{"email", user},
		{"title", digestReq.Title},
		{"description", digestReq.Description},
		{"category", digestReq.Category},
		{"created_at", time.Now()},
	})
})
