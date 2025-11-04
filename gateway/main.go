package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"gateway/middleware"
	"gateway/proxy"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Конфиг
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production-12345")
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:3001")
	orderServiceURL := getEnv("ORDER_SERVICE_URL", "http://localhost:3002")
	rateLimitRPS := getEnvInt("RATE_LIMIT_RPS", 100)
	rateLimitBurst := getEnvInt("RATE_LIMIT_BURST", 200)

	reverseProxy := proxy.NewReverseProxy(userServiceURL, orderServiceURL)
	rateLimiter := middleware.NewRateLimiter(rateLimitRPS, rateLimitBurst)

	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(rateLimiter.Middleware)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	})

	r.Route("/api/v1/users", func(r chi.Router) {
		// Публичные маршруты
		r.Group(func(r chi.Router) {
			r.Post("/register", reverseProxy.ProxyToUserService)
			r.Post("/login", reverseProxy.ProxyToUserService)
		})

		// Защищённые маршруты
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuthMiddleware(jwtSecret))
			r.Get("/profile", reverseProxy.ProxyToUserService)
			r.Put("/profile", reverseProxy.ProxyToUserService)
			r.Get("/", reverseProxy.ProxyToUserService) // Список пользователей
		})
	})

	// Order Service routes
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(jwtSecret))
		r.Post("/", reverseProxy.ProxyToOrderService)
		r.Get("/", reverseProxy.ProxyToOrderService)
		r.Get("/{id}", reverseProxy.ProxyToOrderService)
		r.Put("/{id}/status", reverseProxy.ProxyToOrderService)
		r.Delete("/{id}", reverseProxy.ProxyToOrderService)
	})

	log.Printf("Gateway starting on port %s", port)
	log.Printf("User Service: %s", userServiceURL)
	log.Printf("Order Service: %s", orderServiceURL)
	log.Printf("Rate Limit: %d RPS, Burst: %d", rateLimitRPS, rateLimitBurst)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
