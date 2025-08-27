package repositories

import (
	"managerfact/internal/domain/models"

	"gorm.io/gorm"
)

type CodigoProductoRepo struct {
	db *gorm.DB
}

func NewCodigoProductoRepoRepo(db *gorm.DB) *CodigoProductoRepo {
	return &CodigoProductoRepo{
		db: db,
	}
}

func (r *CodigoProductoRepo) GetByCodigo(codigo string) (*models.Codigo_producto, error) {
	var data models.Codigo_producto
	err := r.db.Where("codigo = ?", codigo).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}
func (r *CodigoProductoRepo) GetAll() (*[]models.Codigo_producto, error) {
	var data []models.Codigo_producto
	err := r.db.Find(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}
