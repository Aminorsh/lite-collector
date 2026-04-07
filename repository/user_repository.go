package repository

import (
	"lite-collector/db"
	"lite-collector/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByOpenID(openID string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	FindByID(id uint64) (*models.User, error)
}

// userRepository implements UserRepository using GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// FindByOpenID finds a user by their WeChat OpenID
func (r *userRepository) FindByOpenID(openID string) (*models.User, error) {
	var user models.User
	result := r.db.Where("open_id = ?", openID).First(&user)
	return &user, result.Error
}

// Create creates a new user
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(id uint64) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	return &user, result.Error
}