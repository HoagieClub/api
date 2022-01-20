package handlers

import (
	"encoding/json"
	"hoagie-profile/auth"
	"hoagie-profile/db"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson"
)

type MailDigestRequest struct {
	Title       string
	Category    string
	Description string
}

var digestSendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	accessToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	user, err := auth.GetUser(accessToken)
	if os.Getenv("HOAGIE_MODE") == "debug" {
		user = "test@princeton.edu"
	} else if err != nil {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	var digestReq MailDigestRequest
	err = json.NewDecoder(r.Body).Decode(&digestReq)
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
