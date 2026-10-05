package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"strings"

	"gorm.io/gorm"
)

type DbConnectionRepository interface {
	Create(connection *models.DbConnection) error
	GetByID(id uint) (*models.DbConnection, error)
	GetByServerName(serverName string) (*models.DbConnection, error)
	GetAll() ([]models.DbConnection, error)
	GetAllActive() ([]models.DbConnection, error)
	GetAllActiveByType(tipo string) ([]models.DbConnection, error)
	Update(connection *models.DbConnection) error
	Delete(id uint) error
	SoftDelete(id uint) error
	TestConnection(id uint) error
	Count() (int64, error)
	GetPaginated(offset, limit int) ([]models.DbConnection, int64, error)
}

type dbConnectionRepository struct {
	db *gorm.DB
}

func NewDbConnectionRepository(db *gorm.DB) DbConnectionRepository {
	return &dbConnectionRepository{
		db: db,
	}
}

func (r *dbConnectionRepository) Create(connection *models.DbConnection) error {
	if connection == nil {
		return errors.New("connection cannot be nil")
	}

	var existingConnection models.DbConnection
	err := r.db.Where("server_name = ?", connection.ServerName).First(&existingConnection).Error
	if err == nil {
		return fmt.Errorf("ya existe una conexión con el nombre '%s'", connection.ServerName)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error verificando nombre único: %v", err)
	}

	if err := r.db.Create(connection).Error; err != nil {
		return fmt.Errorf("error creando conexión: %v", err)
	}

	return nil
}

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

func (r *dbConnectionRepository) GetAll() ([]models.DbConnection, error) {
	var connections []models.DbConnection
	err := r.db.Order("server_name ASC").Find(&connections).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo conexiones: %v", err)
	}

	return connections, nil
}

func (r *dbConnectionRepository) GetAllActive() ([]models.DbConnection, error) {
	var connections []models.DbConnection
	err := r.db.Where("is_active = ?", true).Order("server_name ASC").Find(&connections).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo conexiones activas: %v", err)
	}

	return connections, nil
}

func (r *dbConnectionRepository) GetAllActiveByType(tipo string) ([]models.DbConnection, error) {
	var connections []models.DbConnection
	err := r.db.Where("is_active = ? AND LOWER(type) LIKE ?", true, "%"+strings.ToLower(tipo)+"%").Order("server_name ASC").Find(&connections).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo conexiones activas por tipo: %v", err)
	}

	return connections, nil
}

func (r *dbConnectionRepository) Update(connection *models.DbConnection) error {
	if connection == nil {
		return errors.New("connection cannot be nil")
	}

	var existingConnection models.DbConnection
	err := r.db.First(&existingConnection, connection.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("conexión con ID %d no encontrada", connection.ID)
		}
		return fmt.Errorf("error verificando conexión: %v", err)
	}

	var duplicateConnection models.DbConnection
	err = r.db.Where("server_name = ? AND id != ?", connection.ServerName, connection.ID).First(&duplicateConnection).Error
	if err == nil {
		return fmt.Errorf("ya existe otra conexión con el nombre '%s'", connection.ServerName)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error verificando nombre único: %v", err)
	}

	if err := r.db.Save(connection).Error; err != nil {
		return fmt.Errorf("error actualizando conexión: %v", err)
	}

	return nil
}

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

func (r *dbConnectionRepository) TestConnection(id uint) error {
	connection, err := r.GetByID(id)
	if err != nil {
		return err
	}

	_ = connection
	return nil
}

func (r *dbConnectionRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.DbConnection{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("error contando conexiones: %v", err)
	}

	return count, nil
}

func (r *dbConnectionRepository) GetPaginated(offset, limit int) ([]models.DbConnection, int64, error) {
	var connections []models.DbConnection
	var total int64

	if err := r.db.Model(&models.DbConnection{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error contando conexiones: %v", err)
	}

	err := r.db.Order("server_name ASC").Offset(offset).Limit(limit).Find(&connections).Error
	if err != nil {
		return nil, 0, fmt.Errorf("error obteniendo conexiones paginadas: %v", err)
	}

	return connections, total, nil
}
