package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	serverHost := os.Getenv("HTTP_DOMAIN")
	serverPort := os.Getenv("HTTP_PORT")
	if serverHost == "" || serverPort == "" {
		log.Fatal("SERVER_HOST or SERVER_PORT environment variables not set")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	serverAddress := fmt.Sprintf("%s:%s", serverHost, serverPort)

	log.Printf("Starting server on %s...\n", serverAddress)
	if err := http.ListenAndServe(serverAddress, r); err != nil {
		log.Fatal(err)
	}
}
