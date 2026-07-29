package handlers

import (
	"managerfact/internal/domain/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// LogEnvioHandler expone en solo lectura el registro de intentos de envío
// (logs_envio), para que el front pueda mostrar que el EnvioWorker
// realmente está corriendo.
type LogEnvioHandler struct {
	repo *repositories.LogEnvioRepository
}

func NewLogEnvioHandler(repo *repositories.LogEnvioRepository) *LogEnvioHandler {
	return &LogEnvioHandler{repo: repo}
}

func (h *LogEnvioHandler) GetAll(c *fiber.Ctx) error {
	filtro := repositories.LogEnvioFiltro{
		Tipo:      c.Query("tipo"),
		Resultado: c.Query("resultado"),
		Origen:    c.Query("origen"),
	}
	if sucursalID, err := strconv.ParseUint(c.Query("sucursal_facturador_id"), 10, 32); err == nil {
		filtro.SucursalFacturadorID = uint(sucursalID)
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filtro.Limit = limit
	}

	logs, err := h.repo.GetAll(filtro)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo logs de envío", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Logs de envío obtenidos exitosamente", "data": logs})
}

func (h *LogEnvioHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/logs-envio", h.GetAll)
}
