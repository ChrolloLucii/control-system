package events

import (
	"encoding/json"
	"log"
	"order-service/models"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	OrderCreated       EventType = "ORDER_CREATED"
	OrderStatusUpdated EventType = "ORDER_STATUS_UPDATED"
	OrderCancelled     EventType = "ORDER_CANCELLED"
)

type OrderEvent struct {
	ID        uuid.UUID              `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type EventPublisher interface {
	Publish(event *OrderEvent) error
}

type InMemoryEventPublisher struct{}

func NewInMemoryEventPublisher() *InMemoryEventPublisher {
	return &InMemoryEventPublisher{}
}

func (p *InMemoryEventPublisher) Publish(event *OrderEvent) error {
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("Event Published:\n%s\n", string(eventJSON))
	return nil
}

func NewOrderCreatedEvent(order *models.Order) *OrderEvent {
	return &OrderEvent{
		ID:        uuid.New(),
		Type:      OrderCreated,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"orderId":     order.ID,
			"userId":      order.UserID,
			"totalAmount": order.TotalAmount,
			"itemsCount":  len(order.Items),
			"status":      order.Status,
		},
	}
}

func NewOrderStatusUpdatedEvent(order *models.Order, oldStatus models.OrderStatus) *OrderEvent {
	return &OrderEvent{
		ID:        uuid.New(),
		Type:      OrderStatusUpdated,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"orderId":   order.ID,
			"userId":    order.UserID,
			"oldStatus": oldStatus,
			"newStatus": order.Status,
		},
	}
}

func NewOrderCancelledEvent(order *models.Order) *OrderEvent {
	return &OrderEvent{
		ID:        uuid.New(),
		Type:      OrderCancelled,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"orderId": order.ID,
			"userId":  order.UserID,
		},
	}
}
