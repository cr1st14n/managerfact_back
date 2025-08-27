package handlers

import (
	"managerfact/aplication/services"

	"github.com/gofiber/fiber/v2"
)

type CodigoProductoHandler struct {
	codigoProductoService *services.CodigoProductoService
}

func NewCodigoProductoHandler(s *services.CodigoProductoService) *CodigoProductoHandler {
	return &CodigoProductoHandler{
		codigoProductoService: s,
	}
}
func (h *CodigoProductoHandler) GetAll(c *fiber.Ctx) error {
	data, err := h.codigoProductoService.Get()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Informacion de factura",
		"data":    data,
	})
}
func (h *CodigoProductoHandler) RegisterRoutes(router fiber.Router) {
	connections := router.Group("/codigoproducto")
	connections.Get("/", h.GetAll)
}
