package service

import (
	"errors"
	"order-service/internal/dto"
	"order-service/internal/events"
	"order-service/internal/repository"
	"order-service/models"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(userID uuid.UUID, req *dto.CreateOrderRequest, token string) (*models.Order, error)
	GetOrder(orderID, userID uuid.UUID, isAdmin bool) (*models.Order, error)
	GetUserOrders(userID uuid.UUID, page, limit int, sortBy string) ([]*models.Order, int, error)
	UpdateOrderStatus(orderID, userID uuid.UUID, status string, isAdmin bool) (*models.Order, error)
	CancelOrder(orderID, userID uuid.UUID, isAdmin bool) (*models.Order, error)
}

type orderService struct {
	repo           repository.OrderRepository
	eventPublisher events.EventPublisher
	userClient     UserClient
}

func NewOrderService(repo repository.OrderRepository, eventPublisher events.EventPublisher, userClient UserClient) OrderService {
	return &orderService{
		repo:           repo,
		eventPublisher: eventPublisher,
		userClient:     userClient,
	}
}

func (s *orderService) CreateOrder(userID uuid.UUID, req *dto.CreateOrderRequest, token string) (*models.Order, error) {
	exists, err := s.userClient.UserExists(userID, token)
	if err != nil || !exists {
		return nil, errors.New("user not found or invalid")
	}

	var items []models.OrderItem
	for _, item := range req.Items {
		items = append(items, models.OrderItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	order := models.NewOrder(userID, items)

	err = s.repo.Create(order)
	if err != nil {
		return nil, err
	}

	event := events.NewOrderCreatedEvent(order)
	s.eventPublisher.Publish(event)

	return order, nil
}

func (s *orderService) GetOrder(orderID, userID uuid.UUID, isAdmin bool) (*models.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	if !isAdmin && order.UserID != userID {
		return nil, errors.New("access denied")
	}

	return order, nil
}

func (s *orderService) GetUserOrders(userID uuid.UUID, page, limit int, sortBy string) ([]*models.Order, int, error) {
	return s.repo.FindByUserID(userID, page, limit, sortBy)
}

func (s *orderService) UpdateOrderStatus(orderID, userID uuid.UUID, status string, isAdmin bool) (*models.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	if !isAdmin && order.UserID != userID {
		return nil, errors.New("access denied")
	}

	oldStatus := order.Status
	order.UpdateStatus(models.OrderStatus(status))

	err = s.repo.Update(order)
	if err != nil {
		return nil, err
	}

	event := events.NewOrderStatusUpdatedEvent(order, oldStatus)
	s.eventPublisher.Publish(event)

	return order, nil
}

func (s *orderService) CancelOrder(orderID, userID uuid.UUID, isAdmin bool) (*models.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	if !isAdmin && order.UserID != userID {
		return nil, errors.New("access denied")
	}

	if order.Status == models.StatusCompleted {
		return nil, errors.New("cannot cancel completed order")
	}

	order.Cancel()

	err = s.repo.Update(order)
	if err != nil {
		return nil, err
	}

	event := events.NewOrderCancelledEvent(order)
	s.eventPublisher.Publish(event)

	return order, nil
}
