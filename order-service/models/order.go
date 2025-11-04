package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusCreated    OrderStatus = "created"
	StatusInProgress OrderStatus = "in_progress"
	StatusCompleted  OrderStatus = "completed"
	StatusCancelled  OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Order struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"userId"`
	Items       []OrderItem `json:"items"`
	Status      OrderStatus `json:"status"`
	TotalAmount float64     `json:"totalAmount"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func NewOrder(userID uuid.UUID, items []OrderItem) *Order {
	totalAmount := 0.0
	for _, item := range items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	return &Order{
		ID:          uuid.New(),
		UserID:      userID,
		Items:       items,
		Status:      StatusCreated,
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

func (o *Order) Cancel() {
	o.Status = StatusCancelled
	o.UpdatedAt = time.Now()
}
