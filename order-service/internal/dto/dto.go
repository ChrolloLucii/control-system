package dto

import "github.com/google/uuid"

type OrderItemRequest struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    MetaDTO     `json:"meta"`
}

type MetaDTO struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
	TotalItems int `json:"totalItems"`
}

type UserExistsRequest struct {
	UserID uuid.UUID `json:"userId"`
}
