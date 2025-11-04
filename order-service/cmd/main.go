package main

import (
	"log"
	"net/http"
	"order-service/internal/events"
	"order-service/internal/handlers"
	"order-service/internal/middleware"
	"order-service/internal/repository"
	"order-service/internal/service"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production-12345"
	}

	// Инициализация зависимостей
	orderRepo := repository.NewInMemoryOrderRepository()
	eventPublisher := events.NewInMemoryEventPublisher()
	userClient := service.NewHTTPUserClient()
	orderService := service.NewOrderService(orderRepo, eventPublisher, userClient)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Настройка роутера
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.CORSMiddleware)

	// Регистрация роутов
	orderHandler.RegisterRoutes(r, jwtSecret)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"order-service"}`))
	})

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	log.Printf("Order Service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
