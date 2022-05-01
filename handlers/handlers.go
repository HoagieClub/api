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

func Setup(r *mux.Router, m *jwtmiddleware.JWTMiddleware, cl *mongo.Client) {
	// Handle mail send request
	client = cl

	setupStuffIndex()

	r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("POST")
	r.Handle(stuffUserRoute, m.Handler(stuffSendHandler)).Methods("POST")
	r.Handle(stuffUserRoute, m.Handler(stuffUserHandler)).Methods("GET")
	r.Handle(stuffUserRoute, m.Handler(digestDeleteHandler)).Methods("DELETE")
	r.Handle(stuffRoute, m.Handler(stuffAllHandler)).Methods("GET")

	// princeton_token, err := _refreshToken()
	// if err != nil {
	// 	fmt.Println("Error getting Princeton API token, not going to setup campus handlers.")
	// 	return
	// }
	// print(princeton_token)
}
