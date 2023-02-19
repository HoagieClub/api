package handlers

import (
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var REQUEST_TIMEOUT = 10 * time.Second

// Delete Stuff information every week
var EXPIRATION_DURATION = 60 * 60 * 24 * 10
var client *mongo.Client

const (
	mailRoute      = "/mail"
	mailSendRoute  = "/mail/send"
	stuffRoute     = "/stuff"
	stuffUserRoute = "/stuff/user"
)

func Setup(r *mux.Router, cl *mongo.Client, m *jwtmiddleware.JWTMiddleware) {
	client = cl
	setupStuffIndex()

	if m == nil {
		r.Handle(mailSendRoute, sendHandler).Methods("POST")
		r.Handle(stuffUserRoute, stuffSendHandler).Methods("POST")
		r.Handle(stuffUserRoute, stuffUserHandler).Methods("GET")
		r.Handle(stuffUserRoute, stuffDeleteHandler).Methods("DELETE")
		r.Handle(stuffRoute, stuffAllHandler).Methods("GET")
		return
	} else {
		r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("POST")
		r.Handle(stuffUserRoute, m.Handler(stuffSendHandler)).Methods("POST")
		r.Handle(stuffUserRoute, m.Handler(stuffUserHandler)).Methods("GET")
		r.Handle(stuffUserRoute, m.Handler(stuffDeleteHandler)).Methods("DELETE")
		r.Handle(stuffRoute, m.Handler(stuffAllHandler)).Methods("GET")
	}

	// princeton_token, err := _refreshToken()
	// if err != nil {
	// 	fmt.Println("Error getting Princeton API token, not going to setup campus handlers.")
	// 	return
	// }
	// print(princeton_token)
}
