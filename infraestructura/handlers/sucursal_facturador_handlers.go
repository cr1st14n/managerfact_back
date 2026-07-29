package handlers

import (
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"
	"managerfact/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SucursalFacturadorHandler struct {
	service *services.SucursalFacturadorService
}

func NewSucursalFacturadorHandler(s *services.SucursalFacturadorService) *SucursalFacturadorHandler {
	return &SucursalFacturadorHandler{service: s}
}

type sucursalFacturadorRequest struct {
	Nombre            string `json:"nombre"`
	CodigoSucursalSin int    `json:"codigo_sucursal_sin"`
	PuntoVentaEmisor  string `json:"punto_venta_emisor"`
	UrlLinkFacturador string `json:"url_link_facturador"`
	// TokenAcceso solo es requerido al crear. Al actualizar, vacío significa
	// "no cambiar el token existente".
	TokenAcceso     string `json:"token_acceso"`
	CodigoMonedaBob string `json:"codigo_moneda_bob"`
	CodigoCI        string `json:"codigo_ci"`
	CodigoNit       string `json:"codigo_nit"`
	Activo          *bool  `json:"activo,omitempty"`
}

// sucursalFacturadorResponse envuelve el modelo sin exponer nunca el token
// cifrado/plano, indicando solo si ya hay uno configurado.
type sucursalFacturadorResponse struct {
	models.SucursalFacturador
	TokenConfigurado bool `json:"token_configurado"`
}

func nuevaSucursalFacturadorResponse(s *models.SucursalFacturador) sucursalFacturadorResponse {
	return sucursalFacturadorResponse{
		SucursalFacturador: *s,
		TokenConfigurado:   s.TokenAcceso != "",
	}
}

func nuevasSucursalesFacturadorResponse(sucursales []models.SucursalFacturador) []sucursalFacturadorResponse {
	respuestas := make([]sucursalFacturadorResponse, 0, len(sucursales))
	for _, s := range sucursales {
		respuestas = append(respuestas, nuevaSucursalFacturadorResponse(&s))
	}
	return respuestas
}

func (h *SucursalFacturadorHandler) validarCreate(req *sucursalFacturadorRequest) []string {
	var errValidacion []string
	req.Nombre = utils.ValidarCampoRequerido(&errValidacion, req.Nombre, "El campo nombre es requerido")
	req.UrlLinkFacturador = utils.ValidarCampoRequerido(&errValidacion, req.UrlLinkFacturador, "El campo url_link_facturador es requerido")
	req.TokenAcceso = utils.ValidarCampoRequerido(&errValidacion, req.TokenAcceso, "El campo token_acceso es requerido")
	req.CodigoNit = utils.ValidarCampoRequerido(&errValidacion, req.CodigoNit, "El campo codigo_nit es requerido")
	if req.CodigoSucursalSin == 0 {
		errValidacion = append(errValidacion, "El campo codigo_sucursal_sin es requerido")
	}
	req.PuntoVentaEmisor = utils.ValidarCampoOpcional(&errValidacion, req.PuntoVentaEmisor)
	req.CodigoMonedaBob = utils.ValidarCampoOpcional(&errValidacion, req.CodigoMonedaBob)
	req.CodigoCI = utils.ValidarCampoOpcional(&errValidacion, req.CodigoCI)
	return errValidacion
}

func (h *SucursalFacturadorHandler) validarUpdate(req *sucursalFacturadorRequest) []string {
	var errValidacion []string
	req.Nombre = utils.ValidarCampoRequerido(&errValidacion, req.Nombre, "El campo nombre es requerido")
	req.UrlLinkFacturador = utils.ValidarCampoRequerido(&errValidacion, req.UrlLinkFacturador, "El campo url_link_facturador es requerido")
	req.CodigoNit = utils.ValidarCampoRequerido(&errValidacion, req.CodigoNit, "El campo codigo_nit es requerido")
	if req.CodigoSucursalSin == 0 {
		errValidacion = append(errValidacion, "El campo codigo_sucursal_sin es requerido")
	}
	req.PuntoVentaEmisor = utils.ValidarCampoOpcional(&errValidacion, req.PuntoVentaEmisor)
	req.TokenAcceso = utils.ValidarCampoOpcional(&errValidacion, req.TokenAcceso)
	req.CodigoMonedaBob = utils.ValidarCampoOpcional(&errValidacion, req.CodigoMonedaBob)
	req.CodigoCI = utils.ValidarCampoOpcional(&errValidacion, req.CodigoCI)
	return errValidacion
}

func (h *SucursalFacturadorHandler) Create(c *fiber.Ctx) error {
	var req sucursalFacturadorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}

	if errValidacion := h.validarCreate(&req); len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "errors": errValidacion})
	}

	sucursal, err := h.service.Crear(services.CrearSucursalFacturadorInput{
		Nombre:            req.Nombre,
		CodigoSucursalSin: req.CodigoSucursalSin,
		PuntoVentaEmisor:  req.PuntoVentaEmisor,
		UrlLinkFacturador: req.UrlLinkFacturador,
		TokenAcceso:       req.TokenAcceso,
		CodigoMonedaBob:   req.CodigoMonedaBob,
		CodigoCI:          req.CodigoCI,
		CodigoNit:         req.CodigoNit,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error creando sucursal facturador", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sucursal facturador creada exitosamente",
		"data":    nuevaSucursalFacturadorResponse(sucursal),
	})
}

func (h *SucursalFacturadorHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}

	var req sucursalFacturadorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}

	if errValidacion := h.validarUpdate(&req); len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "errors": errValidacion})
	}

	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	sucursal, err := h.service.Actualizar(services.ActualizarSucursalFacturadorInput{
		ID:                uint(id),
		Nombre:            req.Nombre,
		CodigoSucursalSin: req.CodigoSucursalSin,
		PuntoVentaEmisor:  req.PuntoVentaEmisor,
		UrlLinkFacturador: req.UrlLinkFacturador,
		TokenAcceso:       req.TokenAcceso,
		CodigoMonedaBob:   req.CodigoMonedaBob,
		CodigoCI:          req.CodigoCI,
		CodigoNit:         req.CodigoNit,
		Activo:            activo,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error actualizando sucursal facturador", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Sucursal facturador actualizada exitosamente", "data": nuevaSucursalFacturadorResponse(sucursal)})
}

func (h *SucursalFacturadorHandler) GetAll(c *fiber.Ctx) error {
	sucursales, err := h.service.ListarTodos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo sucursales facturador", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sucursales facturador obtenidas exitosamente", "data": nuevasSucursalesFacturadorResponse(sucursales)})
}

func (h *SucursalFacturadorHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	sucursal, err := h.service.ObtenerPorID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Sucursal facturador no encontrada", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sucursal facturador encontrada", "data": nuevaSucursalFacturadorResponse(sucursal)})
}

func (h *SucursalFacturadorHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	if err := h.service.Eliminar(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error eliminando sucursal facturador", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sucursal facturador eliminada exitosamente"})
}

func (h *SucursalFacturadorHandler) RegisterRoutes(router fiber.Router) {
	sucursales := router.Group("/sucursales-facturador")
	sucursales.Post("/", h.Create)
	sucursales.Get("/", h.GetAll)
	sucursales.Get("/:id", h.GetByID)
	sucursales.Put("/:id", h.Update)
	sucursales.Delete("/:id", h.Delete)
}
