package handlers

import (
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
)

const (
	mailRoute     = "/mail"
	mailSendRoute = "/mail/send"
)

func setupMailHandlers(r *mux.Router, m *jwtmiddleware.JWTMiddleware) {
	// Handle mail send request
	r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("GET")
}

var sendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))
})
