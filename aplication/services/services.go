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

)

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

type ConnectionTestResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ResponseTime time.Duration `json:"response_time"`
	ServerInfo   *ServerInfo   `json:"server_info,omitempty"`
	Error        string        `json:"error,omitempty"`
}

type ServerInfo struct {
	Version     string `json:"version"`
	ProductName string `json:"product_name"`
	Edition     string `json:"edition"`
}

type PaginatedResponse struct {
	Data       []models.DbConnection `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

type dbConnectionService struct {
	repo      repositories.DbConnectionRepository
	validator *validator.Validate
}

func NewDbConnectionService(repo repositories.DbConnectionRepository) DbConnectionService {
	return &dbConnectionService{
		repo:      repo,
		validator: validator.New(),
	}
}

func (s *dbConnectionService) CreateConnection(connection *models.DbConnection) error {
	if connection == nil {
		return fmt.Errorf("la conexión no puede ser nula")
	}

	if err := s.validator.Struct(connection); err != nil {
		return fmt.Errorf("datos de conexión inválidos: %v", err)
	}

	if !connection.IsValid() {
		return fmt.Errorf("faltan campos requeridos en la conexión")
	}

	if err := s.verificarHost(connection); err != nil {
		return fmt.Errorf("no se pudo verificar el servidor: %v", err)
	}

	if err := s.repo.Create(connection); err != nil {
		return fmt.Errorf("error guardando conexión: %v", err)
	}

	log.Printf("Conexión '%s' creada exitosamente", connection.ServerName)
	return nil
}

func (s *dbConnectionService) verificarHost(connection *models.DbConnection) error {
	err := pingHost(connection.Host)
	if errors.Is(err, errPingNoDisponible) {
		log.Printf("Aviso: no hay comando ping en este equipo; no se verificó el host %s", connection.Host)
		return nil
	}
	return err
}

func (s *dbConnectionService) GetConnection(id uint) (*models.DbConnection, error) {
	if id == 0 {
		return nil, fmt.Errorf("ID de conexión inválido")
	}

	return s.repo.GetByID(id)
}

func (s *dbConnectionService) GetConnectionByName(serverName string) (*models.DbConnection, error) {
	if serverName == "" {
		return nil, fmt.Errorf("nombre de servidor no puede estar vacío")
	}

	return s.repo.GetByServerName(serverName)
}

func (s *dbConnectionService) GetAllConnections() ([]models.DbConnection, error) {
	return s.repo.GetAll()
}

func (s *dbConnectionService) GetActiveConnections(tipo string) ([]models.DbConnection, error) {
	if tipo != "" {
		return s.repo.GetAllActiveByType(tipo)
	}
	return s.repo.GetAllActive()
}

func (s *dbConnectionService) UpdateConnection(connection *models.DbConnection) error {
	if connection == nil {
		return fmt.Errorf("la conexión no puede ser nula")
	}

	if connection.ID == 0 {
		return fmt.Errorf("ID de conexión requerido para actualización")
	}

	if err := s.validator.Struct(connection); err != nil {
		return fmt.Errorf("datos de conexión inválidos: %v", err)
	}

	existente, err := s.repo.GetByID(connection.ID)
	if err != nil {
		return err
	}

	if connection.IsActive && existente.Host != connection.Host {
		if err := s.verificarHost(connection); err != nil {
			return fmt.Errorf("no se pudo verificar el servidor: %v", err)
		}
	}

	if err := s.repo.Update(connection); err != nil {
		return fmt.Errorf("error actualizando conexión: %v", err)
	}

	log.Printf("Conexión '%s' actualizada exitosamente", connection.ServerName)
	return nil
}

func (s *dbConnectionService) DeleteConnection(id uint) error {
	if id == 0 {
		return fmt.Errorf("ID de conexión inválido")
	}

	connection, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("error eliminando conexión: %v", err)
	}

	log.Printf("Conexión '%s' eliminada permanentemente", connection.ServerName)
	return nil
}

func (s *dbConnectionService) SoftDeleteConnection(id uint) error {
	if id == 0 {
		return fmt.Errorf("ID de conexión inválido")
	}

	connection, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDelete(id); err != nil {
		return fmt.Errorf("error eliminando conexión: %v", err)
	}

	log.Printf("Conexión '%s' desactivada", connection.ServerName)
	return nil
}

func (s *dbConnectionService) TestConnection(id uint) (*ConnectionTestResult, error) {
	connection, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.TestConnectionByConfig(connection)
}

func (s *dbConnectionService) TestConnectionByConfig(connection *models.DbConnection) (*ConnectionTestResult, error) {
	result := &ConnectionTestResult{
		Success: false,
		Message: "",
	}

	start := time.Now()
	defer func() {
		result.ResponseTime = time.Since(start)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	connString := connection.ConnectionString()

	db, err := sql.Open("mssql", connString)
	if err != nil {
		result.Message = "Error abriendo conexión"
		result.Error = err.Error()
		return result, nil
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 30)

	if err := db.PingContext(ctx); err != nil {
		result.Message = "Error de conectividad ping"
		result.Error = err.Error()
		return result, nil
	}

	serverInfo, err := s.getServerInfo(ctx, db)
	if err != nil {
		result.Message = "Conexión establecida pero error obteniendo información del servidor"
		result.Error = err.Error()
		result.Success = true 
		return result, nil
	}

	result.Success = true
	result.Message = "Conexión exitosa"
	result.ServerInfo = serverInfo

	return result, nil
}

func (s *dbConnectionService) getServerInfo(ctx context.Context, db *sql.DB) (*ServerInfo, error) {
	info := &ServerInfo{}

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

func (s *dbConnectionService) GetConnectionsCount() (int64, error) {
	return s.repo.Count()
}
