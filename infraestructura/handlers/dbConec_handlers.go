package handlers

import (
	"fmt"
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type DbConnectionHandler struct {
	service services.DbConnectionService
}

func NewDbConnectionHandler(service services.DbConnectionService) *DbConnectionHandler {
	return &DbConnectionHandler{
		service: service,
	}
}

type CreateConnectionRequest struct {
	ServerName   string `json:"server_name" validate:"required,min=3,max=100"`
	Host         string `json:"host" validate:"required"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"required,min=1,max=100"`
	Username     string `json:"username" validate:"required,min=1,max=100"`
	Password     string `json:"password" validate:"required,min=1"`
	IsActive     *bool  `json:"is_active,omitempty"`
	Description  string `json:"description,omitempty"`
	Type         string `json:"type" validate:"required,min=1,max=50"`

	Ambiente string `json:"ambiente,omitempty"`
}

type UpdateConnectionRequest struct {
	ID           uint   `json:"id" validate:"required"`
	ServerName   string `json:"server_name" validate:"required,min=3,max=100"`
	Host         string `json:"host" validate:"required"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"required,min=1,max=100"`
	Username     string `json:"username" validate:"required,min=1,max=100"`
		// Password vacío conserva la guardada; solo se devuelve tras validar la conexión nueva.
	Password    string `json:"password,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type" validate:"required,min=1,max=50"`

	Ambiente string `json:"ambiente,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *DbConnectionHandler) CreateConnection(c *fiber.Ctx) error {
	var req CreateConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Datos inválidos",
			Error:   err.Error(),
		})
	}

	ambiente, ok := models.NormalizarAmbiente(req.Ambiente)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Ambiente inválido",
			Error:   "ambiente debe ser produccion, baja o test",
		})
	}

	connection := &models.DbConnection{
		ServerName:   req.ServerName,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
		IsActive:     true, 
		Description:  req.Description,
		Type:         req.Type,
		Ambiente:     ambiente,
	}

	if req.IsActive != nil {
		connection.IsActive = *req.IsActive
	}

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

func (h *DbConnectionHandler) GetAllConnections(c *fiber.Ctx) error {

	activeOnly := c.Query("active_only") == "true"
	tipo := c.Query("tipo")

	var connections []models.DbConnection
	var err error

	if activeOnly {
		fmt.Printf("Obteniendo solo conexiones activas del tipo '%s'\n", tipo)
		connections, err = h.service.GetActiveConnections(tipo)
	} else {
		fmt.Println("Obteniendo todas las conexiones")
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

	if uint(id) != req.ID {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "ID del parámetro no coincide con el ID del cuerpo",
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
	connection.ServerName = req.ServerName
	connection.Host = req.Host
	connection.Port = req.Port
	connection.DatabaseName = req.DatabaseName
	connection.Username = req.Username
	connection.Description = req.Description
	connection.Type = req.Type
	if req.Password != "" {
		connection.Password = req.Password
	}
	if req.Ambiente != "" {
		ambiente, ok := models.NormalizarAmbiente(req.Ambiente)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
				Success: false,
				Message: "Ambiente inválido",
				Error:   "ambiente debe ser produccion, baja o test",
			})
		}
		connection.Ambiente = ambiente
	}
	if req.IsActive != nil {
		connection.IsActive = *req.IsActive
	}

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

func (h *DbConnectionHandler) TestConnectionByConfig(c *fiber.Ctx) error {
	var req CreateConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Datos inválidos",
			Error:   err.Error(),
		})
	}

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

func (h *DbConnectionHandler) GetConnectionsStats(c *fiber.Ctx) error {
	total, err := h.service.GetConnectionsCount()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Error obteniendo estadísticas",
			Error:   err.Error(),
		})
	}

	activeConnections, err := h.service.GetActiveConnections("facturador")
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

func (h *DbConnectionHandler) RegisterRoutes(router fiber.Router, requireAdmin fiber.Handler) {
	connections := router.Group("/connections")

	connections.Get("/", h.GetAllConnections)

	connections.Post("/test", requireAdmin, h.TestConnectionByConfig)
	connections.Get("/stats", requireAdmin, h.GetConnectionsStats)
	connections.Get("/paginated", requireAdmin, h.GetConnectionsPaginated)
	connections.Post("/", requireAdmin, h.CreateConnection)
	connections.Get("/:id", requireAdmin, h.GetConnection)
	connections.Put("/:id", requireAdmin, h.UpdateConnection)
	connections.Delete("/:id", requireAdmin, h.DeleteConnection)
	connections.Patch("/:id/soft-delete", requireAdmin, h.SoftDeleteConnection)
	connections.Post("/:id/test", requireAdmin, h.TestConnection)
}
