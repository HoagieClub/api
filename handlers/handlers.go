package handlers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func Setup(router *mux.Router, middleware *jwtmiddleware.JWTMiddleware, client *mongo.Client) {
	setupMailHandlers(router, middleware, client)
}
