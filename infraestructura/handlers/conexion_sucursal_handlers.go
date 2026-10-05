package handlers

import (
	"errors"
	"managerfact/aplication/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ConexionSucursalHandler struct {
	service *services.ConexionSucursalService
}

func NewConexionSucursalHandler(s *services.ConexionSucursalService) *ConexionSucursalHandler {
	return &ConexionSucursalHandler{service: s}
}

func (h *ConexionSucursalHandler) Sincronizar(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}

	resultado, err := h.service.Sincronizar(uint(id))
	if err != nil {
		status := fiber.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrConexionInvalida):
			status = fiber.StatusBadRequest
		case errors.Is(err, services.ErrOrigenSucursales):
			status = fiber.StatusBadGateway
		}
		return c.Status(status).JSON(fiber.Map{"message": "Error actualizando las sucursales", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Sucursales actualizadas exitosamente", "data": resultado})
}

func (h *ConexionSucursalHandler) Resumen(c *fiber.Ctx) error {
	resumen, err := h.service.Resumen()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo el resumen de sucursales", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Resumen de sucursales", "data": resumen})
}

func (h *ConexionSucursalHandler) ListarTodas(c *fiber.Ctx) error {
	data, err := h.service.ListarTodas()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo las sucursales", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sucursales registradas", "data": data})
}

func (h *ConexionSucursalHandler) RegisterRoutes(router fiber.Router, requireAdmin fiber.Handler) {
	connections := router.Group("/connections")
	connections.Get("/sucursales/resumen", requireAdmin, h.Resumen)
	connections.Post("/:id/sucursales/sync", requireAdmin, h.Sincronizar)

	router.Group("/consultar").Get("/sucursales-todas", h.ListarTodas)
}
