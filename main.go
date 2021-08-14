package main

import (
	"fmt"
	"hoagie-profile/auth"
	"hoagie-profile/handlers"
	"net/http"
	"os"

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
	runtimeMode := os.Getenv("HOAGIE_MODE")

	jwtMiddleware := auth.Middleware(domain, audience)

	r := auth.NewRouter()

	handlers.Setup(r, jwtMiddleware)

	corsWrapper := auth.CorsWrapper(runtimeMode)

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsWrapper.Handler(r),
	}

	fmt.Println(`
	█░█ █▀█ ▄▀█ █▀▀ █ █▀▀
	█▀█ █▄█ █▀█ █▄█ █ ██▄
	`)
	fmt.Println("[i] Running on https://localhost:8080")

	server.ListenAndServe()
}
