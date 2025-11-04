package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"order-service/internal/dto"
	"order-service/internal/middleware"
	"order-service/internal/service"
	"order-service/validator"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if err := validator.ValidateCreateOrderRequest(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:] // Убираем "Bearer "
	}

	order, err := h.orderService.CreateOrder(claims.UserID, &req, token)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "invalid order ID")
		return
	}

	isAdmin := hasRole(claims.Roles, "admin")

	order, err := h.orderService.GetOrder(orderID, claims.UserID, isAdmin)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "ORDER_NOT_FOUND", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusOK, order)
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "createdAt_desc"
	}

	orders, total, err := h.orderService.GetUserOrders(claims.UserID, page, limit, sortBy)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := dto.PaginatedResponse{
		Success: true,
		Data:    orders,
		Meta: dto.MetaDTO{
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
			TotalItems: total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "invalid order ID")
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if err := validator.ValidateOrderStatus(req.Status); err != nil {
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	isAdmin := hasRole(claims.Roles, "admin")

	order, err := h.orderService.UpdateOrderStatus(orderID, claims.UserID, req.Status, isAdmin)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "invalid order ID")
		return
	}

	isAdmin := hasRole(claims.Roles, "admin")

	order, err := h.orderService.CancelOrder(orderID, claims.UserID, isAdmin)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusOK, order)
}

func (h *OrderHandler) RegisterRoutes(r chi.Router, jwtSecret string) {
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtSecret))

		r.Post("/", h.CreateOrder)
		r.Get("/", h.GetUserOrders)
		r.Get("/{id}", h.GetOrder)
		r.Put("/{id}/status", h.UpdateOrderStatus)
		r.Delete("/{id}", h.CancelOrder)
	})
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error: &dto.ErrorDTO{
			Code:    code,
			Message: message,
		},
	})
}

func respondWithSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}
