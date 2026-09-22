package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"time"

	"github.com/go-playground/validator/v10"
	// "github.com/johnfercher/maroto/v2/pkg/repository"
	// "github.com/go-playground/validator/v10"
	// "github.com/yourproject/internal/models"
	// "github.com/yourproject/internal/repository"
	// _ "github.com/microsoft/go-mssqldb" // Driver de SQL Server
)

// DbConnectionService interface define los métodos del servicio
type DbConnectionService interface {
	CreateConnection(connection *models.DbConnection) error
	GetConnection(id uint) (*models.DbConnection, error)
	GetConnectionByName(serverName string) (*models.DbConnection, error)
	GetAllConnections() ([]models.DbConnection, error)
	GetActiveConnections(tipo string) ([]models.DbConnection, error)
	UpdateConnection(connection *models.DbConnection) error
	DeleteConnection(id uint) error
	SoftDeleteConnection(id uint) error
	TestConnection(id uint) (*ConnectionTestResult, error)
	TestConnectionByConfig(connection *models.DbConnection) (*ConnectionTestResult, error)
	GetConnectionsPaginated(page, pageSize int) (*PaginatedResponse, error)
	GetConnectionsCount() (int64, error)
}

// ConnectionTestResult representa el resultado de una prueba de conexión
type ConnectionTestResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ResponseTime time.Duration `json:"response_time"`
	ServerInfo   *ServerInfo   `json:"server_info,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// ServerInfo información del servidor SQL Server
type ServerInfo struct {
	Version     string `json:"version"`
	ProductName string `json:"product_name"`
	Edition     string `json:"edition"`
}

// PaginatedResponse respuesta paginada
type PaginatedResponse struct {
	Data       []models.DbConnection `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// dbConnectionService implementación del servicio
type dbConnectionService struct {
	repo      repositories.DbConnectionRepository
	validator *validator.Validate
}

// NewDbConnectionService crea una nueva instancia del servicio
func NewDbConnectionService(repo repositories.DbConnectionRepository) DbConnectionService {
	return &dbConnectionService{
		repo:      repo,
		validator: validator.New(),
	}
}

// CreateConnection crea una nueva conexión
func (s *dbConnectionService) CreateConnection(connection *models.DbConnection) error {
	if connection == nil {
		return fmt.Errorf("la conexión no puede ser nula")
	}

	// Validar la estructura
	if err := s.validator.Struct(connection); err != nil {
		return fmt.Errorf("datos de conexión inválidos: %v", err)
	}

	// Validar que los campos requeridos no estén vacíos
	if !connection.IsValid() {
		return fmt.Errorf("faltan campos requeridos en la conexión")
	}

	// Al guardar solo se verifica con un ping que el servidor responde; el
	// login y el puerto se validan con "Probar".
	if err := s.verificarHost(connection); err != nil {
		return fmt.Errorf("no se pudo verificar el servidor: %v", err)
	}

	// Guardar en el repositorio
	if err := s.repo.Create(connection); err != nil {
		return fmt.Errorf("error guardando conexión: %v", err)
	}

	log.Printf("Conexión '%s' creada exitosamente", connection.ServerName)
	return nil
}

// verificarHost comprueba con un ping que el servidor responde (sin validar
// puerto ni credenciales). Si el backend no tiene el comando ping no se puede
// verificar y no se bloquea el guardado.
func (s *dbConnectionService) verificarHost(connection *models.DbConnection) error {
	err := pingHost(connection.Host)
	if errors.Is(err, errPingNoDisponible) {
		log.Printf("Aviso: no hay comando ping en este equipo; no se verificó el host %s", connection.Host)
		return nil
	}
	return err
}

// GetConnection obtiene una conexión por ID
func (s *dbConnectionService) GetConnection(id uint) (*models.DbConnection, error) {
	if id == 0 {
		return nil, fmt.Errorf("ID de conexión inválido")
	}

	return s.repo.GetByID(id)
}

// GetConnectionByName obtiene una conexión por nombre
func (s *dbConnectionService) GetConnectionByName(serverName string) (*models.DbConnection, error) {
	if serverName == "" {
		return nil, fmt.Errorf("nombre de servidor no puede estar vacío")
	}

	return s.repo.GetByServerName(serverName)
}

// GetAllConnections obtiene todas las conexiones
func (s *dbConnectionService) GetAllConnections() ([]models.DbConnection, error) {
	return s.repo.GetAll()
}

// GetActiveConnections obtiene solo las conexiones activas
func (s *dbConnectionService) GetActiveConnections(tipo string) ([]models.DbConnection, error) {
	if tipo != "" {
		return s.repo.GetAllActiveByType(tipo)
	}
	return s.repo.GetAllActive()
}

// UpdateConnection actualiza una conexión existente
func (s *dbConnectionService) UpdateConnection(connection *models.DbConnection) error {
	if connection == nil {
		return fmt.Errorf("la conexión no puede ser nula")
	}

	if connection.ID == 0 {
		return fmt.Errorf("ID de conexión requerido para actualización")
	}

	// Validar la estructura
	if err := s.validator.Struct(connection); err != nil {
		return fmt.Errorf("datos de conexión inválidos: %v", err)
	}

	// Verificar que la conexión existe
	existente, err := s.repo.GetByID(connection.ID)
	if err != nil {
		return err
	}

	// Solo se hace ping si cambió el host (para atrapar un error de tipeo):
	// reclasificar el ambiente, cambiar la contraseña o la descripción no
	// depende de que el servidor responda en este momento.
	if connection.IsActive && existente.Host != connection.Host {
		if err := s.verificarHost(connection); err != nil {
			return fmt.Errorf("no se pudo verificar el servidor: %v", err)
		}
	}

	// Actualizar en el repositorio
	if err := s.repo.Update(connection); err != nil {
		return fmt.Errorf("error actualizando conexión: %v", err)
	}

	log.Printf("Conexión '%s' actualizada exitosamente", connection.ServerName)
	return nil
}

// DeleteConnection elimina permanentemente una conexión
func (s *dbConnectionService) DeleteConnection(id uint) error {
	if id == 0 {
		return fmt.Errorf("ID de conexión inválido")
	}

	// Verificar que la conexión existe
	connection, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Eliminar del repositorio
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("error eliminando conexión: %v", err)
	}

	log.Printf("Conexión '%s' eliminada permanentemente", connection.ServerName)
	return nil
}

// SoftDeleteConnection elimina lógicamente una conexión
func (s *dbConnectionService) SoftDeleteConnection(id uint) error {
	if id == 0 {
		return fmt.Errorf("ID de conexión inválido")
	}

	// Verificar que la conexión existe
	connection, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Eliminar lógicamente
	if err := s.repo.SoftDelete(id); err != nil {
		return fmt.Errorf("error eliminando conexión: %v", err)
	}

	log.Printf("Conexión '%s' desactivada", connection.ServerName)
	return nil
}

// TestConnection prueba una conexión existente
func (s *dbConnectionService) TestConnection(id uint) (*ConnectionTestResult, error) {
	connection, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.TestConnectionByConfig(connection)
}

// TestConnectionByConfig prueba una conexión usando configuración
func (s *dbConnectionService) TestConnectionByConfig(connection *models.DbConnection) (*ConnectionTestResult, error) {
	result := &ConnectionTestResult{
		Success: false,
		Message: "",
	}

	start := time.Now()
	defer func() {
		result.ResponseTime = time.Since(start)
	}()

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Construir connection string
	connString := connection.ConnectionString()

	// Intentar conexión
	db, err := sql.Open("mssql", connString)
	if err != nil {
		result.Message = "Error abriendo conexión"
		result.Error = err.Error()
		return result, nil
	}
	defer db.Close()

	// Configurar timeouts
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 30)

	// Ping de conectividad
	if err := db.PingContext(ctx); err != nil {
		result.Message = "Error de conectividad ping"
		result.Error = err.Error()
		return result, nil
	}

	// Query de prueba para obtener información del servidor
	serverInfo, err := s.getServerInfo(ctx, db)
	if err != nil {
		result.Message = "Conexión establecida pero error obteniendo información del servidor"
		result.Error = err.Error()
		result.Success = true // La conexión funciona aunque no podamos obtener info
		return result, nil
	}

	result.Success = true
	result.Message = "Conexión exitosa"
	result.ServerInfo = serverInfo

	return result, nil
}

// getServerInfo obtiene información del servidor SQL Server
func (s *dbConnectionService) getServerInfo(ctx context.Context, db *sql.DB) (*ServerInfo, error) {
	info := &ServerInfo{}

	// Query para obtener versión y información del servidor
	query := `
		SELECT 
			@@VERSION as version,
			SERVERPROPERTY('ProductName') as product_name,
			SERVERPROPERTY('Edition') as edition
	`

	row := db.QueryRowContext(ctx, query)
	err := row.Scan(&info.Version, &info.ProductName, &info.Edition)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información del servidor: %v", err)
	}

	return info, nil
}

// GetConnectionsPaginated obtiene conexiones con paginación
func (s *dbConnectionService) GetConnectionsPaginated(page, pageSize int) (*PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	connections, total, err := s.repo.GetPaginated(offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResponse{
		Data:       connections,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetConnectionsCount obtiene el total de conexiones
func (s *dbConnectionService) GetConnectionsCount() (int64, error) {
	return s.repo.Count()
}
