package services

import (
	"context"
	"database/sql"
	"fmt"
	"managerfact/internal/domain/models"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

type TestConnectionResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ResponseTime int64  `json:"response_time"` 
	Error        string `json:"error,omitempty"`
}

func TestConnection(config *models.DbConnection) *TestConnectionResult {
	start := time.Now()
	result := &TestConnectionResult{Success: false}

	connString := fmt.Sprintf(
		"server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=false;connection timeout=10",
		config.Host, config.Port, config.DatabaseName, config.Username, config.Password,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("mssql", connString)
	if err != nil {
		result.Message = "Error abriendo conexión"
		result.Error = err.Error()
		result.ResponseTime = time.Since(start).Milliseconds()
		return result
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		result.Message = "Error de conectividad"
		result.Error = err.Error()
		result.ResponseTime = time.Since(start).Milliseconds()
		return result
	}

	var testValue int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&testValue)
	if err != nil {
		result.Message = "Conexión OK pero error en query"
		result.Error = err.Error()
		result.Success = true 
		result.ResponseTime = time.Since(start).Milliseconds()
		return result
	}

	result.Success = true
	result.Message = "Conexión exitosa"
	result.ResponseTime = time.Since(start).Milliseconds()
	return result
}

func (s *dbConnectionService) TestConnectionByID(id uint) *TestConnectionResult {
	connection, err := s.GetConnection(id)
	if err != nil {
		return &TestConnectionResult{
			Success: false,
			Message: "Conexión no encontrada",
			Error:   err.Error(),
		}
	}

	return TestConnection(connection)
}
