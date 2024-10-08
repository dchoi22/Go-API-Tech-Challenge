package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/handlers"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Startup failed, err: %v", err)
	}
}

func run() error {

	serverHost := os.Getenv("HTTP_DOMAIN")
	serverPort := os.Getenv("HTTP_PORT")
	if serverHost == "" || serverPort == "" {
		log.Fatal("SERVER_HOST or SERVER_PORT environment variables not set")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("[in run]: %w", err)
	}

	logger := httplog.NewLogger("user-microservice", httplog.Options{
		LogLevel: slog.LevelDebug,
		JSON:     false,
		Concise:  true,
	})
	defer func() {
		if err = db.Close(); err != nil {
			logger.Error("Error closing db connection", "err", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE"},
		MaxAge:         300,
	}))

	courseSvs := services.NewCourseService(db)
	personSvs := services.NewPersonService(db)
	r.Route("/api", func(r chi.Router) {
		r.Route("/course", func(r chi.Router) {
			r.Get("/", handlers.HandleGetCourses(logger, courseSvs))
			r.Get("/{id}", handlers.HandleGetCourse(logger, courseSvs))
			r.Put("/{id}", handlers.HandleUpdateCourse(logger, courseSvs))
			r.Post("/", handlers.HandleCreateCourse(logger, courseSvs))
			r.Delete("/{id}", handlers.HandleDeleteCourse(logger, courseSvs))
		})
		r.Route("/student", func(r chi.Router) {
			r.Get("/", handlers.HandleGetStudents(logger, personSvs))
			r.Get("/{firstName}", handlers.HandleGetStudent(logger, personSvs))
			r.Put("/{firstName}", handlers.HandleUpdateStudent(logger, personSvs))
			r.Post("/", handlers.HandleCreateStudent(logger, personSvs))
			r.Delete("/{firstName}", handlers.HandleDeleteStudent(logger, personSvs))
		})
		r.Route("/professor", func(r chi.Router) {
			r.Get("/", handlers.HandleGetProfessors(logger, personSvs))
			r.Get("/{firstName}", handlers.HandleGetProfessor(logger, personSvs))
			r.Put("/{firstName}", handlers.HandleUpdateProfessor(logger, personSvs))
			r.Post("/", handlers.HandleCreateProfessor(logger, personSvs))
			r.Delete("/{firstName}", handlers.HandleDeleteProfessor(logger, personSvs))
		})
	})

	serverAddress := fmt.Sprintf("0.0.0.0:%s", serverPort)

	logger.Info(fmt.Sprintf("Attempting to start server on %s", serverAddress))

	serverInstance := &http.Server{
		Addr:              serverAddress,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       500 * time.Millisecond,
		WriteTimeout:      500 * time.Millisecond,
		Handler:           r,
	}

	logger.Info("Server configuration complete, ready to listen")

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		fmt.Println()
		logger.Info("Shutdown signal received")

		// Create a context with a timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel() // Ensure that the context is canceled to release resources

		if err := serverInstance.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down server", "err", err)
			log.Fatalf("Error shutting down server: %v", err)
		}

		logger.Info("Server gracefully shut down")
		serverStopCtx() // Notify the main function to finish
	}()

	logger.Info(fmt.Sprintf("Server is listening on %s", serverInstance.Addr))
	if err := http.ListenAndServe(serverAddress, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Failed to start server", "error", err)
		return fmt.Errorf("server failed: %w", err)
	}

	<-serverCtx.Done()
	logger.Info("Shutdown complete")
	return nil
}
