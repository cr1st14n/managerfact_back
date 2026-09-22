package handlers

import (
	"managerfact/aplication/services"
	"managerfact/pkg/utils"
	"strconv"

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

type codigoProductoRequest struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}

func (h *CodigoProductoHandler) validar(req *codigoProductoRequest) []string {
	var errValidacion []string
	req.Codigo = utils.ValidarCampoRequerido(&errValidacion, req.Codigo, "El campo codigo es requerido")
	req.Descripcion = utils.ValidarCampoRequerido(&errValidacion, req.Descripcion, "El campo descripcion es requerido")
	return errValidacion
}

func (h *CodigoProductoHandler) Create(c *fiber.Ctx) error {
	var req codigoProductoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}
	if errValidacion := h.validar(&req); len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "errors": errValidacion})
	}

	data, err := h.codigoProductoService.Crear(req.Codigo, req.Descripcion)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error creando código de producto", "error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Código de producto creado exitosamente", "data": data})
}

func (h *CodigoProductoHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}

	var req codigoProductoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}
	if errValidacion := h.validar(&req); len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "errors": errValidacion})
	}

	data, err := h.codigoProductoService.Actualizar(uint(id), req.Codigo, req.Descripcion)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error actualizando código de producto", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Código de producto actualizado exitosamente", "data": data})
}

func (h *CodigoProductoHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	if err := h.codigoProductoService.Eliminar(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error eliminando código de producto", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Código de producto eliminado exitosamente"})
}

// RegisterRoutes registra las rutas bajo /codigoproducto. El listado queda
// abierto a cualquier autenticado (Reportes y Descarga Mensual lo usan para
// elegir productos, incluido el rol "consultas"); crear/editar/eliminar solo
// admin.
func (h *CodigoProductoHandler) RegisterRoutes(router fiber.Router, requireAdmin fiber.Handler) {
	codigos := router.Group("/codigoproducto")
	codigos.Get("/", h.GetAll)
	codigos.Post("/", requireAdmin, h.Create)
	codigos.Put("/:id", requireAdmin, h.Update)
	codigos.Delete("/:id", requireAdmin, h.Delete)
}
