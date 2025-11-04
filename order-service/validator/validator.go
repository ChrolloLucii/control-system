package validator

import (
	"errors"
	"order-service/internal/dto"
)

func ValidateCreateOrderRequest(req *dto.CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	for i, item := range req.Items {
		if item.ProductName == "" {
			return errors.New("product name is required for all items")
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}
		if item.Price < 0 {
			return errors.New("price cannot be negative")
		}
		if item.Price == 0 {
			return errors.New("price must be greater than 0")
		}
		_ = i
	}

	return nil
}

func ValidateOrderStatus(status string) error {
	validStatuses := map[string]bool{
		"created":     true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}

	if !validStatuses[status] {
		return errors.New("invalid order status")
	}

	return nil
}
