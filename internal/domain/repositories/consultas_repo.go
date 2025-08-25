package repositories

import (
	"managerfact/internal/domain/models"

	"gorm.io/gorm"
)

type ConsutasRepository struct {
	db *gorm.DB
}

func NewConsutasRepository(db *gorm.DB) *ConsutasRepository {
	return &ConsutasRepository{
		db: db,
	}
}

func (r *ConsutasRepository) GetServidorById(id int64) (*models.DbConnection, error) {
	var data models.DbConnection
	err := r.db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}
