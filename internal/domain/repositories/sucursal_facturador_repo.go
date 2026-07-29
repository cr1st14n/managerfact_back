package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"time"

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

// ActualizarEstadoConexion registra el resultado (conectividad, no
// resultado de negocio) del último intento de envío al facturador de esta
// sucursal — usado por el circuit breaker del EnvioWorker.
func (r *SucursalFacturadorRepository) ActualizarEstadoConexion(id uint, estado string, mensaje string, momento *time.Time) error {
	err := r.db.Model(&models.SucursalFacturador{}).Where("id = ?", id).Updates(map[string]interface{}{
		"estado_conexion":       estado,
		"ultimo_error_conexion": momento,
		"ultimo_error_mensaje":  mensaje,
	}).Error
	if err != nil {
		return fmt.Errorf("error actualizando estado de conexión de la sucursal facturador: %w", err)
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
