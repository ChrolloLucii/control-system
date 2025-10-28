package service

import (
	"errors"
	"time"
	"user-service/internal/dto"
	"user-service/internal/repository"
	"user-service/models"

	"github.com/google/uuid"
)

type UserService interface {
	Register(req *dto.RegisterRequest) (*models.User, error)
	Login(req *dto.LoginRequest) (*models.User, string, error)
	GetProfile(userID uuid.UUID) (*models.User, error)
	UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*models.User, error)
	GetUsers(page, limit int, role string) ([]*models.User, int, error)
}

type userService struct {
	repo       repository.UserRepository
	jwtService JWTService
}

func NewUserService(repo repository.UserRepository, jwtService JWTService) UserService {
	return &userService{
		repo:       repo,
		jwtService: jwtService,
	}
}

func (s *userService) Register(req *dto.RegisterRequest) (*models.User, error) {
	_, err := s.repo.FindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	user, err := models.NewUser(req.Email, req.Password, req.Name)
	if err != nil {
		return nil, err
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(req *dto.LoginRequest) (*models.User, string, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *userService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(userID)
}

func (s *userService) UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	user.Name = req.Name
	user.UpdatedAt = time.Now()

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUsers(page, limit int, role string) ([]*models.User, int, error) {
	return s.repo.FindAll(page, limit, role)
}
