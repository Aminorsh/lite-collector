package repository

import (
	"lite-collector/models"

	"gorm.io/gorm"
)

// FormRepository defines the interface for form data access
type FormRepository interface {
	Create(form *models.Form) error
	FindByID(id uint64) (*models.Form, error)
	FindByOwnerID(ownerID uint64) ([]models.Form, error)
	Update(form *models.Form) error
	Publish(formID uint64) error
}

// formRepository implements FormRepository using GORM
type formRepository struct {
	db *gorm.DB
}

// NewFormRepository creates a new form repository instance
func NewFormRepository(db *gorm.DB) FormRepository {
	return &formRepository{db: db}
}

// Create creates a new form
func (r *formRepository) Create(form *models.Form) error {
	return r.db.Create(form).Error
}

// FindByID finds a form by ID
func (r *formRepository) FindByID(id uint64) (*models.Form, error) {
	var form models.Form
	result := r.db.First(&form, id)
	return &form, result.Error
}

// FindByOwnerID finds all forms by owner ID
func (r *formRepository) FindByOwnerID(ownerID uint64) ([]models.Form, error) {
	var forms []models.Form
	result := r.db.Where("owner_id = ?", ownerID).Find(&forms)
	return forms, result.Error
}

// Update updates an existing form
func (r *formRepository) Update(form *models.Form) error {
	return r.db.Save(form).Error
}

// Publish publishes a form (changes status to published)
func (r *formRepository) Publish(formID uint64) error {
	return r.db.Model(&models.Form{}).Where("id = ?", formID).Update("status", 1).Error
}