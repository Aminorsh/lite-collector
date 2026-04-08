package services

import (
	"fmt"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/utils"
)

// FormService handles form-related operations
type FormService struct {
	formRepo repository.FormRepository
}

// NewFormService creates a new FormService instance
func NewFormService(formRepo repository.FormRepository) *FormService {
	return &FormService{formRepo: formRepo}
}

// CreateForm creates a new draft form owned by the given user.
func (s *FormService) CreateForm(ownerID uint64, title, description string, schema []byte) (*models.Form, error) {
	form := &models.Form{
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		Schema:      schema,
		Status:      0, // draft
	}
	if err := s.formRepo.Create(form); err != nil {
		return nil, utils.ErrFormCreateFail
	}
	return form, nil
}

// GetFormsByOwner returns all forms belonging to the given user.
func (s *FormService) GetFormsByOwner(ownerID uint64) ([]models.Form, error) {
	forms, err := s.formRepo.FindByOwnerID(ownerID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	return forms, nil
}

// GetFormByID returns a form by ID, enforcing ownership.
func (s *FormService) GetFormByID(formID string, userID uint64) (*models.Form, error) {
	id, err := parseID(formID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrFormNotFound
	}

	if form.OwnerID != userID {
		return nil, utils.ErrFormForbidden
	}
	return form, nil
}

// UpdateForm updates title, description, and schema of a form, enforcing ownership.
func (s *FormService) UpdateForm(formID string, userID uint64, title, description string, schema []byte) (*models.Form, error) {
	id, err := parseID(formID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrFormNotFound
	}
	if form.OwnerID != userID {
		return nil, utils.ErrFormForbidden
	}

	form.Title = title
	form.Description = description
	form.Schema = schema

	if err := s.formRepo.Update(form); err != nil {
		return nil, utils.ErrFormUpdateFail
	}
	return form, nil
}

// GetPublishedFormByID returns a form by ID without ownership check.
// Only published forms are accessible this way (submitters need to read the schema).
func (s *FormService) GetPublishedFormByID(formID string) (*models.Form, error) {
	id, err := parseID(formID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}
	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrFormNotFound
	}
	if form.Status != 1 {
		return nil, utils.ErrFormNotPublished
	}
	return form, nil
}

// ArchiveForm changes a form's status to archived, enforcing ownership.
func (s *FormService) ArchiveForm(formID string, userID uint64) error {
	id, err := parseID(formID)
	if err != nil {
		return utils.ErrBadRequest
	}
	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return utils.ErrFormNotFound
	}
	if form.OwnerID != userID {
		return utils.ErrFormForbidden
	}
	if err := s.formRepo.Archive(id); err != nil {
		return utils.ErrFormArchiveFail
	}
	return nil
}

// PublishForm changes a form's status to published, enforcing ownership.
func (s *FormService) PublishForm(formID string, userID uint64) error {
	id, err := parseID(formID)
	if err != nil {
		return utils.ErrBadRequest
	}

	form, err := s.formRepo.FindByID(id)
	if err != nil {
		return utils.ErrFormNotFound
	}
	if form.OwnerID != userID {
		return utils.ErrFormForbidden
	}

	if err := s.formRepo.Publish(id); err != nil {
		return utils.ErrFormPublishFail
	}
	return nil
}

// parseID converts a string path parameter to uint64.
func parseID(s string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
