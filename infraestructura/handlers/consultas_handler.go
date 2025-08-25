package handlers

import (
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"

	"github.com/gofiber/fiber/v2"
)

type ConsultasHandler struct {
	services.ConsultasService
}

func NewConsultasHandler(s *services.ConsultasService) *ConsultasHandler {
	return &ConsultasHandler{
		ConsultasService: *s,
	}
}
func (h *ConsultasHandler) DataFacturas(c *fiber.Ctx) error {
	var dataIn models.Json_consulta_data

	if err := c.BodyParser(&dataIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			"error":   err.Error(),
		})
	}
	if dataIn.IdFacturador == "" || dataIn.NumeroDocumento == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			// "error":   err.Error(),
		})

	}

	data, errDaS := h.ConsultasService.DataFacturas(dataIn)
	if errDaS != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al validar los datos",
			"error":   errDaS.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Informacion de factura",
		"data":    data,
	})
}

func (h *ConsultasHandler) RegisterRoutes(router fiber.Router) {
	connections := router.Group("/consultar")

	connections.Post("/", h.DataFacturas)
}
