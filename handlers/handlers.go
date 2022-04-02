package handlers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func Setup(r *mux.Router, m *jwtmiddleware.JWTMiddleware, cl *mongo.Client) {
	// Handle mail send request
	client = cl
	r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("POST")
	r.Handle(stuffUserRoute, m.Handler(digestSendHandler)).Methods("POST")
	r.Handle(stuffUserRoute, m.Handler(digestStatusHandler)).Methods("GET")
	r.Handle(stuffUserRoute, m.Handler(digestDeleteHandler)).Methods("DELETE")
}
