package repository

import (
	"lite-collector/models"

	"gorm.io/gorm"
)

// BaseDataRepository defines the interface for base data access
type BaseDataRepository interface {
	Create(bd *models.BaseData) error
	FindByID(id uint64) (*models.BaseData, error)
	FindByFormID(formID uint64) ([]models.BaseData, error)
	FindByFormIDAndRowKey(formID uint64, rowKey string) (*models.BaseData, error)
	Update(bd *models.BaseData) error
	Delete(id uint64) error
	DeleteByFormID(formID uint64) error
}

type baseDataRepository struct {
	db *gorm.DB
}

func NewBaseDataRepository(db *gorm.DB) BaseDataRepository {
	return &baseDataRepository{db: db}
}

func (r *baseDataRepository) Create(bd *models.BaseData) error {
	return r.db.Create(bd).Error
}

func (r *baseDataRepository) FindByID(id uint64) (*models.BaseData, error) {
	var bd models.BaseData
	result := r.db.First(&bd, id)
	return &bd, result.Error
}

func (r *baseDataRepository) FindByFormID(formID uint64) ([]models.BaseData, error) {
	var list []models.BaseData
	result := r.db.Where("form_id = ?", formID).Order("row_key ASC").Find(&list)
	return list, result.Error
}

func (r *baseDataRepository) FindByFormIDAndRowKey(formID uint64, rowKey string) (*models.BaseData, error) {
	var bd models.BaseData
	result := r.db.Where("form_id = ? AND row_key = ?", formID, rowKey).First(&bd)
	return &bd, result.Error
}

func (r *baseDataRepository) Update(bd *models.BaseData) error {
	return r.db.Save(bd).Error
}

func (r *baseDataRepository) Delete(id uint64) error {
	return r.db.Delete(&models.BaseData{}, id).Error
}

func (r *baseDataRepository) DeleteByFormID(formID uint64) error {
	return r.db.Where("form_id = ?", formID).Delete(&models.BaseData{}).Error
}
