package services

import (
	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/utils"
)

// BaseDataService handles base data operations
type BaseDataService struct {
	baseDataRepo repository.BaseDataRepository
}

// NewBaseDataService creates a new BaseDataService instance
func NewBaseDataService(baseDataRepo repository.BaseDataRepository) *BaseDataService {
	return &BaseDataService{baseDataRepo: baseDataRepo}
}

// ImportRow creates or updates a single base data row for a form.
// If a row with the same form_id + row_key exists, it updates the data.
func (s *BaseDataService) ImportRow(formID uint64, rowKey string, data []byte) (*models.BaseData, error) {
	existing, err := s.baseDataRepo.FindByFormIDAndRowKey(formID, rowKey)
	if err == nil {
		// Update existing row
		existing.Data = data
		if err := s.baseDataRepo.Update(existing); err != nil {
			return nil, utils.ErrInternal
		}
		return existing, nil
	}

	// Create new row
	bd := &models.BaseData{
		FormID: formID,
		RowKey: rowKey,
		Data:   data,
	}
	if err := s.baseDataRepo.Create(bd); err != nil {
		return nil, utils.ErrInternal
	}
	return bd, nil
}

// BatchImport creates or updates multiple base data rows for a form.
func (s *BaseDataService) BatchImport(formID uint64, rows []BaseDataRow) (int, error) {
	count := 0
	for _, row := range rows {
		if _, err := s.ImportRow(formID, row.RowKey, row.Data); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// BaseDataRow is the input format for batch import
type BaseDataRow struct {
	RowKey string `json:"row_key"`
	Data   []byte `json:"data"`
}

// GetByFormID returns all base data rows for a form.
func (s *BaseDataService) GetByFormID(formID uint64) ([]models.BaseData, error) {
	list, err := s.baseDataRepo.FindByFormID(formID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	return list, nil
}

// Lookup returns a single base data row by form_id + row_key.
// This is used by submitters to prefill their form.
func (s *BaseDataService) Lookup(formID uint64, rowKey string) (*models.BaseData, error) {
	bd, err := s.baseDataRepo.FindByFormIDAndRowKey(formID, rowKey)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return bd, nil
}

// DeleteByFormID removes all base data for a form.
func (s *BaseDataService) DeleteByFormID(formID uint64) error {
	if err := s.baseDataRepo.DeleteByFormID(formID); err != nil {
		return utils.ErrInternal
	}
	return nil
}
