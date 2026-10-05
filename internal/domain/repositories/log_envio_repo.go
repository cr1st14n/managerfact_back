package repositories

import (
	"fmt"
	"managerfact/internal/domain/models"

	"gorm.io/gorm"
)

type LogEnvioRepository struct {
	db *gorm.DB
}

func NewLogEnvioRepository(db *gorm.DB) *LogEnvioRepository {
	return &LogEnvioRepository{db: db}
}

func (r *LogEnvioRepository) Create(log *models.LogEnvio) error {
	if err := r.db.Create(log).Error; err != nil {
		return fmt.Errorf("error creando log de envío: %w", err)
	}
	return nil
}

type LogEnvioFiltro struct {
	Tipo                 string
	Resultado            string
	Origen               string
	SucursalFacturadorID uint
	Limit                int
}

func (r *LogEnvioRepository) GetAll(filtro LogEnvioFiltro) ([]models.LogEnvio, error) {
	logs := []models.LogEnvio{}
	query := r.db.Preload("SucursalFacturador")

	if filtro.Tipo != "" {
		query = query.Where("tipo = ?", filtro.Tipo)
	}
	if filtro.Resultado != "" {
		query = query.Where("resultado = ?", filtro.Resultado)
	}
	if filtro.Origen != "" {
		query = query.Where("origen = ?", filtro.Origen)
	}
	if filtro.SucursalFacturadorID != 0 {
		query = query.Where("sucursal_facturador_id = ?", filtro.SucursalFacturadorID)
	}

	limit := filtro.Limit
	if limit <= 0 {
		limit = 200
	}

	if err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("error obteniendo logs de envío: %w", err)
	}
	return logs, nil
}
