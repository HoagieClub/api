package handlers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
)

func Setup(router *mux.Router, middleware *jwtmiddleware.JWTMiddleware) {
	setupMailHandlers(router, middleware)
}
