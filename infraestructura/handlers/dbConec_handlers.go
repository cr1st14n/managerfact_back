package handlers

import (
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DbConnectionHandler maneja las peticiones relacionadas con conexiones de BD
type DbConnectionHandler struct {
	service services.DbConnectionService
}

// NewDbConnectionHandler crea una nueva instancia del handler
func NewDbConnectionHandler(service services.DbConnectionService) *DbConnectionHandler {
	return &DbConnectionHandler{
		service: service,
	}
}

// CreateConnectionRequest estructura para crear conexión
type CreateConnectionRequest struct {
	ServerName   string `json:"server_name" validate:"required,min=3,max=100"`
	Host         string `json:"host" validate:"required"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"required,min=1,max=100"`
	Username     string `json:"username" validate:"required,min=1,max=100"`
	Password     string `json:"password" validate:"required,min=1"`
	IsActive     *bool  `json:"is_active,omitempty"`
	Description  string `json:"description,omitempty"`
}

// UpdateConnectionRequest estructura para actualizar conexión
type UpdateConnectionRequest struct {
	ID           uint   `json:"id" validate:"required"`
	ServerName   string `json:"server_name" validate:"required,min=3,max=100"`
	Host         string `json:"host" validate:"required"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"required,min=1,max=100"`
	Username     string `json:"username" validate:"required,min=1,max=100"`
	Password     string `json:"password" validate:"required,min=1"`
	IsActive     *bool  `json:"is_active,omitempty"`
	Description  string `json:"description,omitempty"`
}

// APIResponse estructura estándar de respuesta
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreateConnection maneja la creación de nuevas conexiones
func (h *DbConnectionHandler) CreateConnection(c *fiber.Ctx) error {
	var req CreateConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Datos inválidos",
			Error:   err.Error(),
		})
	}

	// Convertir request a modelo
	connection := &models.DbConnection{
		ServerName:   req.ServerName,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
		IsActive:     true, // Por defecto activa
		Description:  req.Description,
	}

	// Si se especifica is_active, usar ese valor
	if req.IsActive != nil {
		connection.IsActive = *req.IsActive
	}

	// Crear conexión usando el servicio
	if err := h.service.CreateConnection(connection); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Error creando conexión",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "Conexión creada exitosamente",
		Data:    connection,
	})
}

// GetConnection obtiene una conexión por ID
func (h *DbConnectionHandler) GetConnection(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID inválido",
			Error:   err.Error(),
		})
	}

	connection, err := h.service.GetConnection(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "Conexión no encontrada",
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Conexión encontrada",
		Data:    connection,
	})
}

// GetAllConnections obtiene todas las conexiones
func (h *DbConnectionHandler) GetAllConnections(c *fiber.Ctx) error {
	// Verificar si solo se quieren las activas
	activeOnly := c.Query("active_only") == "true"

	var connections []models.DbConnection
	var err error

	if activeOnly {
		connections, err = h.service.GetActiveConnections()
	} else {
		connections, err = h.service.GetAllConnections()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error obteniendo conexiones",
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Conexiones obtenidas exitosamente",
		Data:    connections,
	})
}

// GetConnectionsPaginated obtiene conexiones con paginación
func (h *DbConnectionHandler) GetConnectionsPaginated(c *fiber.Ctx) error {
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	result, err := h.service.GetConnectionsPaginated(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error obteniendo conexiones paginadas",
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Conexiones obtenidas exitosamente",
		Data:    result,
	})
}

// UpdateConnection actualiza una conexión existente
func (h *DbConnectionHandler) UpdateConnection(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID inválido",
			Error:   err.Error(),
		})
	}

	var req UpdateConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Datos inválidos",
			Error:   err.Error(),
		})
	}

	// Verificar que el ID del parámetro coincida con el del body
	if uint(id) != req.ID {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID del parámetro no coincide con el ID del cuerpo",
		})
	}

	// Convertir request a modelo
	connection := &models.DbConnection{
		ID:           req.ID,
		ServerName:   req.ServerName,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
		IsActive:     true, // Por defecto activa
		Description:  req.Description,
	}

	// Si se especifica is_active, usar ese valor
	if req.IsActive != nil {
		connection.IsActive = *req.IsActive
	}

	// Actualizar conexión usando el servicio
	if err := h.service.UpdateConnection(connection); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Error actualizando conexión",
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Conexión actualizada exitosamente",
		Data:    connection,
	})
}

// DeleteConnection elimina permanentemente una conexión
func (h *DbConnectionHandler) DeleteConnection(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID inválido",
			Error:   err.Error(),
		})
	}

	if err := h.service.DeleteConnection(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Error eliminando conexión",
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Conexión eliminada exitosamente",
	})
}

// SoftDeleteConnection elimina lógicamente una conexión
func (h *DbConnectionHandler) SoftDeleteConnection(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID inválido",
			Error:   err.Error(),
		})
	}

	if err := h.service.SoftDeleteConnection(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Error desactivando conexión",
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Conexión desactivada exitosamente",
	})
}

// TestConnection prueba una conexión existente
func (h *DbConnectionHandler) TestConnection(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID inválido",
			Error:   err.Error(),
		})
	}

	result, err := h.service.TestConnection(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error probando conexión",
			Error:   err.Error(),
		})
	}

	statusCode := fiber.StatusOK
	if !result.Success {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(APIResponse{
		Success: result.Success,
		Message: "Prueba de conexión completada",
		Data:    result,
	})
}

// TestConnectionByConfig prueba una conexión sin guardarla
func (h *DbConnectionHandler) TestConnectionByConfig(c *fiber.Ctx) error {
	var req CreateConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Datos inválidos",
			Error:   err.Error(),
		})
	}

	// Convertir request a modelo
	connection := &models.DbConnection{
		ServerName:   req.ServerName,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
		IsActive:     true,
		Description:  req.Description,
	}

	result, err := h.service.TestConnectionByConfig(connection)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error probando conexión",
			Error:   err.Error(),
		})
	}

	statusCode := fiber.StatusOK
	if !result.Success {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(APIResponse{
		Success: result.Success,
		Message: "Prueba de conexión completada",
		Data:    result,
	})
}

// GetConnectionsStats obtiene estadísticas de las conexiones
func (h *DbConnectionHandler) GetConnectionsStats(c *fiber.Ctx) error {
	total, err := h.service.GetConnectionsCount()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error obteniendo estadísticas",
			Error:   err.Error(),
		})
	}

	activeConnections, err := h.service.GetActiveConnections()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error obteniendo conexiones activas",
			Error:   err.Error(),
		})
	}

	stats := fiber.Map{
		"total_connections":    total,
		"active_connections":   len(activeConnections),
		"inactive_connections": total - int64(len(activeConnections)),
	}

	return c.JSON(APIResponse{
		Success: true,
		Message: "Estadísticas obtenidas exitosamente",
		Data:    stats,
	})
}

// RegisterRoutes registra todas las rutas del handler
func (h *DbConnectionHandler) RegisterRoutes(router fiber.Router) {
	connections := router.Group("/connections")

	connections.Post("/", h.CreateConnection)
	connections.Get("/", h.GetAllConnections)
	connections.Get("/paginated", h.GetConnectionsPaginated)
	connections.Get("/stats", h.GetConnectionsStats)
	connections.Get("/:id", h.GetConnection)
	connections.Put("/:id", h.UpdateConnection)
	connections.Delete("/:id", h.DeleteConnection)
	connections.Patch("/:id/soft-delete", h.SoftDeleteConnection)
	connections.Post("/:id/test", h.TestConnection)
	connections.Post("/test", h.TestConnectionByConfig)
}
