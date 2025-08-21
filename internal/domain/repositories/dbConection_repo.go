package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"

	// "github.com/yourproject/internal/models"
	"gorm.io/gorm"
)

// DbConnectionRepository interface define los métodos del repositorio
type DbConnectionRepository interface {
	Create(connection *models.DbConnection) error
	GetByID(id uint) (*models.DbConnection, error)
	GetByServerName(serverName string) (*models.DbConnection, error)
	GetAll() ([]models.DbConnection, error)
	GetAllActive() ([]models.DbConnection, error)
	Update(connection *models.DbConnection) error
	Delete(id uint) error
	SoftDelete(id uint) error
	TestConnection(id uint) error
	Count() (int64, error)
	GetPaginated(offset, limit int) ([]models.DbConnection, int64, error)
}

// dbConnectionRepository implementación del repositorio
type dbConnectionRepository struct {
	db *gorm.DB
}

// NewDbConnectionRepository crea una nueva instancia del repositorio
func NewDbConnectionRepository(db *gorm.DB) DbConnectionRepository {
	return &dbConnectionRepository{
		db: db,
	}
}

// Create crea una nueva conexión de base de datos
func (r *dbConnectionRepository) Create(connection *models.DbConnection) error {
	if connection == nil {
		return errors.New("connection cannot be nil")
	}

	// Verificar que no exista una conexión con el mismo nombre
	var existingConnection models.DbConnection
	err := r.db.Where("server_name = ?", connection.ServerName).First(&existingConnection).Error
	if err == nil {
		return fmt.Errorf("ya existe una conexión con el nombre '%s'", connection.ServerName)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error verificando nombre único: %v", err)
	}

	// Crear la conexión
	if err := r.db.Create(connection).Error; err != nil {
		return fmt.Errorf("error creando conexión: %v", err)
	}

	return nil
}

// GetByID obtiene una conexión por su ID
func (r *dbConnectionRepository) GetByID(id uint) (*models.DbConnection, error) {
	var connection models.DbConnection
	err := r.db.First(&connection, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("conexión con ID %d no encontrada", id)
		}
		return nil, fmt.Errorf("error obteniendo conexión: %v", err)
	}

	return &connection, nil
}

// GetByServerName obtiene una conexión por su nombre de servidor
func (r *dbConnectionRepository) GetByServerName(serverName string) (*models.DbConnection, error) {
	var connection models.DbConnection
	err := r.db.Where("server_name = ?", serverName).First(&connection).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("conexión '%s' no encontrada", serverName)
		}
		return nil, fmt.Errorf("error obteniendo conexión: %v", err)
	}

	return &connection, nil
}

// GetAll obtiene todas las conexiones (incluidas las inactivas)
func (r *dbConnectionRepository) GetAll() ([]models.DbConnection, error) {
	var connections []models.DbConnection
	err := r.db.Order("server_name ASC").Find(&connections).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo conexiones: %v", err)
	}

	return connections, nil
}

// GetAllActive obtiene solo las conexiones activas
func (r *dbConnectionRepository) GetAllActive() ([]models.DbConnection, error) {
	var connections []models.DbConnection
	err := r.db.Where("is_active = ?", true).Order("server_name ASC").Find(&connections).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo conexiones activas: %v", err)
	}

	return connections, nil
}

// Update actualiza una conexión existente
func (r *dbConnectionRepository) Update(connection *models.DbConnection) error {
	if connection == nil {
		return errors.New("connection cannot be nil")
	}

	// Verificar que la conexión existe
	var existingConnection models.DbConnection
	err := r.db.First(&existingConnection, connection.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("conexión con ID %d no encontrada", connection.ID)
		}
		return fmt.Errorf("error verificando conexión: %v", err)
	}

	// Verificar nombre único (excluyendo la conexión actual)
	var duplicateConnection models.DbConnection
	err = r.db.Where("server_name = ? AND id != ?", connection.ServerName, connection.ID).First(&duplicateConnection).Error
	if err == nil {
		return fmt.Errorf("ya existe otra conexión con el nombre '%s'", connection.ServerName)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error verificando nombre único: %v", err)
	}

	// Actualizar la conexión
	if err := r.db.Save(connection).Error; err != nil {
		return fmt.Errorf("error actualizando conexión: %v", err)
	}

	return nil
}

// Delete elimina permanentemente una conexión
func (r *dbConnectionRepository) Delete(id uint) error {
	result := r.db.Unscoped().Delete(&models.DbConnection{}, id)
	if result.Error != nil {
		return fmt.Errorf("error eliminando conexión: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("conexión con ID %d no encontrada", id)
	}

	return nil
}

// SoftDelete realiza eliminación lógica de una conexión
func (r *dbConnectionRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.DbConnection{}, id)
	if result.Error != nil {
		return fmt.Errorf("error eliminando conexión: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("conexión con ID %d no encontrada", id)
	}

	return nil
}

// TestConnection verifica si es posible conectar a la base de datos
func (r *dbConnectionRepository) TestConnection(id uint) error {
	connection, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Aquí se implementaría la lógica para probar la conexión
	// Por ahora retornamos nil, se implementará en el service
	_ = connection
	return nil
}

// Count obtiene el total de conexiones
func (r *dbConnectionRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.DbConnection{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("error contando conexiones: %v", err)
	}

	return count, nil
}

// GetPaginated obtiene conexiones con paginación
func (r *dbConnectionRepository) GetPaginated(offset, limit int) ([]models.DbConnection, int64, error) {
	var connections []models.DbConnection
	var total int64

	// Obtener el total
	if err := r.db.Model(&models.DbConnection{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error contando conexiones: %v", err)
	}

	// Obtener los registros paginados
	err := r.db.Order("server_name ASC").Offset(offset).Limit(limit).Find(&connections).Error
	if err != nil {
		return nil, 0, fmt.Errorf("error obteniendo conexiones paginadas: %v", err)
	}

	return connections, total, nil
}
