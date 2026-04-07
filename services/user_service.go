package services

import (
	"lite-collector/models"
	"lite-collector/repository"
)

// UserService handles user-related operations
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance with dependency injection
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// FindByOpenID finds a user by their WeChat OpenID
func (s *UserService) FindByOpenID(openID string) (*models.User, error) {
	return s.userRepo.FindByOpenID(openID)
}

// Create creates a new user
func (s *UserService) Create(user *models.User) error {
	return s.userRepo.Create(user)
}

// Update updates an existing user
func (s *UserService) Update(user *models.User) error {
	return s.userRepo.Update(user)
}

// FindByID finds a user by ID
func (s *UserService) FindByID(id uint64) (*models.User, error) {
	return s.userRepo.FindByID(id)
}