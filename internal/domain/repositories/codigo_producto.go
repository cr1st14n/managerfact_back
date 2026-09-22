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
	err := r.db.Order("codigo ASC").Find(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *CodigoProductoRepo) GetByID(id uint) (*models.Codigo_producto, error) {
	var data models.Codigo_producto
	if err := r.db.First(&data, id).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *CodigoProductoRepo) Create(data *models.Codigo_producto) error {
	return r.db.Create(data).Error
}

func (r *CodigoProductoRepo) Update(data *models.Codigo_producto) error {
	return r.db.Save(data).Error
}

// Delete borra de forma definitiva (Unscoped): el índice único de "codigo"
// también cuenta las filas con borrado lógico, así que un soft delete
// impediría volver a crear el mismo código después.
func (r *CodigoProductoRepo) Delete(id uint) error {
	result := r.db.Unscoped().Delete(&models.Codigo_producto{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ExisteCodigo indica si otro registro (distinto de excluirID) ya usa ese
// código. excluirID = 0 al crear.
func (r *CodigoProductoRepo) ExisteCodigo(codigo string, excluirID uint) (bool, error) {
	var count int64
	err := r.db.Unscoped().Model(&models.Codigo_producto{}).
		Where("codigo = ? AND id <> ?", codigo, excluirID).
		Count(&count).Error
	return count > 0, err
}
