package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"user-service/internal/dto"
	"user-service/internal/middleware"
	"user-service/internal/service"
	"user-service/validator"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if err := validator.ValidateRegisterRequest(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusCreated, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if err := validator.ValidateLoginRequest(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user, token, err := h.userService.Login(&req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	response := dto.LoginResponse{
		Token: token,
		User:  user,
	}

	respondWithSuccess(w, http.StatusOK, response)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*service.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	user, err := h.userService.GetProfile(claims.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*service.Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}

	user, err := h.userService.UpdateProfile(claims.UserID, &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	respondWithSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	role := r.URL.Query().Get("role")

	users, total, err := h.userService.GetUsers(page, limit, role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := dto.PaginatedResponse{
		Success: true,
		Data:    users,
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

func (h *UserHandler) RegisterRoutes(r chi.Router, jwtService service.JWTService) {
	r.Route("/api/v1/users", func(r chi.Router) {
		// публично
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		//защищено
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(jwtService))
			r.Get("/profile", h.GetProfile)
			r.Put("/profile", h.UpdateProfile)

			// Админка
			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminMiddleware)
				r.Get("/", h.GetUsers)
			})
		})
	})
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
