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
	"go.mongodb.org/mongo-driver/mongo"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	godotenv.Load(".env.local")
	runtimeMode := os.Getenv("HOAGIE_MODE")

	var server *http.Server
	var client *mongo.Client
	var hostname string = "localhost"
	var port string = "8080"

	if runtimeMode == "debug" {
		// Local development
		var err error

		r := mux.NewRouter()
		r.Host("localhost").Subrouter()
		os.Setenv("MONGO_URI", "mongodb://localhost:27017")
		client, err = db.MongoClient()
		if err != nil {
			log.Fatal("Database connection error" + err.Error())
		}

		handlers.Setup(r, client, nil)

		server = &http.Server{
			Addr:    ":" + port,
			Handler: r,
		}
	} else {
		domain := os.Getenv("AUTH0_DOMAIN")
		audience := os.Getenv("AUTH0_AUDIENCE")
		hostname = os.Getenv("HOAGIE_HOST")
		port = os.Getenv("PORT")

		r := mux.NewRouter()
		r.Host(hostname).Subrouter()

		jwtMiddleware := auth.Middleware(domain, audience)

		client, err := db.MongoClient()
		if err != nil {
			log.Fatal("Database connection error" + err.Error())
		}
		handlers.Setup(r, client, jwtMiddleware)

		corsWrapper := auth.CorsWrapper(runtimeMode)

		server = &http.Server{
			Addr:    ":" + port,
			Handler: corsWrapper.Handler(r),
		}
	}

	fmt.Println(`
	█░█ █▀█ ▄▀█ █▀▀ █ █▀▀
	█▀█ █▄█ █▀█ █▄█ █ ██▄
	`)
	if runtimeMode == "debug" {
		fmt.Println("[i] Debug mode is on.")
		if os.Args[1] == "reset" {
			err := db.SetupInitialDatabase(client)
			if err != nil {
				panic("Failed to create initial database. Make sure you have a clean MongoDB instance running.")
			}
		}
	}
	fmt.Printf("[i] Running on https://%s:%s\n", hostname, port)
	server.ListenAndServe()
}
