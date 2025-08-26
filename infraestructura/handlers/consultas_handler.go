package handlers

import (
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"
	"managerfact/pkg/utils"

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

	var errValidacion []string
	dataIn.IdFacturador = utils.ValidarCampoRequerido(&errValidacion, dataIn.IdFacturador, "El campo idServer es requerido")
	if dataIn.NumeroFactura != "" {
		utils.ValidarEntero(&errValidacion, dataIn.NumeroFactura, "El campo numeroFactura es requerido")
	}
	dataIn.CodigoProducto = utils.ValidarCampoOpcional(&errValidacion, dataIn.CodigoProducto)
	FechaDesde_parse := utils.ValidarFecha(&errValidacion, dataIn.FechaDesde, "El campo fechaDesde es requerido")
	FechaHasta_parse := utils.ValidarFecha(&errValidacion, dataIn.FechaHasta, "El campo fechaHasta es requerido")
	dataIn.Sucursal = utils.ValidarCampoOpcional(&errValidacion, dataIn.Sucursal)
	dataIn.FechaDesde = FechaDesde_parse.Format("2006-01-02")
	dataIn.FechaHasta = FechaHasta_parse.Format("2006-01-02")

	if len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			"errors":  errValidacion,
		})
	}

	data, errDaS := h.ConsultasService.DataFacturas(dataIn)
	if errDaS != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   errDaS.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Informacion de factura",
		"data":    data,
	})
}

func (h *ConsultasHandler) Sucursales(c *fiber.Ctx) error {
	var idServer = c.Query("idServer")
	data, err := h.ConsultasService.Sucursales(idServer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sucursales registradas",
		"data":    data,
	})
	// return nil
}

func (h *ConsultasHandler) RegisterRoutes(router fiber.Router) {
	connections := router.Group("/consultar")

	connections.Post("/", h.DataFacturas)
	connections.Get("/sucursales", h.Sucursales)
}
