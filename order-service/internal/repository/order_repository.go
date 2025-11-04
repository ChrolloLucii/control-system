package repository

import (
	"errors"
	"order-service/models"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(order *models.Order) error
	FindByID(id uuid.UUID) (*models.Order, error)
	FindByUserID(userID uuid.UUID, page, limit int, sortBy string) ([]*models.Order, int, error)
	Update(order *models.Order) error
	Delete(id uuid.UUID) error
}

type InMemoryOrderRepository struct {
	orders map[uuid.UUID]*models.Order
	mu     sync.RWMutex
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[uuid.UUID]*models.Order),
	}
}

func (r *InMemoryOrderRepository) Create(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (r *InMemoryOrderRepository) FindByUserID(userID uuid.UUID, page, limit int, sortBy string) ([]*models.Order, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userOrders []*models.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}

	if sortBy == "createdAt_desc" {
		sort.Slice(userOrders, func(i, j int) bool {
			return userOrders[i].CreatedAt.After(userOrders[j].CreatedAt)
		})
	} else if sortBy == "createdAt_asc" {
		sort.Slice(userOrders, func(i, j int) bool {
			return userOrders[i].CreatedAt.Before(userOrders[j].CreatedAt)
		})
	} else if sortBy == "totalAmount_desc" {
		sort.Slice(userOrders, func(i, j int) bool {
			return userOrders[i].TotalAmount > userOrders[j].TotalAmount
		})
	} else if sortBy == "totalAmount_asc" {
		sort.Slice(userOrders, func(i, j int) bool {
			return userOrders[i].TotalAmount < userOrders[j].TotalAmount
		})
	}

	total := len(userOrders)
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		return []*models.Order{}, total, nil
	}
	if end > total {
		end = total
	}

	return userOrders[start:end], total, nil
}

func (r *InMemoryOrderRepository) Update(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return errors.New("order not found")
	}

	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[id]; !exists {
		return errors.New("order not found")
	}

	delete(r.orders, id)
	return nil
}
