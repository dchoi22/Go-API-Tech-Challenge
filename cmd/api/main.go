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

	r.Route("/api", func(r chi.Router) {
		r.Route("/course", func(r chi.Router) {
			r.Get("/", GetCourses)
			r.Get("/{id}", GetCourse)
			r.Put("/{id}", UpdateCourse)
			r.Post("/", CreateCourse)
			r.Delete("/{id}", DeleteCourse)
		})
		r.Route("/person", func(r chi.Router) {
			r.Get("/", GetPeople)
			r.Get("/{name}", GetPerson)
			r.Put("/{name}", UpdatePerson)
			r.Post("/", CreatePerson)
			r.Delete("/{name}", DeletePerson)
		})
	})

	serverAddress := fmt.Sprintf("%s:%s", serverHost, serverPort)

	log.Printf("Starting server on %s...\n", serverAddress)
	if err := http.ListenAndServe(serverAddress, r); err != nil {
		log.Fatal(err)
	}
}
