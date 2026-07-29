package handlers

import (
	"managerfact/aplication/services"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type FacturaPrevaloradaHandler struct {
	service *services.FacturaPrevaloradaService
}

func NewFacturaPrevaloradaHandler(s *services.FacturaPrevaloradaService) *FacturaPrevaloradaHandler {
	return &FacturaPrevaloradaHandler{service: s}
}

// ImportarExcel recibe el archivo .xlsx de boletos (multipart, campo
// "archivo") junto con la sucursal_facturador_id elegida para todo el lote.
func (h *FacturaPrevaloradaHandler) ImportarExcel(c *fiber.Ctx) error {
	sucursalFacturadorID, err := strconv.ParseUint(c.FormValue("sucursal_facturador_id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "sucursal_facturador_id es requerido y debe ser numérico"})
	}

	observacion := c.FormValue("observacion")
	if strings.TrimSpace(observacion) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "observacion es requerida: indica el motivo de carga del lote"})
	}

	fileHeader, err := c.FormFile("archivo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "El archivo .xlsx (campo 'archivo') es requerido", "error": err.Error()})
	}

	archivo, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "No se pudo abrir el archivo", "error": err.Error()})
	}
	defer archivo.Close()

	resultado, err := h.service.ImportarExcel(archivo, uint(sucursalFacturadorID), observacion)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error importando el Excel", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Importación procesada",
		"data":    resultado,
	})
}

func (h *FacturaPrevaloradaHandler) GetAll(c *fiber.Ctx) error {
	estado := c.Query("estado")
	loteID := c.Query("lote_id")

	facturas, err := h.service.ListarTodos(estado, loteID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo facturas prevaloradas", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Facturas prevaloradas obtenidas exitosamente", "data": facturas})
}

// GetLotes lista el registro de lotes de importación: con qué sucursal
// facturador y tipo se cargó cada uno, y el desglose de estados de envío.
func (h *FacturaPrevaloradaHandler) GetLotes(c *fiber.Ctx) error {
	lotes, err := h.service.ListarLotes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo lotes", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Lotes obtenidos exitosamente", "data": lotes})
}

// DescargarPlantilla entrega el .xlsx de ejemplo con las columnas esperadas
// por ImportarExcel.
func (h *FacturaPrevaloradaHandler) DescargarPlantilla(c *fiber.Ctx) error {
	contenido, err := h.service.GenerarPlantilla()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error generando la plantilla", "error": err.Error()})
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="plantilla_facturas_prevaloradas.xlsx"`)
	return c.Send(contenido)
}

func (h *FacturaPrevaloradaHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	factura, err := h.service.ObtenerPorID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Factura prevalorada no encontrada", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Factura prevalorada encontrada", "data": factura})
}

func (h *FacturaPrevaloradaHandler) RegisterRoutes(router fiber.Router) {
	facturas := router.Group("/facturas-prevaloradas")
	facturas.Post("/importar-excel", h.ImportarExcel)
	facturas.Get("/plantilla", h.DescargarPlantilla)
	facturas.Get("/lotes", h.GetLotes)
	facturas.Get("/", h.GetAll)
	facturas.Get("/:id", h.GetByID)
}
