package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"

	"gorm.io/gorm"
)

type UsuarioRepository struct {
	db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

func (r *UsuarioRepository) Create(usuario *models.Usuario) error {
	if err := r.db.Create(usuario).Error; err != nil {
		return fmt.Errorf("error creando usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) GetByID(id uint) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.Preload("Regional").Preload("Sucursal").First(&usuario, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario con ID %d no encontrado", id)
		}
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}
	return &usuario, nil
}

func (r *UsuarioRepository) GetAll() ([]models.Usuario, error) {
	var usuarios []models.Usuario
	err := r.db.Preload("Regional").Preload("Sucursal").Order("nombre ASC").Find(&usuarios).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuarios: %w", err)
	}
	return usuarios, nil
}

func (r *UsuarioRepository) Update(usuario *models.Usuario) error {
	if err := r.db.Save(usuario).Error; err != nil {
		return fmt.Errorf("error actualizando usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.Usuario{}, id)
	if result.Error != nil {
		return fmt.Errorf("error eliminando usuario: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}
	return nil
}

// GetRegionales lista las regionales disponibles (La Paz, Cochabamba, Santa
// Cruz, Beni).
func (r *UsuarioRepository) GetRegionales() ([]models.Regional, error) {
	var regionales []models.Regional
	err := r.db.Order("nombre ASC").Find(&regionales).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo regionales: %w", err)
	}
	return regionales, nil
}

// GetSucursalesCatalogo lista el catálogo maestro de sucursales.
func (r *UsuarioRepository) GetSucursalesCatalogo() ([]models.SucursalCatalogo, error) {
	var sucursales []models.SucursalCatalogo
	err := r.db.Preload("Regional").Order("codigo_sucursal_sin ASC").Find(&sucursales).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo catálogo de sucursales: %w", err)
	}
	return sucursales, nil
}

// SetAccesos reemplaza por completo la configuración de accesos de un
// usuario: la bandera de acceso total y las regionales/sucursales asignadas.
func (r *UsuarioRepository) SetAccesos(usuarioID uint, accesoTotal bool, regionalesIDs, sucursalesIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Usuario{}).Where("id = ?", usuarioID).
			Update("acceso_total", accesoTotal).Error; err != nil {
			return err
		}
		if err := tx.Where("usuario_id = ?", usuarioID).Delete(&models.UsuarioAccesoRegional{}).Error; err != nil {
			return err
		}
		if err := tx.Where("usuario_id = ?", usuarioID).Delete(&models.UsuarioAccesoSucursal{}).Error; err != nil {
			return err
		}
		for _, regionalID := range regionalesIDs {
			acceso := models.UsuarioAccesoRegional{UsuarioID: usuarioID, RegionalID: regionalID}
			if err := tx.Create(&acceso).Error; err != nil {
				return err
			}
		}
		for _, sucursalID := range sucursalesIDs {
			acceso := models.UsuarioAccesoSucursal{UsuarioID: usuarioID, SucursalID: sucursalID}
			if err := tx.Create(&acceso).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAccesos devuelve las regionales y sucursales asignadas directamente a
// un usuario (sin resolver acceso total).
func (r *UsuarioRepository) GetAccesos(usuarioID uint) ([]models.UsuarioAccesoRegional, []models.UsuarioAccesoSucursal, error) {
	var regionales []models.UsuarioAccesoRegional
	if err := r.db.Preload("Regional").Where("usuario_id = ?", usuarioID).Find(&regionales).Error; err != nil {
		return nil, nil, fmt.Errorf("error obteniendo accesos por regional: %w", err)
	}

	var sucursales []models.UsuarioAccesoSucursal
	if err := r.db.Preload("Sucursal.Regional").Where("usuario_id = ?", usuarioID).Find(&sucursales).Error; err != nil {
		return nil, nil, fmt.Errorf("error obteniendo accesos por sucursal: %w", err)
	}

	return regionales, sucursales, nil
}

// SucursalesPermitidas resuelve el catálogo efectivo de sucursales a las que
// el usuario tiene acceso: todo el catálogo si tiene AccesoTotal, o la unión
// de las sucursales de sus regionales asignadas más las sucursales
// asignadas directamente.
func (r *UsuarioRepository) SucursalesPermitidas(usuarioID uint) ([]models.SucursalCatalogo, error) {
	usuario, err := r.GetByID(usuarioID)
	if err != nil {
		return nil, err
	}

	if usuario.AccesoTotal {
		return r.GetSucursalesCatalogo()
	}

	var regionalesIDs []uint
	if err := r.db.Model(&models.UsuarioAccesoRegional{}).
		Where("usuario_id = ?", usuarioID).Pluck("regional_id", &regionalesIDs).Error; err != nil {
		return nil, fmt.Errorf("error obteniendo regionales asignadas: %w", err)
	}

	var sucursalesDirectasIDs []uint
	if err := r.db.Model(&models.UsuarioAccesoSucursal{}).
		Where("usuario_id = ?", usuarioID).Pluck("sucursal_id", &sucursalesDirectasIDs).Error; err != nil {
		return nil, fmt.Errorf("error obteniendo sucursales asignadas: %w", err)
	}

	if len(regionalesIDs) == 0 && len(sucursalesDirectasIDs) == 0 {
		return []models.SucursalCatalogo{}, nil
	}

	query := r.db.Preload("Regional")
	switch {
	case len(regionalesIDs) > 0 && len(sucursalesDirectasIDs) > 0:
		query = query.Where("regional_id IN ? OR id IN ?", regionalesIDs, sucursalesDirectasIDs)
	case len(regionalesIDs) > 0:
		query = query.Where("regional_id IN ?", regionalesIDs)
	default:
		query = query.Where("id IN ?", sucursalesDirectasIDs)
	}

	var sucursales []models.SucursalCatalogo
	if err := query.Order("codigo_sucursal_sin ASC").Find(&sucursales).Error; err != nil {
		return nil, fmt.Errorf("error resolviendo sucursales permitidas: %w", err)
	}
	return sucursales, nil
}
