package services

import (
	"fmt"

	"lite-collector/models"
	"lite-collector/repository"
)

// FormService handles form-related operations
type FormService struct {
	formRepo repository.FormRepository
}

// NewFormService creates a new FormService instance with dependency injection
func NewFormService(formRepo repository.FormRepository) *FormService {
	return &FormService{
		formRepo: formRepo,
	}
}

// CreateForm creates a new form
func (s *FormService) CreateForm(ownerID uint64, title, description string, schema []byte) (*models.Form, error) {
	form := &models.Form{
		OwnerID:   ownerID,
		Title:     title,
		Description: description,
		Schema:    schema,
		Status:    0, // draft
	}
	err := s.formRepo.Create(form)
	return form, err
}

// GetFormsByOwner gets all forms for a specific owner
func (s *FormService) GetFormsByOwner(ownerID uint64) ([]models.Form, error) {
	return s.formRepo.FindByOwnerID(ownerID)
}

// GetFormByID gets a form by ID, ensuring the user owns it
func (s *FormService) GetFormByID(formID string, userID uint64) (*models.Form, error) {
	// Convert string ID to uint64
	var id uint64
	_, err := fmt.Sscanf(formID, "%d", &id)
	if err != nil {
		return nil, err
	}

	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if form.OwnerID != userID {
		return nil, fmt.Errorf("unauthorized access to form")
	}

	return form, nil
}

// UpdateForm updates an existing form
func (s *FormService) UpdateForm(formID string, userID uint64, title, description string, schema []byte) (*models.Form, error) {
	// Convert string ID to uint64
	var id uint64
	_, err := fmt.Sscanf(formID, "%d", &id)
	if err != nil {
		return nil, err
	}

	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if form.OwnerID != userID {
		return nil, fmt.Errorf("unauthorized access to form")
	}

	// Update form fields
	form.Title = title
	form.Description = description
	form.Schema = schema

	err = s.formRepo.Update(form)
	return form, err
}

// PublishForm publishes a form (changes status to published)
func (s *FormService) PublishForm(formID string, userID uint64) error {
	// Convert string ID to uint64
	var id uint64
	_, err := fmt.Sscanf(formID, "%d", &id)
	if err != nil {
		return err
	}

	// Check ownership first
	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return err
	}

	if form.OwnerID != userID {
		return fmt.Errorf("unauthorized access to form")
	}

	// Publish the form
	return s.formRepo.Publish(id)
}