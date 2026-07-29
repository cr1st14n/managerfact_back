package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"

	"gorm.io/gorm"
)

type SucursalFacturadorRepository struct {
	db *gorm.DB
}

func NewSucursalFacturadorRepository(db *gorm.DB) *SucursalFacturadorRepository {
	return &SucursalFacturadorRepository{db: db}
}

func (r *SucursalFacturadorRepository) Create(sucursal *models.SucursalFacturador) error {
	if err := r.db.Create(sucursal).Error; err != nil {
		return fmt.Errorf("error creando sucursal facturador: %w", err)
	}
	return nil
}

func (r *SucursalFacturadorRepository) GetByID(id uint) (*models.SucursalFacturador, error) {
	var sucursal models.SucursalFacturador
	err := r.db.First(&sucursal, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("sucursal facturador con ID %d no encontrada", id)
		}
		return nil, fmt.Errorf("error obteniendo sucursal facturador: %w", err)
	}
	return &sucursal, nil
}

func (r *SucursalFacturadorRepository) GetAll() ([]models.SucursalFacturador, error) {
	var sucursales []models.SucursalFacturador
	err := r.db.Order("nombre ASC").Find(&sucursales).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sucursales facturador: %w", err)
	}
	return sucursales, nil
}

func (r *SucursalFacturadorRepository) Update(sucursal *models.SucursalFacturador) error {
	if err := r.db.Save(sucursal).Error; err != nil {
		return fmt.Errorf("error actualizando sucursal facturador: %w", err)
	}
	return nil
}

func (r *SucursalFacturadorRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.SucursalFacturador{}, id)
	if result.Error != nil {
		return fmt.Errorf("error eliminando sucursal facturador: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("sucursal facturador con ID %d no encontrada", id)
	}
	return nil
}
