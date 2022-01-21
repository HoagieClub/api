package main

import (
	"fmt"
	"hoagie-profile/auth"
	"hoagie-profile/db"
	"hoagie-profile/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	godotenv.Load(".env.local")
	domain := os.Getenv("AUTH0_DOMAIN")
	audience := os.Getenv("AUTH0_AUDIENCE")
	hostname := os.Getenv("HOAGIE_HOST")
	port := os.Getenv("PORT")
	runtimeMode := os.Getenv("HOAGIE_MODE")

	r := mux.NewRouter()
	r.Host(hostname).Subrouter()

	jwtMiddleware := auth.Middleware(domain, audience)

	client, err := db.MongoClient()
	if err != nil {
		log.Fatal("Database connection error")
	}

	handlers.Setup(r, jwtMiddleware, client)

	corsWrapper := auth.CorsWrapper(runtimeMode)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsWrapper.Handler(r),
	}

	fmt.Println(`
	█░█ █▀█ ▄▀█ █▀▀ █ █▀▀
	█▀█ █▄█ █▀█ █▄█ █ ██▄
	`)
	if runtimeMode == "debug" {
		fmt.Println("[i] Debug mode is on.")
		var user map[string]interface{}
		err := db.FindUser(client, "test@princeton.edu", &user)
		if err != nil {
			fmt.Println("[i] Test user not found in debug mode...")
			panic("Please configure the database.")
		}
	}
	fmt.Printf("[i] Running on https://%s:%s\n", hostname, port)
	server.ListenAndServe()
}
