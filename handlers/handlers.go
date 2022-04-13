package handlers

import (
	"fmt"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

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
	r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("POST")
	r.Handle(stuffUserRoute, m.Handler(digestSendHandler)).Methods("POST")
	r.Handle(stuffUserRoute, m.Handler(digestStatusHandler)).Methods("GET")
	r.Handle(stuffRoute, m.Handler(digestStatusHandler)).Methods("GET")
	r.Handle(stuffUserRoute, m.Handler(digestDeleteHandler)).Methods("DELETE")

	princeton_token, err := _refreshToken()
	if err != nil {
		fmt.Println("Error getting Princeton API token, not going to setup campus handlers.")
		return
	}
	print(princeton_token)
}
