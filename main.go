package main

import (
	"fmt"
	"hoagie-profile/auth"
	"hoagie-profile/handlers"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Response struct {
	Message string `json:"message"`
}

const (
	mailApiRoute     = "/mail"
	mailSendApiRoute = "/mail/send"
)

func main() {
	godotenv.Load(".env.local")
	domain := os.Getenv("AUTH0_DOMAIN")
	audience := os.Getenv("AUTH0_AUDIENCE")
	hostname := os.Getenv("HOAGIE_HOST")
	port := os.Getenv("PORT")
	runtimeMode := os.Getenv("HOAGIE_MODE")

	jwtMiddleware := auth.Middleware(domain, audience)

	r := mux.NewRouter()
	r.Host(hostname).Subrouter()

	handlers.Setup(r, jwtMiddleware)

	corsWrapper := auth.CorsWrapper(runtimeMode)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsWrapper.Handler(r),
	}

	fmt.Println(`
	█░█ █▀█ ▄▀█ █▀▀ █ █▀▀
	█▀█ █▄█ █▀█ █▄█ █ ██▄
	`)
	fmt.Printf("[i] Running on https://%s:%s\n", hostname, port)
	if runtimeMode == "debug" {
		fmt.Println("[i] Debug mode is on.")
	}
	server.ListenAndServe()
}
