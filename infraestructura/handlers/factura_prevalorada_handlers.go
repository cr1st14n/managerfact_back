package handlers

import (
	"errors"
	"managerfact/aplication/services"
	"managerfact/infraestructura/middleware"
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

func usuarioIDDesdeContexto(c *fiber.Ctx) (uint, bool) {
	usuarioID, ok := c.Locals(middleware.UsuarioIDLocal).(uint)
	return usuarioID, ok
}

func (h *FacturaPrevaloradaHandler) ImportarExcel(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDesdeContexto(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
	}

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

	resultado, err := h.service.ImportarExcel(usuarioID, archivo, uint(sucursalFacturadorID), observacion)
	if err != nil {
		if errors.Is(err, services.ErrSinPermisoSucursal) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error importando el Excel", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Importación procesada",
		"data":    resultado,
	})
}

func (h *FacturaPrevaloradaHandler) GetAll(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDesdeContexto(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
	}

	estado := c.Query("estado")
	loteID := c.Query("lote_id")

	facturas, err := h.service.ListarTodos(usuarioID, estado, loteID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo facturas prevaloradas", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Facturas prevaloradas obtenidas exitosamente", "data": facturas})
}

func (h *FacturaPrevaloradaHandler) GetLotes(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDesdeContexto(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
	}

	lotes, err := h.service.ListarLotes(usuarioID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo lotes", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Lotes obtenidos exitosamente", "data": lotes})
}

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
	usuarioID, ok := usuarioIDDesdeContexto(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	factura, err := h.service.ObtenerPorID(usuarioID, uint(id))
	if err != nil {
		if errors.Is(err, services.ErrSinPermisoSucursal) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Factura prevalorada no encontrada", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Factura prevalorada encontrada", "data": factura})
}

func (h *FacturaPrevaloradaHandler) Facturar(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDesdeContexto(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}

	if _, err := h.service.ObtenerPorID(usuarioID, uint(id)); err != nil {
		if errors.Is(err, services.ErrSinPermisoSucursal) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Factura prevalorada no encontrada", "error": err.Error()})
	}

	factura, err := h.service.Facturar(uint(id), "manual")
	if err != nil {
		if errors.Is(err, services.ErrFacturaYaAceptada) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": err.Error()})
		}
		if factura != nil {

			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"message": "Error enviando la factura al facturador", "error": err.Error(), "data": factura})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error facturando", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Factura enviada al facturador", "data": factura})
}

func (h *FacturaPrevaloradaHandler) RegisterRoutes(router fiber.Router, requireNoConsultas fiber.Handler) {
	facturas := router.Group("/facturas-prevaloradas", requireNoConsultas)
	facturas.Post("/importar-excel", h.ImportarExcel)
	facturas.Post("/:id/facturar", h.Facturar)
	facturas.Get("/plantilla", h.DescargarPlantilla)
	facturas.Get("/lotes", h.GetLotes)
	facturas.Get("/", h.GetAll)
	facturas.Get("/:id", h.GetByID)
}
