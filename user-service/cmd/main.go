package main

import (
	"log"
	"net/http"
	"os"
	"user-service/internal/handlers"
	"user-service/internal/middleware"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system variables")

	}
	userRepo := repository.NewInMemoryUserRepository()
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, jwtService)
	userHandler := handlers.NewUserHandler(userService)

	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.CORSMiddleware)

	userHandler.RegisterRoutes(r, jwtService)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"user-service"}`))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("User Service starting up... on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Error starting User Service: %v", err)
	}
}
