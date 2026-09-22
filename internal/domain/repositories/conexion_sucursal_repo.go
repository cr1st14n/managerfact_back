package repositories

import (
	"managerfact/internal/domain/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConexionSucursalRepo struct {
	db *gorm.DB
}

func NewConexionSucursalRepo(db *gorm.DB) *ConexionSucursalRepo {
	return &ConexionSucursalRepo{db: db}
}

// ResumenSucursales es lo que muestra /conexiones por cada conexión: cuántas
// sucursales hay copiadas y cuándo se actualizó por última vez.
type ResumenSucursales struct {
	ConexionID    uint      `json:"conexion_id"`
	Total         int64     `json:"total"`
	ActualizadoAt time.Time `json:"actualizado_at"`
}

func (r *ConexionSucursalRepo) ListByConexion(conexionID uint) ([]models.ConexionSucursal, error) {
	var filas []models.ConexionSucursal
	err := r.db.Where("conexion_id = ?", conexionID).Order("sfe_id ASC").Find(&filas).Error
	return filas, err
}

func (r *ConexionSucursalRepo) Resumen() ([]ResumenSucursales, error) {
	var resumen []ResumenSucursales
	err := r.db.Model(&models.ConexionSucursal{}).
		Select("conexion_id, COUNT(*) AS total, MAX(actualizado_at) AS actualizado_at").
		Group("conexion_id").
		Scan(&resumen).Error
	return resumen, err
}

// Upsert inserta las sucursales nuevas y actualiza las existentes de una
// conexión, y devuelve cuántas eran nuevas. Nunca borra: una sucursal que
// desaparezca del origen se queda en la copia.
func (r *ConexionSucursalRepo) Upsert(conexionID uint, filas []models.ConexionSucursal) (int, error) {
	if len(filas) == 0 {
		return 0, nil
	}

	nuevas := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var antes, despues int64
		if err := tx.Model(&models.ConexionSucursal{}).Where("conexion_id = ?", conexionID).Count(&antes).Error; err != nil {
			return err
		}

		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "conexion_id"}, {Name: "sfe_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"codigo_sucursal", "codigo_sucursal_sin", "nombre", "direccion",
				"municipio_departamento", "estado_sucursal", "actualizado_at",
			}),
		}).CreateInBatches(filas, 200).Error
		if err != nil {
			return err
		}

		if err := tx.Model(&models.ConexionSucursal{}).Where("conexion_id = ?", conexionID).Count(&despues).Error; err != nil {
			return err
		}
		nuevas = int(despues - antes)
		return nil
	})
	return nuevas, err
}

// ListByConexiones trae la copia local de varias conexiones de una sola vez
// (una consulta, no una por conexión).
func (r *ConexionSucursalRepo) ListByConexiones(conexionIDs []uint) ([]models.ConexionSucursal, error) {
	var filas []models.ConexionSucursal
	if len(conexionIDs) == 0 {
		return filas, nil
	}
	err := r.db.Where("conexion_id IN ?", conexionIDs).Order("nombre ASC, sfe_id ASC").Find(&filas).Error
	return filas, err
}
