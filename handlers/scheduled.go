package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"hoagie-profile/auth"
	"hoagie-profile/db"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserScheduledMail struct {
	Status string `json:"status"`
	Mail []ScheduledMail `json:"scheduledMail"`
}

type ScheduledMail struct {
	Header string `json:"header"`
	Sender string `json:"sender"`
	Body   string `json:"body"`
	Email  string `json:"email"`
	UserName string `json:"userName"`
	Schedule time.Time `json:"schedule"`
	CreatedAt time.Time `json:"createdAt"`
}

// Returns true if the date/time schedule is at least a minute
// after the current time, else false
func scheduleValid(schedule string) bool {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("Could not load EST location", err)
		return false
	}
	scheduleTime, err := time.ParseInLocation(time.RFC3339, schedule, est)
	if err != nil {
		fmt.Println("Schedule string is not valid", err)
		return false
	}
	currentTime := time.Now().In(est).Add(time.Minute)
	return scheduleTime.After(currentTime)
}
// Get scheduled mail at a certain time for a given user
// Returns an empty struct if entry doesn't exist
var getScheduled = func(user auth.User, scheduledTime time.Time) (ScheduledMail, error) {
	var response ScheduledMail
	err := db.FindOne(client, "apps", "mail", bson.D{
		{Key: "Email", Value: user.Email},
		{Key: "Schedule", Value: scheduledTime},
	}, &response)
	if err != nil { // Check that error means no docs found?
		return ScheduledMail{}, nil
	}
	return response, nil
}

// Get all scheduled mail for a given user
var getAllScheduled = func(user auth.User) (UserScheduledMail, error) {
	var responses []ScheduledMail

	// Get responses in chronological order
	findOptions := options.Find()
	findOptions.SetSort(bson.D{
		{Key: "Schedule", Value: 1},
	})
	query := bson.D{
		{Key: "Email", Value: user.Email},
	}

	// Perform database search
	resultCursor, err := db.FindMany(client, "apps", "mail", query, findOptions)
	if resultCursor == nil || err != nil {
		return UserScheduledMail{
			Status: "unused",
			Mail: nil, 
		}, fmt.Errorf("error querying scheduled mail: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	// Convert from Cursor to ScheduledMail
	for resultCursor.Next(ctx) {
		var decodedResult ScheduledMail
		err := resultCursor.Decode(&decodedResult)
		if err != nil {
			return UserScheduledMail{
				Status: "unused",
				Mail: nil,
			}, fmt.Errorf("error decoding scheduled mail: %s", err)
		}
		responses = append(responses, decodedResult)
	}
	// If no scheduled mail,, return unused status
	if len(responses) == 0 {
		return UserScheduledMail{
			Status: "unused",
			Mail: nil, 
		}, nil
	}
	defer resultCursor.Close(ctx)
	return UserScheduledMail{
		Status: "used",
		Mail: responses,
	}, nil
}

	// // Scheduled send
	// if mailReq.Schedule != "now" {
	// 	// Validate that schedule is valid
	// 	if !scheduleValid(mailReq.Schedule) {
	// 		deleteVisitor(user.Email)
	// 		http.Error(
	// 			w, 
	// 			"Your email could not be scheduled at the specified time. Please refresh the page and select a later time.",
	// 			http.StatusBadRequest,
	// 		)
	// 		return
	// 	}
	// 	// Convert time to EST and check for errors
	// 	est, err := time.LoadLocation("America/New_York")
	// 	if err != nil {
	// 		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
	// 		deleteVisitor(user.Email)
	// 		return
	// 	}
	// 	scheduleEST, err := time.ParseInLocation(time.RFC3339, mailReq.Schedule, est)
	// 	if err != nil {
	// 		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
	// 		deleteVisitor(user.Email)
	// 		return
	// 	}
	// 	// Add to MongoDB
	// 	db.InsertOne(client, "apps", "mail", bson.D{
	// 		{"Email", mailReq.Email},
	// 		{"Sender", mailReq.Sender},
	// 		{"Header", mailReq.Header},
	// 		{"Body", mailReq.Body},
	// 		{"Schedule", scheduleEST},
	// 		{"UserName", user.Name},
	// 		{"CreatedAt", time.Now()},
	// 	})


// POST /mail/scheduled/user
var scheduledSendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))
	if !success {
		http.Error(w, "You do not have access to schedule mail.", http.StatusBadRequest)
		return
	}
	if len(user.Name) == 0 {
		http.Error(w, `Hoagie has been updated. Please log-out and log-in again.`, http.StatusBadRequest)
	}


})

// GET /mail/scheduled/user
var scheduledUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))
	if !success {
		http.Error(w, "You do not have access to view scheduled mail.", http.StatusBadRequest)
		return
	}
	if len(user.Name) == 0 {
		http.Error(w, `Hoagie has been updated. Please log-out and log-in again.`, http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")

	// Retrieve user's scheduled mail
	userScheduledMail, err := getAllScheduled(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Create JSON response and write
	jsonResp, err := json.Marshal(userScheduledMail)
	if err != nil {
		http.Error(w, "Error in json response marshalling" + err.Error(), http.StatusBadRequest)
	}
	w.Write(jsonResp)
})

// DELETE /mail/scheduled
var scheduledDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

})