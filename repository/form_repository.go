package repository

import (
	"lite-collector/models"

	"gorm.io/gorm"
)

// FormListFilter narrows a form listing by text, status, and sort order.
// Status == nil means "any status". SortBy is one of "updated_at",
// "created_at", or "title"; Order is "asc" or "desc".
type FormListFilter struct {
	Query  string
	Status *int8
	SortBy string
	Order  string
}

// FormRepository defines the interface for form data access
type FormRepository interface {
	Create(form *models.Form) error
	FindByID(id uint64) (*models.Form, error)
	FindByOwnerID(ownerID uint64) ([]models.Form, error)
	FindByOwnerWithFilter(ownerID uint64, filter FormListFilter) ([]models.Form, error)
	Update(form *models.Form) error
	Publish(formID uint64) error
	Archive(formID uint64) error
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

// allowedSortColumns guards against untrusted sort input hitting the DB.
var allowedSortColumns = map[string]string{
	"updated_at": "updated_at",
	"created_at": "created_at",
	"title":      "title",
}

// FindByOwnerWithFilter applies query (title LIKE), status, sort, and order.
func (r *formRepository) FindByOwnerWithFilter(ownerID uint64, filter FormListFilter) ([]models.Form, error) {
	q := r.db.Where("owner_id = ?", ownerID)

	if filter.Query != "" {
		q = q.Where("title LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}

	col, ok := allowedSortColumns[filter.SortBy]
	if !ok {
		col = "updated_at"
	}
	order := "desc"
	if filter.Order == "asc" {
		order = "asc"
	}

	var forms []models.Form
	result := q.Order(col + " " + order).Find(&forms)
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

// Archive archives a form (changes status to archived)
func (r *formRepository) Archive(formID uint64) error {
	return r.db.Model(&models.Form{}).Where("id = ?", formID).Update("status", 2).Error
}