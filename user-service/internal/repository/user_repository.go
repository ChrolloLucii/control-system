package repository

import (
	"errors"
	"sync"
	"user-service/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	FindAll(page, limit int, role string) ([]*models.User, int, error)
}

type InMemoryUserRepository struct {
	users map[uuid.UUID]*models.User
	mu    sync.RWMutex
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == user.Email {
			return errors.New("user with this email already exists")
		}
	}

	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindAll(page, limit int, role string) ([]*models.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*models.User
	for _, user := range r.users {
		if role != "" {
			hasRole := false
			for _, r := range user.Roles {
				if r == role {
					hasRole = true
					break
				}
			}
			if !hasRole {
				continue
			}
		}
		filtered = append(filtered, user)
	}

	total := len(filtered)
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		return []*models.User{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
